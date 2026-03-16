.PHONY: build dev test docker-build podman-build podman-run podman-stop docker-compose-up docker-compose-down ensure-env

build:
	CGO_ENABLED=1 go build -o server ./cmd/server

dev: build
	./server

test:
	go test ./...

# Podman (for local testing)
podman-build:
	podman build -t serverscheduler:latest .

ensure-env:
	@test -f .env || (cp .env.example .env && echo "Created .env from .env.example")

podman-run: ensure-env podman-build
	podman rm -f serverscheduler 2>/dev/null || true
	podman run -d --name serverscheduler -p 8080:8080 \
		--env-file .env \
		-e DB_PATH=/app/data/serverscheduler.db \
		-v $(PWD)/data:/app/data \
		serverscheduler:latest

podman-stop:
	podman stop serverscheduler 2>/dev/null || true
	podman rm serverscheduler 2>/dev/null || true

# Docker (for production)
docker-build:
	docker build -t serverscheduler:latest .

docker-compose-up: ensure-env
	docker compose up -d

docker-compose-down:
	docker compose down
