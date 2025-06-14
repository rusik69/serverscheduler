# Makefile for ServerScheduler

.PHONY: all build-backend build-frontend run test clean docker-build docker-run frontend-serve

# Build variables
BINARY_NAME=serverscheduler
DOCKER_IMAGE ?= serverscheduler
DOCKER_TAG ?= latest
DOCKER_PLATFORMS=linux/amd64,linux/arm64

all: build-backend build-frontend

# Build the backend
build-backend:
	CGO_ENABLED=1 go build -o $(BINARY_NAME) ./cmd/server

# Build the frontend
build-frontend:
	cd frontend && npm install && npm run build

# Run the application
run: build-backend
	./$(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -rf bin/
	rm -rf frontend/dist/

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Push Docker image
docker-push:
	docker buildx build \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--cache-from type=registry,ref=$(DOCKER_IMAGE):latest \
		--cache-to type=registry,ref=$(DOCKER_IMAGE):latest \
		--push \
		.

# Run Docker container
docker-run:
	docker run -p 8080:8080 -v $(PWD)/data:/app/data $(DOCKER_IMAGE):$(DOCKER_TAG)

# Serve frontend in development mode
frontend-serve:
	cd frontend && npm install && npm run serve 