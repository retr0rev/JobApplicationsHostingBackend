# Job Apps Site

Backend API for managing job applications. Built with Go, Chi router, and SQLite.

## Tech Stack

- **Go 1.22** — language
- **Chi v5** — HTTP router
- **SQLite** — database (via modernc.org/sqlite, pure Go, no CGO)
- **bcrypt** — password hashing
- **JWT** — authentication tokens

## Quick Start

```bash
git clone <repo>
cd job-apps-site

# Set required env vars
export JWT_SECRET=$(openssl rand -hex 32)

# Run
go run ./cmd/server
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | **Yes** | — | Secret key for JWT signing (generate with `openssl rand -hex 32`) |
| `PORT` | No | `8080` | Server port |
| `DB_PATH` | No | `./database/database.db` | Path to SQLite database file |
| `APP_URL` | No | `http://localhost:8080` | Public URL (used in email verification links) |
| `CORS_ORIGIN` | No | `*` | Allowed CORS origin |
| `SMTP_HOST` | No* | — | SMTP server hostname (required for real email delivery) |
| `SMTP_PORT` | No | `587` | SMTP port |
| `SMTP_USER` | No* | — | SMTP username |
| `SMTP_PASS` | No* | — | SMTP password |
| `SMTP_FROM` | No* | — | From address for outgoing emails |

\* Required only if `SMTP_HOST` is set. Without SMTP, emails are logged to console (dev mode).

## API Endpoints

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/auth/signup` | — | Create account (sends verification email) |
| `POST` | `/api/auth/login` | — | Login (requires verified email) |
| `GET` | `/api/auth/verify?token=` | — | Verify email address |
| `POST` | `/api/auth/forgot-password` | — | Request password reset email |
| `POST` | `/api/auth/reset-password` | — | Reset password with token |

### Job Applications

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/jobs` | Client | List my job applications |
| `POST` | `/api/jobs` | Client | Create job application |
| `GET` | `/api/jobs/{id}` | Client | Get job application details |
| `PUT` | `/api/jobs/{id}` | Client | Update my job application |
| `DELETE` | `/api/jobs/{id}` | Client | Delete my job application |

### Admin

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/admin/login` | — | Admin login |
| `GET` | `/api/admin/jobs` | Admin | List all job applications |
| `PUT` | `/api/admin/jobs/{id}/status` | Admin | Approve or reject a job |

### Other

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |

## Seed Admin

```bash
go run ./cmd/server seed-admin --email=admin@site.com --password=yourpass
```

## Docker

```bash
# Build and run
docker compose up --build

# Seed admin
docker compose run app ./server seed-admin --email=admin@site.com --password=yourpass
```

Create a `.env` file:

```env
JWT_SECRET=your-64-char-hex-secret
APP_URL=https://yourdomain.com
CORS_ORIGIN=https://yourfrontend.com
SMTP_HOST=smtp.example.com
SMTP_USER=you@example.com
SMTP_PASS=yourpassword
SMTP_FROM=noreply@example.com
```
