.PHONY: help
help:
	@echo 'Usage: make [TARGET] [EXTRA_ARGUMENTS]'
	@echo ''
	@echo 'Targets:'
	@echo '  help     	 show this help text'
	@echo '  up       	 start all services in the foreground or just the one specified by service= argument'
	@echo '  start    	 start all services in the background or just the one specified by service= argument'
	@echo '  restart  	 restart all services or just the one specified by service= argument'
	@echo '  stop     	 stop all services or just the one specified by service= argument'
	@echo '  down     	 down all services and remove their data/images/etc.'
	@echo '  status   	 show status of all services or just the one specified by service= argument'
	@echo '  ps       	 show status of all services or just the one specified by service= argument'
	@echo '  logs     	 show logs of all services or just the one specified by service= argument'
	@echo '  redis-cli   connect to the redis datastore'
	@echo '  connect-bot connect to the local bot interface to send commands'
	@echo ''

.PHONY: up
up:
	@docker-compose up $(service)

.PHONY: start
start:
	@docker-compose up -d $(service)

.PHONY: stop
stop:
	@docker-compose stop $(service)

.PHONY: down
down:
	@docker-compose down --rmi local --volumes

.PHONY: restart
restart:
	@docker-compose stop $(service)
	@docker-compose up -d $(service)

.PHONY: status
status:
	@docker-compose ps

.PHONY: ps
ps:
	@docker-compose ps

.PHONY: logs
logs:
	@docker-compose logs --tail=100 -f $(service)

.PHONY: redis-cli
redis-cli:
	@docker-compose exec redis redis-cli

.PHONY: connect-bot
connect-bot:
	@nc localhost 5000

%.tar: %/Dockerfile
	@docker build --rm -t "twitch-plays-crosswords-$*" "$*"
	@docker save -o "$*.tar" "twitch-plays-crosswords-$*"
	@docker image rm "twitch-plays-crosswords-$*"

.PHONY: deploy
deploy: api.tar bot.tar converter.tar ui.tar
	@scp $^ homelab:~/
	@ssh homelab 'for name in $^; do docker load -i $${name}; done && rm $^'
	@ssh homelab 'sudo systemctl restart docker-compose@twitch-plays-crosswords && docker image prune -f'
	@rm $^