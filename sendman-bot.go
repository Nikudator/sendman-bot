package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v2"
)

var pool *pgxpool.Pool
var rconn *amqp.Connection
var bot *tgbotapi.BotAPI

type Mess struct { //структура для отправки сообщений в очередь
	ID   int64  `json:"id"`
	Text string `json:"name"`
}

func main() {
	//Читаем конфиг
	const configPath = "config.yml"
	type Cfg struct {
		TELEGRAM_BOT_API_TOKEN  string `yaml:"token"`
		POSTGRES_HOST           string `yaml:"postgres_host"`
		POSTGRES_PORT           int    `yaml:"postgres_port"`
		POSTGRES_DB             string `yaml:"postgres_db"`
		POSTGRES_USER           string `yaml:"postgres_user"`
		POSTGRES_PASS           string `yaml:"postgres_pass"`
		POSTGRES_SSL            string `yaml:"postgres_ssl"`
		POSTGRES_POOL_MAX_CONNS int    `yaml:"postgres_pool_max_conns"`
		ADMIN_ID                int    `yaml:"admin_id"`
		RABBIT_HOST             string `yaml:"rabbit_host"`
		RABBIT_PORT             int    `yaml:"rabbit_port"`
		RABBIT_USER             string `yaml:"rabbit_user"`
		RABBIT_PASS             string `yaml:"rabbit_pass"`
	}
	var AppConfig *Cfg
	f, err := os.Open(configPath)
	failOnError(err, "Can't open config.\n")
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&AppConfig)
	failOnError(err, "Can't decode config.\n")

	bot_token := AppConfig.TELEGRAM_BOT_API_TOKEN
	postgres_host := AppConfig.POSTGRES_HOST
	postgres_port := AppConfig.POSTGRES_PORT
	postgres_db := AppConfig.POSTGRES_DB
	postgres_user := AppConfig.POSTGRES_USER
	postgres_pass := AppConfig.POSTGRES_PASS
	postgres_ssl := AppConfig.POSTGRES_SSL
	postgres_pool_max_conns := AppConfig.POSTGRES_POOL_MAX_CONNS
	admin_id := AppConfig.ADMIN_ID
	rabbit_host := AppConfig.RABBIT_HOST
	rabbit_port := AppConfig.RABBIT_PORT
	rabbit_user := AppConfig.RABBIT_USER
	rabbit_pass := AppConfig.RABBIT_PASS
	//Инициализация БД
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		postgres_user, postgres_pass, postgres_host, postgres_port, postgres_db, postgres_ssl, postgres_pool_max_conns)

	pool, err = pgxpool.New(context.Background(), dbURL)
	failOnError(err, "Unable to connection to database: %v.\n")
	defer pool.Close()
	log.Print("Connected to database!\n")

	//Инициализация RabbitMQ
	rconn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbit_user, rabbit_pass, rabbit_host, rabbit_port))
	failOnError(err, "Failed to connect to RabbitMQ\n")
	defer rconn.Close()

	//Создаём бота
	bot, err = tgbotapi.NewBotAPI(bot_token)
	failOnError(err, "Can't registration bot token.\n")

	bot.Debug = true

	log.Printf("Bot is connected %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s\n", update.Message.From.UserName, string(update.Message.Text))
			//createUser не вынесена в "start", потому что в случае краша базы, пользователи повторно будут добавляться в новую.
			createUser(update.Message.Chat.ID, update.Message.From.UserName)
			var msg tgbotapi.MessageConfig
			switch update.Message.Command() {
			case "start":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Приветствую! Я бот для информирования мужчин о работе по борьбе за мужские права.\nТеперь иногда вы будете получать от меня важные информационные сообщения.")
			case "help":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Я поддерживаю следующие комманды:\n/start - Старт бота\n/help - Показать помощь\n/petition - Получить список петиций.\nЕсли хотите написать администратору сообщение, просто напишите его и, если нужно, прикрепите фото или видео.")
			case "petition":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Список петиций где нужно голосовать ЗА:\n \n \nСписок петиций где нужно голосовать ПРОТИВ: \n \n \n")
			default:
				if getUserRole(update.Message.Chat.ID) > 0 { //Если сообщение пришло от админа, то генерируем сообщения брокеру.
					Message2queue := Mess{ID: 1, Text: update.Message.Text}
					err = sendMessageToQueue(Message2queue)
					failOnError(err, "Cann't send message to queue\n")
				} else { //Если сообщение пришло от не админа, пересылаем его админу.
					var msg_adm tgbotapi.ForwardConfig
					msg_adm = tgbotapi.NewForward(int64(admin_id), update.Message.From.ID, update.Message.MessageID)
					bot.Send(msg_adm)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ваше сообщение отправлено администратору.")
				}
			}
			bot.Send(msg)
		}
		//Если входящих нет, начинаем рассылку из очереди.
		sendMessageToUser(5)
		failOnError(err, "Can't send messages to users.\n")
	}
}

