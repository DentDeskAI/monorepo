docker:
	docker compose up --build -d

docker-down:
	docker compose down

.PHONY: docker, docker-down