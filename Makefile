.PHONY: up down logs api-logs worker-logs web-logs db rebuild fmt test

up:
	docker compose up -d --build
	@echo ""
	@echo "CRM:  http://localhost:5173 (admin@demo.kz / demo1234)"
	@echo "API:  http://localhost:8080"

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

api-logs:
	docker compose logs -f --tail=100 api

worker-logs:
	docker compose logs -f --tail=100 worker

web-logs:
	docker compose logs -f --tail=100 web

db:
	docker compose exec postgres psql -U dentdesk

rebuild:
	docker compose build --no-cache

# Локальная разработка Go без докера (нужен Go 1.22+)
run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

fmt:
	gofmt -w .

# Dev фронта
web-dev:
	cd web && npm install && npm run dev
