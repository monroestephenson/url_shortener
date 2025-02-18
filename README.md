# URL Shortener Service

A production-ready URL shortener service with Redis caching, MySQL storage, and Prometheus metrics.

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21 or higher

### Running the Service

1. Clone the repository:
   ```bash
   git clone <your-repo-url>
   cd url-shortener
   ```

2. Start the service:
   ```bash
   make start
   ```

This will:
- Start MySQL and Redis in Docker containers
- Run database migrations
- Start the URL shortener service

The service will be available at `http://localhost:3000`

To stop everything:
```bash
make stop
```

## API Documentation

The API documentation is available in two formats:

1. **Swagger UI**: Visit `http://localhost:3000/docs` in your browser for an interactive API documentation
2. **OpenAPI Specification**: Available at `http://localhost:3000/swagger.yaml`

### Key Endpoints

1. **Authentication**
   - POST `/auth/signup` - Create a new user
   - POST `/auth/login` - Get JWT token

2. **URL Management**
   - POST `/api/shorten` - Create short URL
   - GET `/api/shorten/{shortCode}` - Get URL details
   - PUT `/api/shorten/{shortCode}` - Update URL
   - DELETE `/api/shorten/{shortCode}` - Delete URL
   - GET `/api/shorten/{shortCode}/stats` - Get URL statistics

3. **Redirect**
   - GET `/{shortCode}` - Redirect to original URL

4. **Monitoring**
   - GET `/metrics` - Prometheus metrics

## Features

- ✨ URL shortening with cryptographically secure short codes
- 🔒 JWT Authentication
- 🚀 Redis caching for fast access
- 📊 Prometheus metrics
- 🛡️ Rate limiting
- 📝 Access statistics
- 📚 OpenAPI/Swagger documentation

## Development

- Run tests: `make test`
- Run API tests: `make test-api`
- Run linters: `make lint`
- Build binary: `make build`
- Clean up: `make clean`

## Monitoring

Access Prometheus metrics at: `http://localhost:3000/metrics`

## Available Make Commands

- `make start` - Start all services
- `make stop` - Stop all services
- `make docker-up` - Start only the Docker containers
- `make docker-down` - Stop the Docker containers
- `make test` - Run tests
- `make test-api` - Run API integration tests
- `make lint` - Run linters
- `make build` - Build the binary
- `make clean` - Clean up

## Environment Variables

The following environment variables can be configured:

- `PORT` - API server port (default: 3000)
- `MYSQL_DSN` - MySQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - Secret for JWT tokens

## Security Features

- Secure short code generation using crypto/rand
- JWT-based authentication
- Rate limiting per IP
- URL validation and sanitization
- Protection against malicious URLs
- HTTPS scheme enforcement
