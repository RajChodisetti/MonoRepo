.PHONY: start up down logs dev dev-detached

# Use the monorepo Docker service (infra/docker profile "voice").
# Pipecat/numba require Python >=3.10,<3.14 — Docker is the supported local run path.
COMPOSE_FILE ?= ../infra/docker/docker-compose.yml

start up:
	mkdir -p data
	docker compose -f $(COMPOSE_FILE) --profile voice up --build voice-sales-agent voice-sales-redis

dev:
	mkdir -p data
	docker compose -f $(COMPOSE_FILE) --profile voice up --build voice-sales-agent voice-sales-redis

dev-detached:
	mkdir -p data
	docker compose -f $(COMPOSE_FILE) --profile voice up -d --build voice-sales-agent voice-sales-redis

down:
	docker compose -f $(COMPOSE_FILE) --profile voice stop voice-sales-agent voice-sales-redis
	docker compose -f $(COMPOSE_FILE) --profile voice rm -f voice-sales-agent voice-sales-redis

logs:
	docker compose -f $(COMPOSE_FILE) --profile voice logs -f voice-sales-agent
