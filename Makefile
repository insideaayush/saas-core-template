SHELL := /bin/sh

.PHONY: infra-up infra-down dev-api dev-ui test ci

infra-up:
	docker compose up -d postgres redis

infra-down:
	docker compose down

dev-api:
	cd backend && go run ./cmd/api

dev-ui:
	cd frontend && npm run dev

test:
	cd backend && go test ./...
	cd frontend && npm run lint && npm run typecheck

ci: test
	cd backend && go vet ./... && go build ./cmd/api
	cd frontend && npm run build
