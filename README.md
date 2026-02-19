# Sherry Archive

A manga reading platform. Go backend + React frontend.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go, Gin, sqlx, golang-migrate, golang-jwt, Cobra, Viper |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4, Framer Motion |
| Storage | PostgreSQL 16, MinIO (S3-compatible) |

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker (for infrastructure)

### 1. Start infrastructure

```bash
docker compose up -d
```

Starts PostgreSQL on `:5432` and MinIO on `:9000` (console at http://localhost:9001).

### 2. Configure backend

```bash
cp backend/config.example.yaml backend/config.yaml
# edit config.yaml as needed
```

Key env vars (override via `DB__HOST`, `JWT__ACCESS_SECRET`, etc. — `__` maps to `.`):

| Env var | Default | Description |
|---|---|---|
| `DB__HOST` | `localhost` | PostgreSQL host |
| `DB__PASSWORD` | `postgres` | PostgreSQL password |
| `DB__MIGRATIONS_SOURCE` | `file://migrations` | Migration source URL |
| `JWT__ACCESS_SECRET` | *(change me)* | JWT signing secret |
| `JWT__REFRESH_SECRET` | *(change me)* | Refresh token secret |
| `MINIO__ENDPOINT` | `localhost:9000` | MinIO endpoint |

### 3. Run migrations

```bash
go run -C backend ./cmd migrate
```

### 4. Start backend

```bash
go run -C backend ./cmd serve
# API available at http://localhost:8080/api/v1
```

### 5. Start frontend

```bash
npm --prefix frontend install
npm --prefix frontend run dev
# UI available at http://localhost:5173
```

The Vite dev server proxies `/api/*` → `http://localhost:8080`.

## Project Structure

```
sherry-archive/
├── backend/
│   ├── cmd/main.go          # CLI entry point (Cobra)
│   ├── serve/server.go      # DI wiring + HTTP server lifecycle
│   ├── migrate/migrate.go   # Standalone migration runner
│   ├── internal/
│   │   ├── config/          # Viper config (application.go + load.go)
│   │   ├── model/           # Domain models (UUID v7 PKs)
│   │   ├── dto/             # Request/response structs
│   │   ├── repository/      # Data access (sqlx, raw SQL)
│   │   ├── service/         # Business logic
│   │   ├── handler/         # Gin HTTP handlers + router
│   │   └── middleware/      # JWT auth middleware
│   ├── pkg/                 # storage, token, password, slug, pagination
│   ├── migrations/          # SQL migration files
│   └── config.example.yaml
└── frontend/
    └── src/
        ├── pages/           # HomePage, MangaDetailPage, ReaderPage, LoginPage
        ├── components/      # Navbar, MangaCard, Spinner, etc.
        ├── contexts/        # AuthContext
        ├── lib/             # API client (api.ts, manga.ts, auth.ts)
        └── types/           # TypeScript interfaces
```

## API

Base path: `/api/v1`

| Resource | Endpoints |
|---|---|
| Auth | `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`; `GET /auth/me` |
| Manga | `GET/POST /mangas`; `GET/PATCH/DELETE /mangas/:id`; `PUT /mangas/:id/cover` |
| Chapters | `GET/POST /mangas/:id/chapters`; `GET/PATCH/DELETE /mangas/:id/chapters/:id` |
| Pages | `POST .../pages` (multipart), `POST .../pages/zip` (zip upload, replaces all) |
| Bookmarks | `GET/PUT/DELETE /users/me/bookmarks/:mangaID` |
| Users | `GET /users/:id`; `PATCH /users/me`; `PUT /users/me/avatar` |

## Other Commands

```bash
# Backend
go build -C backend ./...              # build
go test -C backend ./...               # test all
go test -C backend ./internal/service/ # test a package

# Frontend
npm --prefix frontend run build        # production build
npm --prefix frontend run lint         # ESLint
```
