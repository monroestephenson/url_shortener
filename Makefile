.PHONY: build test run clean lint help docker-up docker-down start stop

# Go related variables
BINARY_NAME=urlshortener
MAIN_PATH=./cmd/server
API_PORT=3000

# MySQL variables
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=urlshortener
MYSQL_PORT=3306
MYSQL_DSN=root:$(MYSQL_ROOT_PASSWORD)@tcp(127.0.0.1:$(MYSQL_PORT))/$(MYSQL_DATABASE)?parseTime=true

# Redis variables
REDIS_URL=localhost:6380

# Test variables
TEST_EMAIL=test@example.com
TEST_PASSWORD=testpass123
TEST_URL=https://www.google.com
TEST_UPDATED_URL=https://www.github.com
TEST_INVALID_URL=not-a-valid-url
TEST_LONG_URL=https://www.example.com/very/long/path/that/should/still/work/fine/with/our/system/and/test/the/capacity
TEST_MALICIOUS_URL=javascript:alert(1)

# Colors for output
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

docker-up: ## Start all docker containers
	@echo "$(BOLD)Starting docker containers...$(NC)"
	@echo "Cleaning up any existing containers..."
	docker compose down --remove-orphans
	docker rm -f urlshortener-mysql urlshortener-redis 2>/dev/null || true
	@echo "Starting fresh containers..."
	docker compose up -d
	@echo "$(GREEN)✓ Containers started$(NC)"
	@echo "Waiting for services to be ready..."
	@sleep 10

docker-down: ## Stop all docker containers
	@echo "$(BOLD)Stopping docker containers...$(NC)"
	docker compose down
	@echo "$(GREEN)✓ Containers stopped$(NC)"

start: docker-up ## Start the application and all its dependencies
	@echo "$(BOLD)Starting URL shortener application...$(NC)"
	PORT=$(API_PORT) \
	MYSQL_DSN="$(MYSQL_DSN)" \
	REDIS_URL="$(REDIS_URL)" \
	go run $(MAIN_PATH)

stop: docker-down ## Stop the application and all its dependencies
	@echo "$(GREEN)✓ Application stopped$(NC)"

build: ## Build the application
	@echo "$(BOLD)Building application...$(NC)"
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build complete$(NC)"

test: ## Run tests
	@echo "$(BOLD)Running tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

clean: ## Clean up
	@echo "$(BOLD)Cleaning up...$(NC)"
	go clean
	rm -f $(BINARY_NAME)
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

