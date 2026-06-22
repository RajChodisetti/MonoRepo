GO ?= go
COMPOSE_FILE ?= infra/docker/docker-compose.yml

.PHONY: api worker test fmt db-up db-down db-reset migrate-up migrate-down setup dev

api:
	$(GO) run ./backend/cmd/api

worker:
	$(GO) run ./backend/cmd/worker

test:
	$(GO) test ./backend/...

fmt:
	gofmt -w $$(find backend -name '*.go')

db-up:
	docker compose -f $(COMPOSE_FILE) up -d --wait

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-reset:
	docker compose -f $(COMPOSE_FILE) down -v

migrate-up:
	$(GO) run ./backend/cmd/migrate up

migrate-down:
	$(GO) run ./backend/cmd/migrate down

setup: db-up migrate-up

dev: setup api
