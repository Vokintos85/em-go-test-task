# Subscriptions Service

A Go 1.22 HTTP service for managing customer subscriptions backed by PostgreSQL. The
service exposes CRUD endpoints plus a monthly revenue summary and ships with
OpenAPI documentation and a Swagger UI container for quick exploration.

## Features

- Chi router with structured middleware, JSON handlers, and graceful shutdown.
- PostgreSQL persistence via `pgxpool` including CRUD operations and aggregated
  monthly summaries.
- Strict month parsing (`MM-YYYY` or `YYYY-MM`) stored as the first day of the
  month in UTC.
- Database migration with indexes and trigger-managed `updated_at` timestamps.
- Dockerfile, docker-compose stack, and Makefile helpers for building, running,
  and testing the service.

## Requirements

- Go 1.22+
- PostgreSQL 16 (or Docker to run the included compose stack)
- GNU Make (for the provided convenience targets)

## Getting Started

1. Copy the sample environment and update values as needed:

   ```bash
   cp .env.example .env
   ```

2. Ensure PostgreSQL is running and reachable at `DATABASE_URL`. You can start
   the included database with Docker Compose:

   ```bash
   docker compose up db
   ```

3. Apply the baseline migration (idempotent):

   ```bash
   make migrate
   ```

4. Launch the API locally:

   ```bash
   make run
   ```

   The server listens on `http://localhost:8080` by default and exposes a health
   probe at `/healthz`.

### Docker Compose Stack

To run the entire stack (API, PostgreSQL, and Swagger UI) use:

```bash
docker compose up --build
```

- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8081`

Stop the stack with:

```bash
docker compose down
```

### Testing

Run the Go unit tests with:

```bash
make test
```

## API Overview

All endpoints accept and respond with JSON and require the `Content-Type:
application/json` header for requests with bodies.

### Health

- `GET /healthz` – returns `200 OK` when the service is ready.

### Subscriptions

- `POST /subscriptions` – create a subscription.
- `GET /subscriptions` – list subscriptions ordered by newest first.
- `GET /subscriptions/{id}` – fetch a subscription by ID.
- `PUT /subscriptions/{id}` – update a subscription.
- `DELETE /subscriptions/{id}` – delete a subscription.
- `GET /subscriptions/summary?month=YYYY-MM` – total `amount_cents` for the
  specified month (also accepts `MM-YYYY`).

#### Payload Schema

```json
{
  "user_id": "customer-123",
  "plan": "pro",
  "amount_cents": 9900,
  "currency": "USD",
  "billing_period": "2024-05"
}
```

- `billing_period` must be `MM-YYYY` or `YYYY-MM`; the service stores it as the
  first day of the month in UTC.

#### Example Workflow

Create a subscription:

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"customer-123","plan":"pro","amount_cents":9900,"currency":"USD","billing_period":"2024-05"}'
```

List subscriptions:

```bash
curl http://localhost:8080/subscriptions
```

Fetch the May 2024 summary:

```bash
curl "http://localhost:8080/subscriptions/summary?month=2024-05"
```

## Swagger / OpenAPI

The OpenAPI specification lives at `swagger/openapi.yaml`. When the compose
stack is running, visit `http://localhost:8081` for the interactive Swagger UI.

## Project Layout

- `cmd/server` – application entry point.
- `internal/api` – HTTP handlers and routing.
- `internal/config` – environment configuration loader.
- `internal/database` – PostgreSQL connection setup.
- `internal/subscription` – domain model and repository implementation.
- `migrations` – SQL migrations.
- `swagger` – OpenAPI specification.

## License

This project is licensed under the [MIT License](LICENSE).
