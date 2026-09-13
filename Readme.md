# Task Manager API

A backend Task Management API built with Go. The project includes user authentication, role-based access control, task management, caching, rate limiting, structured logging, metrics, and distributed tracing.

The main goal of the project was to build a clean backend structure with clear separation between HTTP handling, business logic, and data access, while also covering common production concerns such as authentication, observability, testing, and error handling.

---

## Features

### Authentication & Authorization

* User registration with password validation
* User login with username and password
* JWT-based authentication with access and refresh tokens
* Refresh token endpoint
* Role-based access control (RBAC)
* Protected API endpoints

### Task Management

* Create, read, update, and delete tasks
* Task status management:

    * `pending`
    * `in_progress`
    * `completed`
* Assign tasks to users
* Filter tasks by status and assignee
* Pagination
* Task-level authorization

### Performance & Observability

* Redis caching
* Request rate limiting
* Structured application logging
* Prometheus metrics
* Distributed tracing with OpenTelemetry/Jaeger
* Request recovery middleware

### Security

* Password hashing with bcrypt
* JWT expiration
* Configurable CORS
* Request validation
* Protection against SQL injection through GORM
* Environment-based configuration for sensitive values

---

## Architecture

The application follows a layered architecture. HTTP-specific concerns stay in the API layer, while business rules are handled by services.

```text
                         Client
                           │
                           │ HTTP
                           ▼
                  ┌──────────────────┐
                  │   Gin Router     │
                  └────────┬─────────┘
                           │
                  ┌────────▼─────────┐
                  │    Middleware    │
                  │                  │
                  │ Authentication   │
                  │ Authorization    │
                  │ Logging          │
                  │ CORS             │
                  │ Rate Limiting    │
                  │ Metrics          │
                  │ Recovery         │
                  └────────┬─────────┘
                           │
                  ┌────────▼─────────┐
                  │     Handlers     │
                  │                  │
                  │ Users            │
                  │ Tasks            │
                  │ Health           │
                  └────────┬─────────┘
                           │
                  ┌────────▼─────────┐
                  │    Services      │
                  │                  │
                  │ UserService      │
                  │ TaskService      │
                  │ TokenService     │
                  └───────┬──────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
        PostgreSQL      Redis       Tracing
        Database        Cache       / Jaeger
```

### Main layers

**Router / API layer**

Responsible for registering routes and connecting middleware and handlers.

**Middleware**

Handles cross-cutting concerns such as authentication, authorization, logging, rate limiting, CORS, metrics, recovery, and tracing.

**Handlers**

Handle HTTP requests, validate input, call the appropriate service, and build HTTP responses.

**Services**

Contain the application's business logic. Keeping this logic outside the handlers makes it easier to test and maintain.

**Data layer**

Contains database models and the PostgreSQL and Redis integrations.

---

## Technology Stack

| Area             | Technology              |
| ---------------- | ----------------------- |
| Language         | Go 1.25+                |
| HTTP Framework   | Gin 1.12                |
| Database         | PostgreSQL 17           |
| ORM              | GORM                    |
| Cache            | Redis 8                 |
| Authentication   | JWT                     |
| Password Hashing | bcrypt                  |
| Logging          | Zap / Zerolog           |
| Metrics          | Prometheus              |
| Tracing          | OpenTelemetry / Jaeger  |
| Testing          | Go testing / Testify    |
| Containers       | Docker / Docker Compose |

---

# Getting Started

## Requirements

You will need:

* Docker and Docker Compose
* Git

---

## 1. Clone the repository

```bash
git clone https://github.com/MahdiPeydai/TaskManager-Go.git
cd TaskManager-Go
```

---

## 2. Configure environment variables

Create the environment file from the example:

```bash
cp docker/.env.example docker/.env
```

Then update the values if necessary:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secure_password_123
POSTGRES_DB=taskmanager

REDIS_PASSWORD=redis_password_123

