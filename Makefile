# Makefile for ServerScheduler

BINARY=bin/server
DOCKER_IMAGE ?= serverscheduler
DOCKER_TAG ?= latest
DOCKER_DATA_DIR=$(shell pwd)/data
DOCKER_PLATFORMS=linux/amd64,linux/arm64

.PHONY: all build test clean docker-build docker-run docker-clean docker-push

all: build

build:
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin

docker-build:
	docker buildx build \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--cache-from type=registry,ref=$(DOCKER_IMAGE):buildcache \
		--cache-to type=registry,ref=$(DOCKER_IMAGE):buildcache,mode=max \
		.

docker-push:
	docker buildx build \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--push \
		--cache-from type=registry,ref=$(DOCKER_IMAGE):buildcache \
		--cache-to type=registry,ref=$(DOCKER_IMAGE):buildcache,mode=max \
		.

docker-run:
	@mkdir -p $(DOCKER_DATA_DIR)
	docker run -p 8080:8080 -v $(DOCKER_DATA_DIR):/app/data $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean:
	docker stop $$(docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG)) 2>/dev/null || true
	docker rm $$(docker ps -a -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG)) 2>/dev/null || true 