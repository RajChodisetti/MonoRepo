GO ?= go

.PHONY: api worker test fmt migrate-up migrate-down

api:
	$(GO) run ./backend/cmd/api

worker:
	$(GO) run ./backend/cmd/worker

test:
	$(GO) test ./backend/...

fmt:
	gofmt -w $$(find backend -name '*.go')

migrate-up:
	$(GO) run ./backend/cmd/migrate up

migrate-down:
	$(GO) run ./backend/cmd/migrate down