PORT=8080
```

Application configuration is environment-specific and is loaded by the Go configuration package.

---

## 3. Start Project

From the project root:

```bash
docker compose -f docker/docker-compose.yml up -d
```

Check the running containers:

```bash
docker compose -f docker/docker-compose.yml ps
```

To view logs:

```bash
docker compose -f docker/docker-compose.yml logs -f
```

The API is available at:

```text
http://localhost:8080
```

---

# API

## Base URL

```text
http://localhost:8080/api/v1
```

All JSON requests should use:

```http
Content-Type: application/json
```

Authenticated endpoints require:

```http
Authorization: Bearer <access-token>
```

---

## Response Format

Successful responses generally follow this structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {},
  "statusCode": 200,
  "errors": null
}
```

Error responses use the same general structure while providing information about the failure.

---

## HTTP Status Codes

| Status | Meaning                                     |
| ------ | ------------------------------------------- |
| `200`  | Request completed successfully              |
| `201`  | Resource created                            |
| `400`  | Invalid request or validation error         |
| `401`  | Authentication required or token is invalid |
| `403`  | Authenticated but not authorized            |
| `404`  | Resource not found                          |
| `429`  | Rate limit exceeded                         |
| `500`  | Internal server error                       |

---

# API Examples

## Authentication

### Register

```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ali123",
    "firstName": "Ali",
    "lastName": "Ahmadi",
    "email": "ali@example.com",
    "password": "SecurePass123!"
  }'
```

A successful registration returns access and refresh tokens:

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "accessTokenExpireTime": 1663420000,
    "refreshTokenExpireTime": 1663506400
  },
  "statusCode": 201
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ali123",
    "password": "SecurePass123!"
  }'
```

### Refresh token

```bash
curl -X POST http://localhost:8080/api/v1/users/refresh-token \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

---

# Tasks

## Create a task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "title": "Complete API implementation",
    "description": "Implement the remaining API endpoints",
    "status": "in_progress",
    "assignee_id": 1
  }'
```

Example response:

```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "id": 1,
    "title": "Complete API implementation",
    "description": "Implement the remaining API endpoints",
    "status": "in_progress",
    "assignee_id": 1
  },
  "statusCode": 201
}
```

## List tasks

Get the first page:

```bash
curl -X GET "http://localhost:8080/api/v1/tasks?page=1&pageSize=10" \
  -H "Authorization: Bearer <access-token>"
```

Filter by status:

```bash
curl -X GET "http://localhost:8080/api/v1/tasks?status=pending&page=1&pageSize=10" \
  -H "Authorization: Bearer <access-token>"
```

Filter by assignee:

```bash
curl -X GET "http://localhost:8080/api/v1/tasks?assignee_id=1&page=1&pageSize=10" \
  -H "Authorization: Bearer <access-token>"
```

Example response:

```json
{
  "success": true,
  "message": "Tasks retrieved successfully",
  "data": {
    "items": [
      {
        "id": 1,
        "title": "Complete API implementation",
        "description": "Implement the remaining API endpoints",
        "status": "in_progress",
        "assignee_id": 1
      },
      {
        "id": 2,
        "title": "Write integration tests",
        "description": "Add tests for the API endpoints",
        "status": "pending",
        "assignee_id": null
      }
    ],
    "totalCount": 2,
    "page": 1,
    "pageSize": 10
  },
  "statusCode": 200
}
```

## Get a task

```bash
curl -X GET http://localhost:8080/api/v1/tasks/1 \
  -H "Authorization: Bearer <access-token>"
```

## Update a task

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "status": "completed",
    "title": "Complete API implementation"
  }'
```

## Delete a task

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1 \
  -H "Authorization: Bearer <access-token>"
```

---

# Health Check

The health endpoint does not require authentication.

```bash
curl http://localhost:8080/api/v1/health
```

Example response:

```json
{
  "success": true,
  "message": "Server is healthy"
}
```

---

# Testing

The project contains both unit tests and integration tests.

## Unit tests

Run all tests:

```bash
cd src
go test ./... -v
```

Run tests with coverage:

```bash
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

On macOS:

```bash
open coverage.html
```

---

# Integration Tests

The integration tests exercise the API through the HTTP layer rather than calling services directly.

They cover the complete request/response flow, including:

