# Makefile for ServerScheduler

.PHONY: build run test clean frontend-build frontend-serve docker-build docker-run docker-push docker-build-backend docker-build-frontend docker-push-backend docker-push-frontend docker-push-all docker-compose-up docker-compose-down

# Build variables
BINARY_NAME=serverscheduler
DOCKER_IMAGE?=ghcr.io/rusik69/serverscheduler
DOCKER_TAG?=latest

# Backend commands
build:
	go build -o $(BINARY_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

# Frontend commands
frontend-build:
	cd frontend && npm install && npm run build

frontend-serve:
	cd frontend && npm run serve

# Docker commands for combined image (legacy)
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run:
	docker run -d \
		--name $(BINARY_NAME) \
		-p 8080:8080 \
		-v $(PWD)/data:/app/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Separate Docker builds
docker-build-backend:
	docker build -f Dockerfile.backend -t $(DOCKER_IMAGE)-backend:$(DOCKER_TAG) .

docker-build-frontend:
	docker build -f frontend/Dockerfile -t $(DOCKER_IMAGE)-frontend:$(DOCKER_TAG) ./frontend

docker-build-all: docker-build-backend docker-build-frontend

# Push separate images
docker-push-backend:
	docker build -f Dockerfile.backend -t $(DOCKER_IMAGE)-backend:$(DOCKER_TAG) .
	docker push $(DOCKER_IMAGE)-backend:$(DOCKER_TAG)

docker-push-frontend:
	docker build -f frontend/Dockerfile -t $(DOCKER_IMAGE)-frontend:$(DOCKER_TAG) ./frontend
	docker push $(DOCKER_IMAGE)-frontend:$(DOCKER_TAG)

docker-push-all: docker-push-backend docker-push-frontend

# Docker Compose commands
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-build:
	docker-compose build

docker-compose-logs:
	docker-compose logs -f

docker-compose-restart:
	docker-compose restart

# Development commands
dev-backend:
	go run ./cmd/server

dev-frontend:
	cd frontend && npm run serve

# Clean up Docker resources
docker-clean:
	docker-compose down -v
	docker system prune -f 