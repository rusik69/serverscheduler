# Makefile for ServerScheduler

BINARY=bin/server
DOCKER_IMAGE=serverscheduler
DOCKER_TAG=latest
DOCKER_DATA_DIR=$(shell pwd)/data

.PHONY: all build test clean docker-build docker-run docker-clean

all: build

build:
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	@mkdir -p $(DOCKER_DATA_DIR)
	docker run -p 8080:8080 -v $(DOCKER_DATA_DIR):/app/data $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean:
	docker stop $$(docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG)) 2>/dev/null || true
	docker rm $$(docker ps -a -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG)) 2>/dev/null || true 