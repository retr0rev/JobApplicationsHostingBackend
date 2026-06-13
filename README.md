# JobApplicationsHostingBackend

Backend API for managing job applications. Built with Go, Chi router, and SQLite.

## Tech Stack

- **Go 1.22** — language
- **Chi v5** — HTTP router
- **SQLite** — database (via `modernc.org/sqlite`, pure Go, no CGO)
- **bcrypt** — password hashing
- **JWT** (HS256) — authentication tokens

## Quick Start

```bash
git clone https://github.com/retr0rev/JobApplicationsHostingBackend.git
cd JobApplicationsHostingBackend

# Generate a strong JWT secret (required, minimum 32 characters)
export JWT_SECRET=$(openssl rand -hex 32)

# Optional: set a real frontend origin (otherwise the server logs a WARNING
# and falls back to "*", which is not recommended for production)
export CORS_ORIGIN=https://yourfrontend.com

# Run
go run ./cmd/server
```

Server starts on `http://localhost:8080` by default.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | **Yes** | — | Secret key for JWT signing. **Must be at least 32 characters.** Generate with `openssl rand -hex 32`. |
| `PORT` | No | `8080` | Server port. |
| `DB_PATH` | No | `./database/database.db` | Path to the SQLite database file (auto-created on first run). |
| `APP_URL` | No | `http://localhost:8080` | Public URL used as the base for email verification and password reset links. |
| `CORS_ORIGIN` | No | `*` | Allowed CORS origin. **Set a specific origin in production** — the server logs a `WARNING` if this is unset or `*`. |
| `TLS_ENABLED` | No | `false` | Set to `true` to emit the `Strict-Transport-Security` header. Use when serving over HTTPS. |
| `SMTP_HOST` | No* | — | SMTP server hostname. Required for real email delivery. |
| `SMTP_PORT` | No | `587` | SMTP port. |
| `SMTP_USER` | No* | — | SMTP username. |
| `SMTP_PASS` | No* | — | SMTP password. |
| `SMTP_FROM` | No* | — | From address for outgoing emails. |

\* Required only if `SMTP_HOST` is set. Without SMTP, emails are logged to the console (dev mode).

## Security

- **Passwords**: minimum 8 characters, must contain uppercase, lowercase, and digit characters (`SecurePass1` is valid; `password` is not).
- **JWT_SECRET**: must be at least 32 characters; the server refuses to start otherwise.
- **Tokens**: email-verification and password-reset tokens are stored as SHA-256 hashes — a leaked database cannot be used to log into accounts.
- **Emails**: matched case-insensitively (normalized to lowercase), so `User@x.com` and `user@x.com` refer to the same account.
- **Forgot-password**: returns the same generic response and equalized timing whether or not the email exists, to prevent user enumeration.
- **Transport security**: `Strict-Transport-Security` is emitted when `TLS_ENABLED=true`. Terminate TLS at your reverse proxy (Caddy, nginx, etc.) or in front of the binary.
- **Security headers**: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `Cache-Control: no-store`.
- **CORS**: when `CORS_ORIGIN` is a specific value, responses include `Vary: Origin` and `Access-Control-Allow-Credentials: true`, with `Access-Control-Max-Age: 600`.

## API Endpoints

All endpoints return JSON. Authentication uses `Authorization: Bearer <jwt>`.

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/auth/signup` | — | Create account. Sends a verification email. |
| `POST` | `/api/auth/login` | — | Login. Requires a verified email. |
| `GET` | `/api/auth/verify?token=` | — | Verify an email address using the token from the verification email. |
| `POST` | `/api/auth/forgot-password` | — | Request a password-reset email. |
| `POST` | `/api/auth/reset-password` | — | Reset password using a reset token. |

### Job Applications (Client)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/jobs` | Client | List the authenticated client's job applications. |
| `POST` | `/api/jobs` | Client | Create a new job application. |
| `GET` | `/api/jobs/{id}` | Client | Get one of the client's own job applications. |
| `PUT` | `/api/jobs/{id}` | Client | Update one of the client's own job applications. |
| `DELETE` | `/api/jobs/{id}` | Client | Delete one of the client's own job applications. |

### Admin

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/admin/login` | — | Admin login. |
| `GET` | `/api/admin/jobs` | Admin | List every job application across all clients. |
| `PUT` | `/api/admin/jobs/{id}/status` | Admin | Approve or reject a job (`status: "approved"` or `"rejected"`). |

### Other

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check (returns `{"status":"ok"}`). |

## Seed Admin

```bash
# Locally
go run ./cmd/server seed-admin --email=admin@site.com --password=AdminPass123

# The password must satisfy the complexity rules (see Security).
# Re-running with the same email fails with a clear "admin already exists" error.
```

## Docker

```bash
# Build and run
docker compose up --build

# Seed admin (in a second terminal)
docker compose run --rm app ./server seed-admin --email=admin@site.com --password=AdminPass123
```

Create a `.env` file in the project root before running:

```env
JWT_SECRET=your-64-char-hex-secret
APP_URL=https://yourdomain.com
CORS_ORIGIN=https://yourfrontend.com
TLS_ENABLED=true

# Optional — required for real email delivery
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=you@example.com
SMTP_PASS=yourpassword
SMTP_FROM=noreply@example.com
```

> **Note:** `.env` is in `.gitignore` and should never be committed.

## Project Layout

```
.
├── cmd/server/         # main entry point
├── internal/
│   ├── email/          # email sender (SMTP + console fallback)
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # auth, CORS, security headers, validators
│   ├── models/         # request/response shapes and domain types
│   └── repository/     # SQLite access layer
├── pkg/database/       # DB connection + embedded migration
├── database/schema.sql # human-readable schema (mirrors the migration)
├── Dockerfile
├── docker-compose.yml
├── go.mod / go.sum
└── README.md
```
