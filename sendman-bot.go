package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v2"
)

func failOnError(err error, msg string) { //Делаем более читаемую и компактную обработку ошибок.
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
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

	//Инициализация БД
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		postgres_user, postgres_pass, postgres_host, postgres_port, postgres_db, postgres_ssl, postgres_pool_max_conns)

	pool, err := pgxpool.New(context.Background(), dbURL)
	failOnError(err, "Unable to connection to database: %v.\n")

	defer pool.Close()
	log.Print("Connected to database!\n")

	//Создаём бота
	bot, err := tgbotapi.NewBotAPI(bot_token)
	failOnError(err, "Can't registration bot token.\n")

	bot.Debug = true

	log.Printf("Бот подключился %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			var msg tgbotapi.MessageConfig

			switch update.Message.Command() {
			case "start":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Приветствую! Я бот для информирования мужчин о работе по борьбе за мужские права.\nТеперь иногда вы будете получать от меня важные информационные сообщения.")
			case "help":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Я поддерживаю следующие комманды:\n/start - Старт бота\n/help - Показать помощь\n/petition - Получить список петиций, в которых необходимо ваше участие\nЕсли хотите написать администратору сообщение, просто напишите его и, если нужно, прикрепите фото или видео.")
			case "petition":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Список петиций где нужно голосовать ЗА:\n \n \nСписок петиций где нужно голосовать ПРОТИВ: \n \n \n")
			default:
				var msg_adm tgbotapi.ForwardConfig
				msg_adm = tgbotapi.NewForward(int64(admin_id), update.Message.From.ID, update.Message.MessageID)
				bot.Send(msg_adm)
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ваше сообщение отправлено администратору.")
			}

			bot.Send(msg)
		}
	}
}
