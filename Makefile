.PHONY: help
help:
	@echo 'Usage: make [TARGET] [EXTRA_ARGUMENTS]'
	@echo ''
	@echo 'Service targets:'
	@echo '  start        start all services in the background or just the one specified by service= argument'
	@echo '  stop         stop all services or just the one specified by service= argument'
	@echo '  restart      restart all services or just the one specified by service= argument'
	@echo '  down         down all services and remove their data/images/etc.'
	@echo '  ps           show status of all services or just the one specified by service= argument'
	@echo '  logs         show logs of all services or just the one specified by service= argument'
	@echo ''
	@echo 'Tool targets:'
	@echo '  connect-bot  connect to the local bot interface to send commands'
	@echo '  redis-cli    run redis-cli within the redis container'
	@echo ''

.PHONY: start
start:
	@docker compose up -d $(service)

.PHONY: stop
stop:
	@docker compose stop $(service)

.PHONY: restart
restart:
	@docker compose stop $(service)
	@docker compose up -d $(service)

.PHONY: down
down:
	@docker compose down --rmi local --volumes

.PHONY: ps
ps:
	@docker compose ps

.PHONY: logs
logs:
	@docker compose logs --tail=100 -f $(service)

.PHONY: connect-bot
connect-bot:
	@nc localhost 5000

.PHONY: redis-cli
redis-cli:
	@docker compose exec redis redis-cli

.PHONY: test
test:
	@for dir in api bot controller; do \
     pushd $${dir} >/dev/null;       \
     go test ./...;                  \
     popd >/dev/null;                \
  done
