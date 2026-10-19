# Music Platform Backend

Initial Go monolith scaffold for a music discovery and collections platform.

## Requirements

- Go 1.25+
- Docker Desktop with Docker Compose
- Optional for local generation: `sqlc` and `golang-migrate`

## Local setup

1. Copy `.env.example` to `.env`.
2. Replace `POSTGRES_PASSWORD` with a long random local password.
3. Start the stack:

   ```sh
   docker compose up --build
   ```

4. Check endpoints:

   ```sh
   curl -i http://localhost:8080/healthz
   curl -i http://localhost:8080/readyz
   ```

The first endpoint checks that the HTTP process responds. The second checks PostgreSQL connectivity.

Stop the stack with `Ctrl+C`. To remove the local database volume as well:

```sh
docker compose down -v
```

**Warning:** `docker compose down -v` permanently deletes the local PostgreSQL data volume.

## Go checks

```sh
go mod tidy
go test ./...
go vet ./...
go build ./cmd/server
```

## Migrations

Migrations are applied by the `migrate` service before the app starts. The first migration intentionally creates no lasting domain tables; domain schema will be added in the next verified stage.

## sqlc

After adding SQL queries, generate code with:

```sh
sqlc generate
```

Generated code is intentionally not committed yet because there are no application queries.

## Current scope

This scaffold includes configuration, a PostgreSQL connection pool, HTTP health endpoints, migration runner, Docker Compose, and basic CI. Catalog, authentication, library, collections, editorial content, and admin features will be added in separate stages.
