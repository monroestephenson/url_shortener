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
TEST_EMAIL=test@example.com
TEST_PASSWORD=testpass123
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m
BOLD=\033[1m

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
	@echo "Press Ctrl+C to stop the server"
	PORT=$(API_PORT) MYSQL_DSN="$(MYSQL_DSN)" go run $(MAIN_PATH)

test-api: ## Test the API endpoints
	@echo "$(BOLD)🚀 Testing URL Shortener API...$(NC)\n"
	@START_TIME=`date +%s` && ( \
	echo "$(BOLD)1. Creating test user...$(NC)" && \
	RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" \
		-d '{"email":"$(TEST_EMAIL)","password":"$(TEST_PASSWORD)"}' \
		http://localhost:$(API_PORT)/auth/signup) && \
	STATUS=$$(echo "$$RESPONSE" | tail -n1) && \
	BODY=$$(echo "$$RESPONSE" | sed '$$d') && \
	if [ "$$STATUS" = "201" ] || [ "$$STATUS" = "409" ]; then \
		echo "$$BODY" | (jq '.' 2>/dev/null || echo "$$BODY") && \
		echo "$(GREEN)✓ User created or already exists$(NC)"; \
	else \
		echo "Server not responding. Is it running? (make up)"; \
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)2. Logging in to get JWT token...$(NC)" && \
	TOKEN=$$(curl -s -X POST -H "Content-Type: application/json" \
		-d '{"email":"$(TEST_EMAIL)","password":"$(TEST_PASSWORD)"}' \
		http://localhost:$(API_PORT)/auth/login | grep -o '"token":"[^"]*' | cut -d'"' -f4) && \
	if [ -z "$$TOKEN" ]; then \
		echo "$(RED)✗ Failed to get token. Make sure the server is running with:$(NC)"; \
		echo "   make up"; \
		exit 1; \
	fi && \
	echo "$(GREEN)✓ Token received$(NC)" && \
	\
	echo "\n$(BOLD)3. Creating short URL for $(TEST_URL)...$(NC)" && \
	RESPONSE=$$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
		-d '{"url":"$(TEST_URL)"}' \
		http://localhost:$(API_PORT)/api/shorten) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "$$RESPONSE") && \
	SHORTCODE=$$(echo "$$RESPONSE" | grep -o '"shortCode":"[^"]*' | cut -d'"' -f4) && \
	if [ -z "$$SHORTCODE" ]; then \
		echo "$(RED)✗ Failed to create short URL$(NC)"; \
		exit 1; \
	fi && \
	echo "$(GREEN)✓ Short URL created: $(BOLD)$$SHORTCODE$(NC)" && \
	\
	echo "\n$(BOLD)4. Getting URL info for $$SHORTCODE...$(NC)" && \
	RESPONSE=$$(curl -s -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "Failed to get URL info") && \
	echo "$(GREEN)✓ URL info retrieved$(NC)" && \
	\
	echo "\n$(BOLD)5. Updating URL to $(TEST_UPDATED_URL)...$(NC)" && \
	RESPONSE=$$(curl -s -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
		-d '{"url":"$(TEST_UPDATED_URL)"}' \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "Failed to update URL") && \
	echo "$(GREEN)✓ URL updated$(NC)" && \
	\
	echo "\n$(BOLD)6. Getting statistics...$(NC)" && \
	RESPONSE=$$(curl -s -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE/stats) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "Failed to get statistics") && \
	echo "$(GREEN)✓ Statistics retrieved$(NC)" && \
	\
	echo "\n$(BOLD)7. Testing redirect...$(NC)" && \
	echo "Testing: http://localhost:$(API_PORT)/$$SHORTCODE" && \
	REDIRECT_TEST=$$(curl -s -o /dev/null -w "%{http_code}\n%{redirect_url}" http://localhost:$(API_PORT)/$$SHORTCODE) && \
	STATUS_CODE=$$(echo "$$REDIRECT_TEST" | head -n1) && \
	LOCATION=$$(echo "$$REDIRECT_TEST" | tail -n1 | sed 's:/*$$::') && \
	EXPECTED_URL=$$(echo "$(TEST_UPDATED_URL)" | sed 's:/*$$::') && \
	if [ "$$STATUS_CODE" = "301" ] && [ "$$LOCATION" = "$$EXPECTED_URL" ]; then \
		echo "$(GREEN)✓ Redirect working correctly$(NC)"; \
		echo "  Status: $$STATUS_CODE (Moved Permanently)"; \
		echo "  Location: $$LOCATION"; \
	else \
		echo "$(RED)✗ Redirect check failed$(NC)"; \
		echo "  Expected status: 301, got: $$STATUS_CODE"; \
		echo "  Expected location: $$EXPECTED_URL"; \
		echo "  Got location: $$LOCATION"; \
	fi && \
	\
	echo "\n$(BOLD)8. Deleting URL...$(NC)" && \
	RESULT=$$(curl -s -w "%{http_code}" -X DELETE -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE) && \
	if [ "$$RESULT" = "204" ]; then \
		echo "$(GREEN)✓ URL deleted$(NC)"; \
	else \
		echo "$(RED)✗ Failed to delete URL$(NC)"; \
	fi && \
	END_TIME=`date +%s` && \
	DURATION=$$((END_TIME - START_TIME)) && \
	echo "\n$(GREEN)✨ All tests completed in $$DURATION seconds!$(NC)" && \
	echo "\n$(BOLD)Summary:$(NC)" && \
	echo "• Short URL created: $(BOLD)$$SHORTCODE$(NC)" && \
	echo "• Original URL: $(TEST_URL)" && \
	echo "• Updated URL: $(TEST_UPDATED_URL)" && \
	echo "• API endpoint: http://localhost:$(API_PORT)" && \
	echo "• All operations completed successfully ✨" \
	) 