lint: ## Run linters
	@echo "$(BOLD)Running linters...$(NC)"
	go vet ./...
	go fmt ./...
	@echo "$(GREEN)✓ Lint complete$(NC)"

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
	echo "$(BOLD)1. Testing Authentication...$(NC)" && \
	echo "\n$(BOLD)1.1 Creating test user...$(NC)" && \
	RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" \
		-d '{"username":"$(TEST_EMAIL)","password":"$(TEST_PASSWORD)"}' \
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
	echo "\n$(BOLD)1.2 Testing invalid login...$(NC)" && \
	RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" \
		-d '{"username":"wrong@example.com","password":"wrongpass"}' \
		http://localhost:$(API_PORT)/auth/login) && \
	STATUS=$$(echo "$$RESPONSE" | tail -n1) && \
	if [ "$$STATUS" = "401" ]; then \
		echo "$(GREEN)✓ Invalid login rejected$(NC)"; \
	else \
		echo "$(RED)✗ Invalid login not properly handled$(NC)"; \
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)1.3 Logging in with correct credentials...$(NC)" && \
	LOGIN_RESPONSE=$$(curl -s -X POST -H "Content-Type: application/json" \
		-d '{"username":"$(TEST_EMAIL)","password":"$(TEST_PASSWORD)"}' \
		http://localhost:$(API_PORT)/auth/login) && \
	echo "Login Response: $$LOGIN_RESPONSE" && \
	TOKEN=$$(echo "$$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4) && \
	if [ -z "$$TOKEN" ]; then \
		echo "$(RED)✗ Failed to get token$(NC)"; \
		exit 1; \
	fi && \
	echo "$(GREEN)✓ Token received$(NC)" && \
	\
	echo "\n$(BOLD)2. Testing URL Validation...$(NC)" && \
	echo "\n$(BOLD)2.1 Testing invalid URL...$(NC)" && \
	RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
		-d '{"url":"$(TEST_INVALID_URL)"}' \
		http://localhost:$(API_PORT)/api/shorten) && \
	STATUS=$$(echo "$$RESPONSE" | tail -n1) && \
	if [ "$$STATUS" = "400" ]; then \
		echo "$(GREEN)✓ Invalid URL rejected$(NC)"; \
	else \
		echo "$(RED)✗ Invalid URL not properly handled$(NC)"; \
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)2.2 Testing malicious URL...$(NC)" && \
	RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
		-d '{"url":"$(TEST_MALICIOUS_URL)"}' \
		http://localhost:$(API_PORT)/api/shorten) && \
	STATUS=$$(echo "$$RESPONSE" | tail -n1) && \
	if [ "$$STATUS" = "400" ]; then \
		echo "$(GREEN)✓ Malicious URL rejected$(NC)"; \
	else \
		echo "$(RED)✗ Malicious URL not properly handled$(NC)"; \
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)3. Testing URL Operations...$(NC)" && \
	echo "\n$(BOLD)3.1 Creating short URL...$(NC)" && \
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
	echo "\n$(BOLD)3.2 Creating short URL for long URL...$(NC)" && \
	RESPONSE=$$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
		-d '{"url":"$(TEST_LONG_URL)"}' \
		http://localhost:$(API_PORT)/api/shorten) && \
	LONG_SHORTCODE=$$(echo "$$RESPONSE" | grep -o '"shortCode":"[^"]*' | cut -d'"' -f4) && \
	if [ -z "$$LONG_SHORTCODE" ]; then \
		echo "$(RED)✗ Failed to create short URL for long URL$(NC)"; \
		exit 1; \
	fi && \
	echo "$(GREEN)✓ Long URL shortened: $(BOLD)$$LONG_SHORTCODE$(NC)" && \
	\
	echo "\n$(BOLD)3.3 Getting URL info...$(NC)" && \
	RESPONSE=$$(curl -s -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "$$RESPONSE") && \
	echo "$(GREEN)✓ URL info retrieved$(NC)" && \
	\
	echo "\n$(BOLD)3.4 Updating URL...$(NC)" && \
	RESPONSE=$$(curl -s -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
		-d '{"url":"$(TEST_UPDATED_URL)"}' \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "$$RESPONSE") && \
	echo "$(GREEN)✓ URL updated$(NC)" && \
	\
	echo "\n$(BOLD)3.5 Getting statistics...$(NC)" && \
	RESPONSE=$$(curl -s -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE/stats) && \
	echo "$$RESPONSE" | (jq '.' 2>/dev/null || echo "$$RESPONSE") && \
	echo "$(GREEN)✓ Statistics retrieved$(NC)" && \
	\
	echo "\n$(BOLD)4. Testing Redirects...$(NC)" && \
	echo "\n$(BOLD)4.1 Testing normal redirect...$(NC)" && \
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
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)4.2 Testing non-existent shortcode...$(NC)" && \
	RESPONSE=$$(curl -s -w "%{http_code}" http://localhost:$(API_PORT)/nonexistent) && \
	STATUS=$$(echo "$$RESPONSE" | tail -n1) && \
	if [ "$$STATUS" = "404" ]; then \
		echo "$(GREEN)✓ Non-existent shortcode handled correctly$(NC)"; \
	else \
		echo "$(RED)✗ Non-existent shortcode not properly handled$(NC)"; \
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)5. Cleanup...$(NC)" && \
	echo "\n$(BOLD)5.1 Deleting first URL...$(NC)" && \
	RESULT=$$(curl -s -w "%{http_code}" -X DELETE -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$SHORTCODE) && \
	if [ "$$RESULT" = "204" ]; then \
		echo "$(GREEN)✓ First URL deleted$(NC)"; \
	else \
		echo "$(RED)✗ Failed to delete first URL$(NC)"; \
		exit 1; \
	fi && \
	\
	echo "\n$(BOLD)5.2 Deleting second URL...$(NC)" && \
	RESULT=$$(curl -s -w "%{http_code}" -X DELETE -H "Authorization: Bearer $$TOKEN" \
		http://localhost:$(API_PORT)/api/shorten/$$LONG_SHORTCODE) && \
	if [ "$$RESULT" = "204" ]; then \
		echo "$(GREEN)✓ Second URL deleted$(NC)"; \
	else \
		echo "$(RED)✗ Failed to delete second URL$(NC)"; \
		exit 1; \
	fi && \
	\
	END_TIME=`date +%s` && \
	DURATION=$$((END_TIME - START_TIME)) && \
	echo "\n$(GREEN)✨ All tests completed in $$DURATION seconds!$(NC)" && \
	echo "\n$(BOLD)Test Summary:$(NC)" && \
	echo "• Authentication: $(GREEN)✓$(NC)" && \
	echo "• URL Validation: $(GREEN)✓$(NC)" && \
	echo "• URL Operations: $(GREEN)✓$(NC)" && \
	echo "• Redirects: $(GREEN)✓$(NC)" && \
	echo "• Cleanup: $(GREEN)✓$(NC)" && \
	echo "• API endpoint: http://localhost:$(API_PORT)" && \
	echo "$(GREEN)✨ All operations completed successfully$(NC)" \
	) 