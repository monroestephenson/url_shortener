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

## API Usage

### 1. Create a user
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}' \
  http://localhost:3000/auth/signup
```

### 2. Login to get JWT token
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}' \
  http://localhost:3000/auth/login
```

### 3. Create a short URL
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"url":"https://www.example.com/very/long/url"}' \
  http://localhost:3000/api/shorten
```

### 4. Use the short URL
```bash
curl -L http://localhost:3000/YOUR_SHORT_CODE
```

## Features

- ✨ URL shortening with custom codes
- 🔒 JWT Authentication
- 🚀 Redis caching for fast access
- 📊 Prometheus metrics
- 🛡️ Rate limiting
- 📝 Access statistics

## Development

- Run tests: `make test`
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
- `make lint` - Run linters
- `make build` - Build the binary
- `make clean` - Clean up

## Environment Variables

The following environment variables can be configured:

- `PORT` - API server port (default: 3000)
- `MYSQL_DSN` - MySQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - Secret for JWT tokens
