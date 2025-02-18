# URL Shortener Service

A RESTful API service that allows users to create, manage, and use shortened URLs. Built with Go and MySQL.

## Features

- Create short URLs from long URLs
- Retrieve original URLs using short codes
- Update existing short URLs
- Delete short URLs
- Track access statistics for each short URL
- Automatic redirection from short URL to original URL

## Prerequisites

- Go 1.21 or higher
- MySQL 9.x
- Homebrew (for macOS users)

## Setup

1. **Install MySQL:**
   ```bash
   brew install mysql
   brew services start mysql
   ```

2. **Create the database:**
   ```bash
   mysql -u root -e "CREATE DATABASE url_shortener;"
   ```

3. **Run the application:**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`.

## API Documentation

### Create a Short URL
- **Method:** POST
- **Endpoint:** `/shorten`
- **Body:**
  ```json
  {
    "url": "https://www.example.com/some/long/url"
  }
  ```
- **Response:** (201 Created)
  ```json
  {
    "id": 1,
    "shortCode": "abc123",
    "originalUrl": "https://www.example.com/some/long/url",
    "accessCount": 0,
    "createdAt": "2024-02-18T10:00:00Z",
    "updatedAt": "2024-02-18T10:00:00Z"
  }
  ```

### Get Original URL Info
- **Method:** GET
- **Endpoint:** `/shorten/{shortCode}`
- **Response:** (200 OK)
  ```json
  {
    "id": 1,
    "shortCode": "abc123",
    "originalUrl": "https://www.example.com/some/long/url",
    "accessCount": 5,
    "createdAt": "2024-02-18T10:00:00Z",
    "updatedAt": "2024-02-18T10:00:00Z"
  }
  ```

### Update Short URL
- **Method:** PUT
- **Endpoint:** `/shorten/{shortCode}`
- **Body:**
  ```json
  {
    "url": "https://www.example.com/updated/url"
  }
  ```
- **Response:** (200 OK)
  ```json
  {
    "id": 1,
    "shortCode": "abc123",
    "originalUrl": "https://www.example.com/updated/url",
    "accessCount": 5,
    "createdAt": "2024-02-18T10:00:00Z",
    "updatedAt": "2024-02-18T10:30:00Z"
  }
  ```

### Delete Short URL
- **Method:** DELETE
- **Endpoint:** `/shorten/{shortCode}`
- **Response:** (204 No Content)

### Get URL Statistics
- **Method:** GET
- **Endpoint:** `/shorten/{shortCode}/stats`
- **Response:** (200 OK)
  ```json
  {
    "id": 1,
    "shortCode": "abc123",
    "originalUrl": "https://www.example.com/some/long/url",
    "accessCount": 5,
    "createdAt": "2024-02-18T10:00:00Z",
    "updatedAt": "2024-02-18T10:00:00Z"
  }
  ```

### Use Short URL (Redirect)
- **Method:** GET
- **Endpoint:** `/{shortCode}`
- **Behavior:** Redirects to the original URL
- **Response:** (301 Moved Permanently)

## Example Usage

Here are some example curl commands to interact with the API:

```bash
# Create a short URL
curl -X POST -H "Content-Type: application/json" \
     -d '{"url":"https://www.example.com/very/long/url"}' \
     http://localhost:8080/shorten

# Get URL info
curl http://localhost:8080/shorten/abc123

# Update URL
curl -X PUT -H "Content-Type: application/json" \
     -d '{"url":"https://www.example.com/updated/url"}' \
     http://localhost:8080/shorten/abc123

# Get statistics
curl http://localhost:8080/shorten/abc123/stats

# Delete URL
curl -X DELETE http://localhost:8080/shorten/abc123
```

To use a shortened URL, simply open `http://localhost:8080/abc123` in your browser, where `abc123` is your short code.

## Error Handling

The API returns appropriate HTTP status codes:
- 200: Success
- 201: Created
- 204: No Content (successful deletion)
- 400: Bad Request (invalid input)
- 404: Not Found
- 500: Internal Server Error

## Development

### Environment Variables

- `MYSQL_DSN`: MySQL connection string (default: "root@tcp(127.0.0.1:3306)/url_shortener?parseTime=true")

### Database Schema

The service automatically creates the required database table (`short_urls`) with the following structure:
- `id`: Auto-incrementing primary key
- `short_code`: Unique identifier for the shortened URL
- `original_url`: The original long URL
- `access_count`: Number of times the URL has been accessed
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last update

## Security Considerations

For production deployment, consider:
- Adding authentication/authorization
- Setting up HTTPS
- Configuring proper MySQL credentials
- Implementing rate limiting
- Adding input validation for URLs
- Setting up monitoring and logging
