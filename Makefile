.PHONY: build test run clean lint help up mysql-up mysql-down test-api

# Go related variables
BINARY_NAME=urlshortener
MAIN_PATH=./cmd/server
API_PORT=3000

# MySQL variables
MYSQL_CONTAINER=urlshortener-mysql
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=urlshortener
MYSQL_PORT=3306
MYSQL_DSN=root:$(MYSQL_ROOT_PASSWORD)@tcp(127.0.0.1:$(MYSQL_PORT))/$(MYSQL_DATABASE)?parseTime=true

# API test variables
TEST_URL=https://www.google.com
TEST_UPDATED_URL=https://www.github.com

# Check for Go installation
GOVERSION := $(shell go version 2>/dev/null)
GO_CHECK := $(shell which go)

# Check for Docker installation
DOCKER_CHECK := $(shell which docker)

# Default target
.DEFAULT_GOAL := help

help: ## Display this help message
	@echo "Usage:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

check-docker: ## Check if Docker is installed
	@if [ "$(DOCKER_CHECK)" = "" ]; then \
		echo "Error: Docker is not installed or not in PATH. Please install Docker first:"; \
		echo "Visit: https://docs.docker.com/get-docker/"; \
		exit 1; \
	fi

check-mysql: ## Check if MySQL container is running
	@if [ "$$(docker ps -q -f name=$(MYSQL_CONTAINER))" ]; then \
		echo "MySQL is running"; \
	else \
		echo "MySQL is not running. Starting MySQL..."; \
		make mysql-up; \
		echo "Waiting for MySQL to be ready..."; \
		sleep 10; \
	fi

mysql-up: check-docker ## Start MySQL container
	@if [ ! "$$(docker ps -q -f name=$(MYSQL_CONTAINER))" ]; then \
		if [ "$$(docker ps -aq -f status=exited -f name=$(MYSQL_CONTAINER))" ]; then \
			docker rm $(MYSQL_CONTAINER); \
		fi; \
		docker run --name $(MYSQL_CONTAINER) \
			-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
			-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
			-p $(MYSQL_PORT):3306 \
			-d mysql:8.0; \
	else \
		echo "MySQL is already running"; \
	fi

mysql-down: check-docker ## Stop MySQL container
	@if [ "$$(docker ps -q -f name=$(MYSQL_CONTAINER))" ]; then \
		docker stop $(MYSQL_CONTAINER); \
		docker rm $(MYSQL_CONTAINER); \
	fi

check-go: ## Check if Go is installed
	@if [ "$(GO_CHECK)" = "" ]; then \
		echo "Error: Go is not installed or not in PATH. Please install Go first:"; \
		echo "Visit: https://golang.org/doc/install"; \
		echo "For MacOS, you can use: brew install go"; \
		exit 1; \
	fi
	@echo "Found Go installation: $(GOVERSION)"

clean: ## Remove build artifacts
	go clean
	rm -f $(BINARY_NAME)

lint: check-go ## Run linters
	go vet ./...
	go fmt ./...

test: check-go ## Run tests
	go test -v ./...

build: check-go ## Build the binary
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run: check-go ## Run the application
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" go run $(MAIN_PATH)

dev: check-go ## Run the application with hot reload
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" go install github.com/cosmtrek/air@latest
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" air

tidy: check-go ## Tidy and vendor dependencies
	go mod tidy
	go mod vendor

# Database related commands
migrate-up: check-go ## Run database migrations up
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" go run $(MAIN_PATH) migrate up

migrate-down: check-go ## Run database migrations down
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" go run $(MAIN_PATH) migrate down

# Docker related commands (if needed)
docker-build: ## Build docker image
	docker build -t $(BINARY_NAME) .

docker-run: ## Run docker container
	docker run -p $(API_PORT):$(API_PORT) $(BINARY_NAME)

up: check-go check-mysql tidy migrate-up ## Set up and run the entire application
	@echo "Starting URL shortener application on port $(API_PORT)..."
	@echo "You can access the API at http://localhost:$(API_PORT)"
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" go run $(MAIN_PATH)

test-api: ## Test the API endpoints
	@echo "\nCreating a short URL for $(TEST_URL)..."
	@curl -X POST -H "Content-Type: application/json" \
		-d '{"url":"$(TEST_URL)"}' \
		http://localhost:$(API_PORT)/shorten
	
	@echo "\n\nWait a few seconds before testing the created URL..."
	@sleep 2
	
	@echo "\nGetting the last created short URL info (replace SHORTCODE with the code you received)..."
	@echo "curl http://localhost:$(API_PORT)/shorten/SHORTCODE"
	
	@echo "\nUpdating the URL (replace SHORTCODE with your code)..."
	@echo "curl -X PUT -H \"Content-Type: application/json\" \\"
	@echo "     -d '{\"url\":\"$(TEST_UPDATED_URL)\"}' \\"
	@echo "     http://localhost:$(API_PORT)/shorten/SHORTCODE"
	
	@echo "\nGetting statistics (replace SHORTCODE with your code)..."
	@echo "curl http://localhost:$(API_PORT)/shorten/SHORTCODE/stats"
	
	@echo "\nTesting redirect (replace SHORTCODE with your code)..."
	@echo "curl -i http://localhost:$(API_PORT)/SHORTCODE"
	
	@echo "\nDeleting the URL (replace SHORTCODE with your code)..."
	@echo "curl -X DELETE http://localhost:$(API_PORT)/shorten/SHORTCODE"
	
	@echo "\n\nReplace SHORTCODE in the above commands with the code you received from the first command." 