* Request parsing and validation
* HTTP status codes
* Response structure
* Authentication
* Authorization and RBAC
* Error handling
* Route registration
* User endpoints
* Task endpoints
* Health endpoint

This provides confidence that the different layers of the application work together correctly.

## Integration test structure

```text
src/tests/integration/
├── setup_test.go       # Test setup and shared helpers
├── health_test.go      # Health endpoint tests
├── users_test.go       # Registration, login and refresh tests
└── tasks_test.go       # Task CRUD and authorization tests
```

## Test database

Integration tests use a separate PostgreSQL database so that test data is isolated from the development database.

By default:

```text
Database: taskmanager_test
Host:     localhost
Port:     5432
User:     postgres
Password: postgres
```

Redis is mocked in the integration tests where appropriate.

---

## Setting up the test database

### Using Docker

Start the infrastructure:

```bash
docker compose -f docker/docker-compose-test.yml up -d
```

Create the test database:

```bash
docker exec graph-taskmanager-test-postgres \
  createdb -U postgres taskmanager_test
```

If your Docker container has a different name, use the actual PostgreSQL container name shown by:

```bash
docker ps
```

## Run all integration tests

```bash
cd src
go test ./tests/integration -v
```

## Run a specific test suite

Users:

```bash
go test ./tests/integration -v -run TestUsers
```

Tasks:

```bash
go test ./tests/integration -v -run TestTasks
```

Health:

```bash
go test ./tests/integration -v -run TestHealth
```

## Run a specific test

For example:

```bash
go test ./tests/integration -v \
  -run TestUsersEndpoint_Register
```

Or a specific sub-test:

```bash
go test ./tests/integration -v \
  -run TestTasksEndpoint_Create/successful_task_creation
```

## Integration test coverage

```bash
go test ./tests/integration -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

# Test Configuration

The integration test configuration is defined in `setup_test.go`.

The default configuration points to:

```text
PostgreSQL:
  host: localhost
  port: 5432
  database: taskmanager_test
  user: postgres
  password: postgres
```

If a different environment is required, the test configuration can be adjusted in `loadTestConfig()`.

The tests also generate their own test data and clean up resources after execution where applicable.

If a test run is interrupted and leaves data behind, the test database can be cleaned manually:

```sql
DELETE FROM tasks;
DELETE FROM users;
```

---

# Design Decisions and Trade-offs

The project deliberately keeps some parts simple because this is a focused backend task rather than a production-scale distributed system.

## JWT vs. server-side sessions

I chose JWT-based authentication because the API is stateless and does not need to keep an authentication session in server memory.

### Advantages

* Stateless authentication
* Works well with APIs and SPAs
* Easy to use across multiple application instances
* No session lookup is required for every request

### Trade-offs

* Revoking an already-issued access token is not immediate
* Tokens are larger than traditional session IDs
* Token lifetime and refresh-token handling need to be designed carefully

A server-side session stored in Redis would be a reasonable alternative if immediate session revocation were a stronger requirement.

---

## PostgreSQL vs. NoSQL

PostgreSQL was chosen because the application has clear relationships between users and tasks and benefits from transactional consistency.

### Advantages

* ACID transactions
* Strong data integrity
* Relational queries
* Foreign keys and constraints
* Mature indexing and query capabilities

A NoSQL database could be useful for a different workload, but it would not provide a significant advantage for this data model.

---

## Redis vs. in-memory caching

Redis is used instead of an in-process cache.

The main reason is that Redis can be shared by multiple application instances. An in-memory cache would be simpler and potentially faster for a single instance, but each application instance would have its own cache.

Redis also provides built-in expiration and can be operated independently from the API process.

The trade-off is an additional network dependency and slightly more operational complexity.

---

## Structured logging

The application uses structured logging rather than plain text logs.

Structured logs make it easier to search and analyze fields such as:

* request ID
* user ID
* HTTP method
* path
* status code
* error information
* request duration

For a larger production deployment, these logs could be shipped to a centralized system such as Loki or Elasticsearch.

---

## Tracing

Tracing is included to make request execution easier to understand, especially when a request crosses multiple layers or external dependencies.

Jaeger is used as the tracing backend during development.

For a larger production environment, the tracing backend could be replaced without changing the application's core business logic.

---

# Observability

## Prometheus

Metrics are exposed through:

```bash
curl http://localhost:8080/metrics
```

The metrics middleware tracks information about incoming HTTP requests and application performance.

---

## Jaeger

The tracing UI is available at:

```text
http://localhost:16686
```

When the application and Jaeger are running, traces can be inspected through the Jaeger dashboard.

---

# Project Structure

```text
TaskManager-Go/
├── docker/
│   ├── docker-compose.yml
│   ├── .env
│   └── .env.example
│
├── src/
│   ├── cmd/
│   │   └── main.go
│   │
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── users.go
│   │   │   ├── tasks.go
│   │   │   └── health.go
│   │   │
│   │   ├── routers/
│   │   │   ├── users.go
│   │   │   ├── tasks.go
│   │   │   └── health.go
│   │   │
│   │   ├── middlewares/
│   │   │   ├── auth.go
│   │   │   ├── logger.go
│   │   │   ├── cors.go
│   │   │   ├── limiter.go
│   │   │   └── prometheus.go
│   │   │
│   │   ├── dto/
│   │   │   ├── user.go
│   │   │   └── task.go
│   │   │
│   │   ├── helpers/
│   │   │   └── http_response.go
│   │   │
│   │   └── api.go
│   │
│   ├── services/
│   │   ├── user.go
│   │   ├── task.go
│   │   └── token.go
│   │
│   ├── data/
│   │   ├── db/
│   │   ├── cache/
│   │   └── models/
│   │       ├── user.go
│   │       └── task.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── constants/
│   ├── common/
│   │
│   ├── pkg/
│   │   ├── logging/
│   │   └── metrics/
│   │
│   ├── tests/
│   │   ├── integration/
│   │   ├── mocks/
│   │   └── unit/
│   │
│   ├── go.mod
│   └── go.sum
│
├── README.md
└── .gitignore
```

---

## Test database does not exist

Create it with:

```bash
docker exec taskmanager-postgres \
  createdb -U postgres taskmanager_test
