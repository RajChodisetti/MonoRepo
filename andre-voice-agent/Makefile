.PHONY: start up down logs dev ui dial venv tunnel

PORT ?= 8001
PYTHON ?= /opt/homebrew/opt/python@3.12/bin/python3.12

venv:
	$(PYTHON) -m venv .venv
	.venv/bin/pip install -U pip
	.venv/bin/pip install -r requirements.txt

start up:
	mkdir -p data
	docker compose up --build

dev:
	mkdir -p data
	.venv/bin/uvicorn bot:app --host 0.0.0.0 --port $(PORT)

dev-reload:
	mkdir -p data
	.venv/bin/uvicorn bot:app --host 0.0.0.0 --port $(PORT) --reload --reload-dir . --reload-exclude .venv --reload-exclude __pycache__

# Quick tunnel for Twilio Media Streams. Copy printed URL into PUBLIC_BASE_URL then restart agent.
tunnel:
	cloudflared tunnel --url http://127.0.0.1:$(PORT) --no-autoupdate

ui:
	.venv/bin/python ui.py

dial:
	@echo "Usage: make dial TO=+9198… LANG=hi"
	.venv/bin/python dial.py $(TO) --language $(or $(LANG),auto)

down:
	docker compose down

logs:
	docker compose logs -f agent
