# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Sherry Archive** — a manga reading platform with a Go backend and a React frontend.

## Commands

### Frontend (`frontend/`)

```bash
npm --prefix frontend install          # install dependencies
npm --prefix frontend run dev          # dev server (http://localhost:5173)
npm --prefix frontend run build        # production build
npm --prefix frontend run lint         # ESLint
```

### Backend (`backend/`)

```bash
go run -C backend ./cmd migrate        # apply pending DB migrations
go run -C backend ./cmd serve          # run dev server (http://localhost:8080)
go build -C backend ./...              # build all
go test -C backend ./...               # run all tests
go test -C backend ./internal/service/ # run a specific package's tests
```

## Architecture

### Backend

- **Entry**: `backend/cmd/main.go` — Cobra CLI with two subcommands: `migrate` and `serve`
- **`migrate`**: `backend/migrate/migrate.go` — loads config, connects DB, runs migrations standalone
- **`serve`**: `backend/serve/server.go` — DI wiring (repos → services → handlers), HTTP server lifecycle with graceful shutdown
- **Config**: `backend/internal/config/` — Viper-based; `application.go` defines struct + `loadDefault()`, `load.go` does the loading. Env vars use `__` as separator (e.g. `DB__HOST`, `JWT__ACCESS_SECRET`). Duration fields are strings (`"15m"`). `DB__MIGRATIONS_SOURCE` controls migration file path — auto-resolved from relative to absolute.
- **Handlers**: `backend/internal/handler/` — Gin handlers + `router.go`
- **Services**: `backend/internal/service/` — business logic, ownership enforcement
- **Repositories**: `backend/internal/repository/postgres/` — sqlx, raw SQL
- **Models**: `backend/internal/model/` — UUID v7 PKs everywhere
- **DTOs**: `backend/internal/dto/` — all request/response structs; never define inline in handlers
- **Utilities**: `backend/pkg/` — storage (MinIO), token (JWT), password, slug, pagination
- **Migrations**: `backend/migrations/` — SQL files run via `./cmd migrate`
- Go module: `github.com/yumikokawaii/sherry-archive`

### Key Backend Conventions

- UUID v7 for all IDs: `uuid.Must(uuid.NewV7())`
- Images stored as object keys in DB (not presigned URLs); URLs resolved at read time in handlers
- Pages uploaded via zip (replace semantics, filename-sorted order)
- `backend/config.yaml` is gitignored — copy from `config.example.yaml`

### Frontend

- **Entry**: `frontend/src/main.tsx` → `App.tsx` (React Router root)
- **Pages**: `frontend/src/pages/` — one component per route
- **Components**: `frontend/src/components/` — shared UI
- **API client**: `frontend/src/lib/` — `api.ts` (fetch wrapper + `ApiError`), `manga.ts`, `auth.ts`
- **Auth**: `frontend/src/contexts/AuthContext.tsx` — JWT stored in localStorage, session restored on mount
- **Types**: `frontend/src/types/`
- Vite dev server proxies `/api/*` → `http://localhost:8080` (`vite.config.ts`)
- Tailwind v4 `@theme` in `index.css` — custom tokens: `forest-*`, `jade-*`, `mint-*`
- Animations via Framer Motion (`motion.*`, `AnimatePresence`)
