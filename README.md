# go-chi-mfa-psql

A production-ready Go API template featuring Google OAuth 2.0 authentication, PASETO tokens, PostgreSQL persistence, and a companion CLI — all wired together and ready to clone and extend.

---

## What it does

This template gives you a working authentication system out of the box:

1. **Browser-based OAuth flow** — a web endpoint hands off to Google via [Goth](https://github.com/markbates/goth), creates the user in PostgreSQL, and stores a session cookie.
2. **CLI OAuth flow** — the CLI spins up a local HTTP server on a random port, opens the Google consent page, captures the authorization code, and exchanges it with the API for a PASETO access + refresh token pair that is stored in `~/.config/go-chi-mfa-dynamo-cli/tokens.json`.
3. **Token refresh** — a custom `http.RoundTripper` in the CLI transparently retries any `401` response by exchanging the refresh token before propagating the error.
4. **Protected endpoints** — an `AuthMiddleware` validates the PASETO access token on every request and injects the user email into the request context.

### Endpoints

| Method | Path | Auth required | Description |
|--------|------|:---:|-------------|
| `GET` | `/auth/` | — | Begin Google browser OAuth flow |
| `GET` | `/auth/google/callback` | — | Google redirects here after consent |
| `POST` | `/auth/google` | — | CLI token exchange (code → PASETO pair) |
| `POST` | `/auth/refresh` | — | Exchange a refresh token for a new pair |
| `GET` | `/auth/me` | ✓ | Return the authenticated user's email |

### CLI commands

```
cli login    --client-id <id>  --api-url http://localhost:8080
cli whoami   --api-url http://localhost:8080
```

---

## Project layout

```
.
├── cmd/
│   ├── api/main.go          # API entry point
│   └── cli/main.go          # CLI entry point
├── internal/
│   ├── api/
│   │   ├── routes.go        # chi router setup
│   │   └── middlewares/
│   │       └── auth.go      # PASETO validation middleware
│   ├── auth/
│   │   ├── controller/      # HTTP handlers
│   │   ├── model/           # User + Provider types
│   │   ├── repositories/    # PostgreSQL queries
│   │   └── services/        # UserService, TokenService
│   ├── cli/
│   │   ├── auth/            # Google PKCE flow + token file store
│   │   ├── client/          # AuthTransport (auto-refresh)
│   │   └── commands/        # cobra commands
│   └── config/              # Env config, pgxpool setup, DB migrations, Goth, logger
├── tf/                      # Terraform – AWS RDS PostgreSQL
├── test/auth/               # Hurl integration test
├── process-compose.yaml     # Local dev: postgres + pgadmin containers
├── devbox.json              # Reproducible dev environment
└── air.api.toml             # Live-reload config for the API
```

---

## Getting started

### Prerequisites

- [Devbox](https://www.jetify.com/devbox) (installs Go, Air, psql, Podman, Hurl automatically)
- A Google Cloud project with an OAuth 2.0 client ID

### 1. Clone and configure

```bash
cp .env.example .env
# Fill in GOOGLE_ACCESS_KEY_ID and GOOGLE_SECRET_ACCESS_KEY
```

### 2. Start local dependencies

```bash
devbox run up
# Starts PostgreSQL on :5432 and pgAdmin on :8001
```

### 3. Run the API

```bash
# With live-reload
air -c air.api.toml

# Or directly
go run ./cmd/api/main.go
```

The API runs on `http://localhost:8080`. On first startup it runs migrations and creates the `users` table automatically.

### 4. Try the CLI

```bash
go run ./cmd/cli/main.go login --client-id $GOOGLE_ACCESS_KEY_ID
go run ./cmd/cli/main.go whoami
```

---

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | API bind address |
| `PORT` | `8080` | API port |
| `DEBUG` | `false` | Enables development logger |
| `TOKEN_SECRET` | *(see .env.example)* | 32-byte key for PASETO encryption — **change this** |
| `POSTGRES_HOST` | `localhost` | |
| `POSTGRES_PORT` | `5432` | |
| `POSTGRES_USER` | `postgres` | |
| `POSTGRES_PASSWORD` | `postgres` | |
| `POSTGRES_DB` | `app` | |
| `POSTGRES_SSLMODE` | `disable` | Set to `require` in production |
| `GOOGLE_ACCESS_KEY_ID` | *(required)* | Google OAuth client ID |
| `GOOGLE_SECRET_ACCESS_KEY` | *(required)* | Google OAuth client secret |

---

## Deploying to AWS

Terraform files are provided under `tf/` to provision an RDS PostgreSQL instance.

```bash
cd tf
terraform init
terraform apply \
  -var="vpc_id=vpc-xxx" \
  -var="private_subnet_ids=[\"subnet-aaa\",\"subnet-bbb\"]" \
  -var="app_security_group_id=sg-xxx" \
  -var="db_password=changeme"
```

The module creates a subnet group, a dedicated security group (port 5432 open only to your app SG), a parameter group, and an RDS instance. In production (`environment=prod`) it enables multi-AZ, deletion protection, and a 7-day backup window automatically.

---

## How to extend

### Add a new OAuth provider

1. Add the provider credentials to `GothConfig` in `internal/config/config.go`.
2. Register it in `internal/config/goth.go` via `goth.UseProviders(...)`.
3. Add a constant to `internal/auth/model/provider.go`.
4. Wire a new callback route in `internal/auth/controller/controller.go`, following the pattern in `google_callback.go`.

### Add a new protected endpoint

1. Create a handler function in `internal/auth/controller/`.
2. Register it inside the `r.Group(func(r chi.Router) { r.Use(middlewares.AuthMiddleware(...)) ... })` block in `controller.go`.
3. Access the authenticated user's email from context: `r.Context().Value(middlewares.UserEmailKey).(string)`.

### Add a new database table

1. Add a `CREATE TABLE IF NOT EXISTS` statement to the `Migrate` function in `internal/config/postgres.go`.
2. Create a model struct in `internal/auth/model/` (or a new `internal/<feature>/model/` package).
3. Create a repository in `internal/auth/repositories/` (or a new package) that accepts `*pgxpool.Pool`.
4. Create a service that wraps the repository and inject it through `main.go`.

### Add a new CLI command

1. Create a file in `internal/cli/commands/`, using `cobra.Command` following the pattern in `whoami.go`.
2. Register it with `rootCmd.AddCommand(commands.NewYourCommand())` in `cmd/cli/main.go`.
3. Use the `client.AuthTransport` HTTP client for any authenticated API calls — it handles token refresh automatically.

### Add a new migration

Append idempotent SQL to the `Migrate` function in `internal/config/postgres.go`. For anything more complex (ordering, rollbacks), consider replacing it with a migration library like [goose](https://github.com/pressly/goose) and pointing it at the same `pgxpool.Pool`.

---

## Running tests

```bash
# Unit tests
go test ./...

# Integration test (requires a running API on :8081)
hurl --test test/auth/google-auth.hurl --variable host=http://localhost:8081
```

---

## Key dependencies

| Package | Purpose |
|---------|---------|
| [go-chi/chi](https://github.com/go-chi/chi) | HTTP router |
| [markbates/goth](https://github.com/markbates/goth) | OAuth provider abstraction |
| [o1egl/paseto](https://github.com/o1egl/paseto) | PASETO v2 token encryption |
| [jackc/pgx](https://github.com/jackc/pgx) | PostgreSQL driver + connection pool |
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [caarlos0/env](https://github.com/caarlos0/env) | Struct-based env config |
| [go.uber.org/zap](https://pkg.go.dev/go.uber.org/zap) | Structured logging |
