.PHONY: api worker test fmt db-up db-down db-reset migrate-up migrate-down setup dev seed-admin seed-demo-fixture seed-restaurants-data import-outreach import-restaurants-outreach ingest-daily openapi swagger up down logs start stop-all voice-up voice-down voice-logs andre-voice-dev andre-voice-install tuvi-website-dev tuvi-website-build restaurant-services-catalog-dev restaurant-services-catalog-build agent-context agent-context-check

GO ?= go
COMPOSE_FILE ?= infra/docker/docker-compose.yml
COMPOSE_DIR ?= infra/docker
RESTAURANT_SERVICES_CATALOG_DIR ?= apps/restaurant-services-catalog
ANDRE_VOICE_DIR ?= andre-voice-agent
TUVI_WEBSITE_DIR ?= web

OPENAPI_SPEC ?= docs/openapi/openapi.yaml
OPENAPI_DIR ?= docs/openapi
SWAGGER_PORT ?= 8081
AGENT_PATHS ?=

# AGENT_PATHS is for whitespace-free paths; invoke scripts/agent-context.sh
# directly when a path contains spaces so shell argument boundaries are kept.
agent-context:
	@./scripts/agent-context.sh $(AGENT_PATHS)

agent-context-check:
	@./scripts/check-agent-context.sh

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

# Import restaurant JSON to local Docker DB
import-outreach: db-up migrate-up
	cd automation/outreach && \
	(test -d .venv && . .venv/bin/activate; python import_to_db.py --restaurants-only)

# Retired: durable city ingestion must enter through POST /api/v1/scrape-jobs.
ingest-daily:
	@echo "ingest-daily is retired; use the private city-scrape API and scrape-worker." >&2
	@exit 1

# Legacy: Go seed (uses backend/.env DATABASE_URL)
import-restaurants-outreach: db-up migrate-up
	$(GO) run ./backend/cmd/seed-restaurants-data

test:
	$(GO) test ./backend/...

fmt:
	gofmt -w $$(find backend -name '*.go')

db-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres --wait

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-reset:
	docker compose -f $(COMPOSE_FILE) down -v

migrate-up:
	$(GO) run ./backend/cmd/migrate up

migrate-down:
	$(GO) run ./backend/cmd/migrate down

setup: db-up migrate-up

# API only (legacy quick start — no worker)
dev: setup api

# Local: postgres + migrate + api + worker (one terminal, Ctrl+C to stop app processes)
start:
	@chmod +x scripts/start-all.sh scripts/stop-all.sh scripts/up-stack.sh
	@./scripts/start-all.sh

stop-all:
	@chmod +x scripts/stop-all.sh
	@./scripts/stop-all.sh

# VM / Docker stack: postgres + redis + migrate + api + worker + scrape-worker
up:
	@chmod +x scripts/up-stack.sh
	@./scripts/up-stack.sh

down:
	docker compose -f $(COMPOSE_FILE) --profile stack down

logs:
	docker compose -f $(COMPOSE_FILE) --profile stack logs -f

# Voice sales agent (Docker profile "voice")
# Requires voice-sales-agent/.env (copy from .env.example)
voice-up:
	mkdir -p voice-sales-agent/data
	docker compose -f $(COMPOSE_FILE) --profile voice up -d --build voice-sales-agent voice-sales-redis

voice-down:
	docker compose -f $(COMPOSE_FILE) --profile voice stop voice-sales-agent voice-sales-redis
	docker compose -f $(COMPOSE_FILE) --profile voice rm -f voice-sales-agent voice-sales-redis

voice-logs:
	docker compose -f $(COMPOSE_FILE) --profile voice logs -f voice-sales-agent

# Real-estate voice agent (Ananya) — local uvicorn on :8001
andre-voice-install:
	cd $(ANDRE_VOICE_DIR) && \
	(test -d .venv || python3 -m venv .venv) && \
	.venv/bin/pip install -r requirements.txt

andre-voice-dev:
	@test -f $(ANDRE_VOICE_DIR)/.env || (echo "Missing $(ANDRE_VOICE_DIR)/.env — copy from .env.example"; exit 1)
	@test -x $(ANDRE_VOICE_DIR)/.venv/bin/uvicorn || (echo "Run: make andre-voice-install"; exit 1)
	cd $(ANDRE_VOICE_DIR) && .venv/bin/uvicorn bot:app --host 0.0.0.0 --port 8001

tuvi-website-dev:
	npm --prefix $(TUVI_WEBSITE_DIR) run dev -- -p 3001

tuvi-website-build:
	npm --prefix $(TUVI_WEBSITE_DIR) run build

restaurant-services-catalog-dev:
	npm --prefix $(RESTAURANT_SERVICES_CATALOG_DIR) run dev

restaurant-services-catalog-build:
	npm --prefix $(RESTAURANT_SERVICES_CATALOG_DIR) run build

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
