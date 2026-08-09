# Plan API

A Go-based REST API for goals management with OpenAPI/Swagger documentation.

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL database

### Environment Variables

```bash
# Required for database connection
DATABASE_URL=postgres://user:password@localhost:5432/plan

# Optional
PORT=8080                    # Default: 8080
GIN_MODE=release             # Default: release
CORS_ALLOWED_ORIGINS=https://example.com  # Default: *
AUTH_DISABLED=true           # For development only
```

### Run

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/server/main.go
```

### Development with Auth Disabled

```bash
AUTH_DISABLED=true go run cmd/server/main.go
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/goals` | List all goals |
| POST | `/api/goals` | Create a goal |
| GET | `/api/goals/:id` | Get a goal |
| PATCH | `/api/goals/:id` | Update a goal |
| DELETE | `/api/goals/:id` | Delete a goal |
| GET | `/api/goals/prioritized` | Get prioritized goals |
| POST | `/api/goals/relations` | Create a relation |
| PATCH | `/api/goals/relations` | Update a relation |
| DELETE | `/api/goals/relations` | Delete a relation |
| POST | `/api/goals/dependencies` | Create a dependency |
| DELETE | `/api/goals/dependencies` | Delete a dependency |

## Authentication

The API uses Bearer token authentication. Include the header:

```
Authorization: Bearer <token>
```

For development, use the format `user_<id>` (e.g., `user_1` for user ID 1).

## Swagger UI

When the server is running, access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

## OpenAPI Spec

The OpenAPI specification is available at:

```
http://localhost:8080/swagger/doc.json
```

## Project Structure

```
src/api/
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── database/        # Database connection
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Auth and CORS middleware
│   ├── models/         # Data models
│   └── repository/     # Database operations
├── docs/               # Swagger documentation
├── go.mod
└── README.md
```

## TODO

- [ ] Replace development auth with proper Zitadel/OIDC JWT validation
- [ ] Add database migrations (currently assumes schema exists)
- [ ] Add request validation with proper error messages
- [ ] Add rate limiting
- [ ] Add logging middleware
- [ ] Add unit tests