func failOnError(err error, msg string) { //Делаем более читаемую и компактную обработку ошибок.
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func createUser(tid int64, uname string) error {
	queryCount := "SELECT COUNT(*) FROM botusers WHERE tid = $1"
	var count int
	err := pool.QueryRow(context.Background(), queryCount, tid).Scan(&count)
	failOnError(err, "Can't check user for adding user.\n")
	if count < 1 {
		queryCreate := "INSERT INTO botusers (tid, uname) VALUES ($1, $2) RETURNING id"
		var id int
		err := pool.QueryRow(context.Background(), queryCreate, tid, uname).Scan(&id)
		failOnError(err, "Can't create user.\n")
		log.Printf("Created user with ID: %d, TID: %d, NAME: %s.\n", id, tid, uname)
	}
	return err
}

func getUserRole(tid int64) int { //проверяем, является ли пользователь админом.
	queryCheck := "SELECT uadmin FROM botusers WHERE tid = $1"
	var uadmin bool
	err := pool.QueryRow(context.Background(), queryCheck, tid).Scan(&uadmin)
	failOnError(err, "Can't get user role \n")
	if uadmin {
		return 1
	}
	return 0
}

func sendMessageToQueue(body Mess) error {
	rch, err := rconn.Channel()
	failOnError(err, "Failed to open a channel\n")
	defer rch.Close()
	q, err := rch.QueueDeclare(
		"sender", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue\n")

	queryTid := "SELECT tid FROM botusers"
	rows, err := pool.Query(context.Background(), queryTid)
	failOnError(err, "Can't check user for adding user.\n")
	defer rows.Close()
	for rows.Next() {
		var tid int64
		err := rows.Scan(&tid)
		failOnError(err, "Error read row for tid.\n")
		body.ID = tid
		boddy, err := json.Marshal(body)
		failOnError(err, "Cannot convert message to JSON\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = rch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(boddy),
			})
		failOnError(err, "Failed to publish a message to queue\n")
	}
	return err
}

func sendMessageToUser(pause int) error {
	rch, err := rconn.Channel()
	failOnError(err, "Failed to open a channel\n")
	defer rch.Close()
	q, err := rch.QueueDeclare(
		"sender", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue\n")

	err = rch.ExchangeDeclare(
		"message", // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = rch.QueueBind(
		q.Name,    // queue name
		"",        // routing key
		"message", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := rch.Consume(
		q.Name,        // queue
		"sendman-bot", // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")
	var Message2queue Mess
	for d := range msgs {
		err := json.Unmarshal(d.Body, &Message2queue)
		failOnError(err, "Failed to convert message from JSON")
		msg := tgbotapi.NewMessage(Message2queue.ID, Message2queue.Text)
		bot.Send(msg)
		log.Printf(" [x] %s", d.Body)
		time.Sleep(time.Duration(1000/pause) * time.Microsecond)
	}
	return err
}
