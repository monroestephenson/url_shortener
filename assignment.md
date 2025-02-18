You are required to create a simple RESTful API that allows users to shorten long URLs. The API should provide endpoints to create, retrieve, update, and delete short URLs. It should also provide statistics on the number of times a short URL has been accessed.

URL Shortening Service

Requirements
You should create a RESTful API for a URL shortening service. The API should allow users to perform the following operations:

Create a new short URL
Retrieve an original URL from a short URL
Update an existing short URL
Delete an existing short URL
Get statistics on the short URL (e.g., number of times accessed)
You can optionally setup a minimal frontend to interact with the API and setup redirects for the short URLs to the original URLs.

API Endpoints
Given below are the details for each API operation.

Create Short URL
Create a new short URL using the POST method

POST /shorten
{
  "url": "https://www.example.com/some/long/url"
}
The endpoint should validate the request body and return a 201 Created status code with the newly created short URL i.e.

{
  "id": "1",
  "url": "https://www.example.com/some/long/url",
  "shortCode": "abc123",
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:00:00Z"
}
or a 400 Bad Request status code with error messages in case of validation errors. Short codes must be unique and should be generated randomly.

Retrieve Original URL
Retrieve the original URL from a short URL using the GET method

GET /shorten/abc123
The endpoint should return a 200 OK status code with the original URL i.e.

{
  "id": "1",
  "url": "https://www.example.com/some/long/url",
  "shortCode": "abc123",
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:00:00Z"
}
or a 404 Not Found status code if the short URL was not found. Your frontend should be responsible for retrieving the original URL using the short URL and redirecting the user to the original URL.

Update Short URL
Update an existing short URL using the PUT method

PUT /shorten/abc123
{
  "url": "https://www.example.com/some/updated/url"
}
The endpoint should validate the request body and return a 200 OK status code with the updated short URL i.e.

{
  "id": "1",
  "url": "https://www.example.com/some/updated/url",
  "shortCode": "abc123",
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:30:00Z"
}
or a 400 Bad Request status code with error messages in case of validation errors. It should return a 404 Not Found status code if the short URL was not found.

Delete Short URL
Delete an existing short URL using the DELETE method

DELETE /shorten/abc123
The endpoint should return a 204 No Content status code if the short URL was successfully deleted or a 404 Not Found status code if the short URL was not found.

Get URL Statistics
Get statistics for a short URL using the GET method

GET /shorten/abc123/stats
The endpoint should return a 200 OK status code with the statistics i.e.

{
  "id": "1",
  "url": "https://www.example.com/some/long/url",
  "shortCode": "abc123",
  "createdAt": "2021-09-01T12:00:00Z",
  "updatedAt": "2021-09-01T12:00:00Z",
  "accessCount": 10
}
or a 404 Not Found status code if the short URL was not found.

