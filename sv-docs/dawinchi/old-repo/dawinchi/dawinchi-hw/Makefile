COMMIT ?= $(shell git rev-parse HEAD|head -c 7)

DOCKER_TAG=dawinchi-hw:latest
PROD_DOCKER_TAG=gcr.io/createviewprototypeproject/dawinchi-hw:latest
DAWINCHI_HW_PROD_DOCKER_TAG=eu.gcr.io/createviewprototypeproject/dawinchi-hw:$(COMMIT)

.PHONY: all
all: test

.PHONY: test
test:
	go test --count=1 -v ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: install-deps
install-deps: ## golangci-lint
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.0
	go mod tidy

swag: # Generates the Swagger API docs
	swag --version
	swag init -g rpc/routes.go -o docs/swagger -d ./internal

.PHONY: build-docker
build-docker:
	docker build -t $(DOCKER_TAG) .

.PHONY: build-docker-cloud-push
build-docker-cloud-push:
	docker buildx build -t $(PROD_DOCKER_TAG) -f Dockerfile.cloud --platform=linux/amd64 --push .

.PHONY: up
up: ## Starts docker containers for dependent services
	@docker compose up -d --build --remove-orphans

.PHONY: stop
stop: ## Stops docker containers for dependent services
	@docker compose stop

.PHONY: down
down: ## Removes docker containers for dependent services
	@docker compose down --remove-orphans --volumes

build-docker-cloud-push:
	@go mod vendor
	@docker buildx build -t $(DAWINCHI_HW_PROD_DOCKER_TAG) -f Dockerfile --platform=linux/amd64 --push .
