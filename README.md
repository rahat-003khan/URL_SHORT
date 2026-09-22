# URL Shortener - Full Simple Project

A beginner-friendly URL Shortener API and frontend built with Go.

## Technologies

- Go
- net/http
- JSON
- HTML
- CSS
- JavaScript

No external Go package is required.

## Features

1. Create a short URL.
2. Random 6-character code.
3. Custom alias.
4. Redirect to original URL.
5. URL validation.
6. Click counter.
7. Statistics API.
8. JSON file persistence.
9. CORS support.
10. Simple frontend.

## Project Structure

```text
url-shortener/
│
├── main.go
├── go.mod
├── data.json       (created automatically)
├── README.md
│
└── public/
    ├── index.html
    ├── style.css
    └── script.js
```

## Run the project

Open PowerShell inside the project folder:

```powershell
go run main.go
```

Then open:

```text
http://localhost:8080
```

## API Endpoints

### 1. Health Check

```http
GET /health
```

Example:

```text
http://localhost:8080/health
```

### 2. Create Short URL

```http
POST /shorten
Content-Type: application/json
```

Body:

```json
{
    "url": "https://www.google.com",
    "alias": ""
}
```

Response:

```json
{
    "short_url": "http://localhost:8080/aB12x9",
    "code": "aB12x9"
}
```

### 3. Custom Alias

Body:

```json
{
    "url": "https://www.google.com",
    "alias": "google"
}
```

Response:

```json
{
    "short_url": "http://localhost:8080/google",
    "code": "google"
}
```

Then:

```text
http://localhost:8080/google
```

redirects to Google.

### 4. Statistics

```http
GET /api/stats/google
```

Example response:

```json
{
    "code": "google",
    "original_url": "https://www.google.com",
    "clicks": 5,
    "created_at": "2026-09-15T09:00:00Z"
}
```

## How data is stored

The project uses a simple `data.json` file instead of a database.

Example:

```json
{
    "links": {
        "google": {
            "code": "google",
            "original_url": "https://www.google.com",
            "clicks": 2,
            "created_at": "2026-09-15T09:00:00Z"
        }
    }
}
```

This means data remains after restarting the Go server.

## Testing with PowerShell

Health:

```powershell
curl http://localhost:8080/health
```

Create short URL:

```powershell
Invoke-RestMethod `
    -Uri http://localhost:8080/shorten `
    -Method POST `
    -ContentType "application/json" `
    -Body '{"url":"https://www.google.com","alias":"google"}'
```

Statistics:

```powershell
curl http://localhost:8080/api/stats/google
```

## Important

This is an educational/local project. It is not production-ready.

For a production version, add:

- PostgreSQL/MySQL
- User authentication
- Rate limiting
- HTTPS
- Better random ID generation
- Admin dashboard
- QR code
- Detailed analytics
