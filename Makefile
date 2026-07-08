GO ?= go
COMPOSE_FILE ?= infra/docker/docker-compose.yml
COMPOSE_DIR ?= infra/docker
RESTAURANT_SERVICES_CATALOG_DIR ?= apps/restaurant-services-catalog

.PHONY: api worker test fmt db-up db-down db-reset migrate-up migrate-down setup dev seed-admin seed-demo-fixture seed-restaurants-data import-outreach ocr-all sanitize-import verify-leads-ocr import-restaurants-outreach ingest-daily openapi swagger up down logs start stop-all voice-up voice-down voice-logs restaurant-services-catalog-dev restaurant-services-catalog-build

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

# Run OCR on all restaurant JSON + import to local Docker DB
ocr-all: db-up migrate-up
	cd automation/outreach && \
	(test -d .venv && . .venv/bin/activate; \
	python menu_image_ocr.py --sanitize-data ../../data/restaurants_data.json && \
	python menu_image_ocr.py --batch-data ../../data/restaurants_data.json && \
	python import_to_db.py --restaurants-only)

# Quick fix: strip menu boards from dish cards + re-import (no API calls)
sanitize-import: db-up migrate-up
	cd automation/outreach && \
	(test -d .venv && . .venv/bin/activate; \
	python menu_image_ocr.py --sanitize-data ../../data/restaurants_data.json && \
	python import_to_db.py --restaurants-only)

# Import restaurant JSON to local Docker DB
import-outreach: db-up migrate-up
	cd automation/outreach && \
	(test -d .venv && . .venv/bin/activate; python import_to_db.py --restaurants-only)

# Verify unverified DB leads via menu OCR (nightly cron or manual)
verify-leads-ocr: db-up migrate-up
	cd automation/outreach && \
	(test -d .venv && . .venv/bin/activate; \
	LEAD_OCR_VERIFICATION_ENABLED=true python verify_leads_from_db.py --force)

# Daily lead ingestion — Apollo fetch + Places scrape + DB import (500 req budget)
ingest-daily: db-up migrate-up
	cd automation/outreach && \
	(test -d .venv && . .venv/bin/activate; \
	LEAD_INGESTION_ENABLED=true python daily_ingestion.py)

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

# VM / Docker: postgres + migrate + api + worker containers
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