```

Or locally:

```bash
createdb taskmanager_test
```

---

## Connection refused

Make sure PostgreSQL is listening on port `5432`:

```bash
docker compose -f docker/docker-compose.yml ps
```

Also verify that the credentials in the application/test configuration match the PostgreSQL configuration.

---

## Tests fail with authentication errors

Check that the JWT configuration used by the tests matches the configuration expected by the authentication middleware.

The integration tests generate authentication tokens specifically for the test environment, so changes to token claims, signing keys, or expiration rules may require corresponding test updates.

---

## Port already in use

If port `8080` is already occupied, either stop the process using it or change the application's configured port.

For Docker services, ports can be changed in:

```text
docker/docker-compose.yml
```

For example:

```yaml
ports:
  - "8081:8080"
```

---

# Production Considerations

This project is intended as an interview/task project, but a few additional steps would be required before running it as a production service.

Some examples include:

1. Use HTTPS in front of the API.
2. Keep all secrets in environment variables or a secret manager.
3. Configure database backups and recovery procedures.
4. Use a centralized logging system.
5. Configure Prometheus and tracing with appropriate retention.
6. Tune rate limits based on actual traffic patterns.
7. Use appropriate database connection pool limits.
8. Add health/readiness checks for external dependencies.
9. Add database migrations as a dedicated deployment step.
10. Rotate JWT signing secrets according to the application's security requirements.

---

# Useful Commands

### Start infrastructure

```bash
docker compose -f docker/docker-compose.yml up -d
```

### Stop infrastructure

```bash
docker compose -f docker/docker-compose.yml down
```

### View logs

```bash
docker compose -f docker/docker-compose.yml logs -f
```

### Run application

```bash
cd src
go run ./cmd/main.go
```

### Run all tests

```bash
cd src
go test ./... -v
```

### Run integration tests

```bash
cd src
go test ./tests/integration -v
```

### Run with race detector

```bash
cd src
go test ./... -race
```

---

# Repository

GitHub:

https://github.com/MahdiPeydai/TaskManager-Go

---

## License

This project is licensed under the MIT License.
