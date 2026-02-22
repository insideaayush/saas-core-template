SHELL := /bin/sh

.PHONY: infra-up infra-down dev-api dev-ui dev-worker test ci smoke-local

infra-up:
	docker compose up -d postgres redis otel-collector

infra-down:
	docker compose down

dev-api:
	cd backend && go run ./cmd/api

dev-worker:
	cd backend && go run ./cmd/worker

dev-ui:
	cd frontend && npm run dev

smoke-local:
	bash scripts/smoke-local.sh

test:
	cd backend && go test ./...
	cd frontend && npm run lint && npm run typecheck

ci: test
	cd backend && go vet ./... && go build ./cmd/api
	cd frontend && npm run build
