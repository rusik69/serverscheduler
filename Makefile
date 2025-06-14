# Makefile for ServerScheduler

.PHONY: build run test clean docker-build docker-run docker-push

# Build variables
BINARY_NAME=serverscheduler
DOCKER_IMAGE ?= serverscheduler
DOCKER_TAG ?= latest
DOCKER_PLATFORMS=linux/amd64,linux/arm64

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

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