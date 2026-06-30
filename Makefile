GO ?= go
COMPOSE_FILE ?= infra/docker/docker-compose.yml

.PHONY: api worker test fmt db-up db-down db-reset migrate-up migrate-down setup dev seed-admin seed-demo-fixture seed-restaurants-data import-restaurants-outreach openapi swagger

OPENAPI_SPEC ?= docs/openapi/openapi.yaml
OPENAPI_DIR ?= docs/openapi
SWAGGER_PORT ?= 8081

api:
	$(GO) run ./backend/cmd/api

worker:
	$(GO) run ./backend/cmd/worker

seed-admin:
	$(GO) run ./backend/cmd/seed-admin

seed-demo-fixture:
	$(GO) run ./backend/cmd/seed-demo-fixture

seed-restaurants-data:
	$(GO) run ./backend/cmd/seed-restaurants-data

# Run migrations + import restaurants_data.json using DATABASE_URL from outreach/.env
import-restaurants-outreach:
	@set -a && . ./automation/outreach/.env && set +a && \
	$(GO) run ./backend/cmd/migrate up && \
	$(GO) run ./backend/cmd/seed-restaurants-data

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

openapi:
	@command -v npx >/dev/null 2>&1 || { echo "npx required: install Node.js or validate manually at editor.swagger.io"; exit 1; }
	npx --yes @redocly/cli@1 lint $(OPENAPI_SPEC) --skip-rule=no-server-example.com

swagger:
	@command -v docker >/dev/null 2>&1 || { echo "docker required to run Swagger UI"; exit 1; }
	@echo "Swagger UI → http://localhost:$(SWAGGER_PORT)"
	docker run --rm -p $(SWAGGER_PORT):8080 \
		-e SWAGGER_JSON=/openapi/openapi.yaml \
		-v "$(CURDIR)/$(OPENAPI_DIR):/openapi" \
		swaggerapi/swagger-ui
