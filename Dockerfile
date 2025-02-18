# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download and verify dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/url-shortener ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary and config from builder
COPY --from=builder /app/url-shortener .
COPY --from=builder /app/config ./config

# Create non-root user
RUN adduser -D appuser
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./url-shortener"] 