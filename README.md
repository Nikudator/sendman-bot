# suggest-bot
It is a pet-project telegram-bot for sending messages to subscribers of bot (golang, postgresql, rabbitmq)

1. Put your bot-token in config.yml
Sample:
token: 0000000000:QwERTyuiOPL897656LKJHGFfds

2. Put your data in ./docker/.env
Sample:
COMPOSE_PROJECT_NAME=sendman-bot
PROJECT_IP_MASK=172.25.3
PG_DB_NAME=sendman
PG_DB_USER=botdbuser
PG_DB_PASS=botpass
PG_DB_DIR=/var/lib/postgresql/data
RM_USER=rabbituser
RM_PASS=rabbitpass
