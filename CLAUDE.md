# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Workflow Rules

- **Never `git push` unless explicitly asked.** Commit freely, but always wait for the user to say "push" before pushing to remote.
- Always include `Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>` in every commit.

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
go run -C backend ./cmd migrate up     # apply pending DB migrations
go run -C backend ./cmd serve          # run dev server (http://localhost:8080)
go run -C backend ./cmd aggregate-user-interests  # run interest aggregation job
go build -C backend ./...              # build all
go test -C backend ./...               # run all tests
go test -C backend ./internal/service/ # run a specific package's tests
```

## Architecture

### Backend

- **Entry**: `backend/cmd/main.go` — Cobra CLI with three subcommands: `migrate`, `serve`, `jobs`
- **`migrate`**: `backend/migrate/migrate.go` — loads config, connects DB, runs migrations standalone
- **`serve`**: `backend/serve/server.go` — DI wiring (repos → services → handlers → tracking → analytics), HTTP server lifecycle with graceful shutdown
- **`jobs`**: `backend/jobs/jobs.go` — incremental interest aggregation job; reads events since watermark, resolves device→user mapping, upserts to `user_interests`, populates Redis cache, updates watermark
- **Config**: `backend/internal/config/` — Viper-based; `application.go` defines struct + `loadDefault()`, `load.go` does the loading. Env vars use `__` as separator (e.g. `DB__HOST`, `JWT__ACCESS_SECRET`, `REDIS__ADDR`). Duration fields are strings (`"15m"`). `DB__MIGRATIONS_SOURCE` controls migration file path — auto-resolved from relative to absolute.
- **Handlers**: `backend/internal/handler/` — Gin handlers + `router.go`
- **Services**: `backend/internal/service/` — business logic, ownership enforcement
- **Repositories**: `backend/internal/repository/postgres/` — sqlx, raw SQL
- **Models**: `backend/internal/model/` — UUID v7 PKs everywhere
- **DTOs**: `backend/internal/dto/` — all request/response structs; never define inline in handlers
- **Utilities**: `backend/pkg/` — storage (S3/CloudFront), token (JWT), password, slug, pagination, urlcache
- **Migrations**: `backend/migrations/` — SQL files (000001–000015) run via `./cmd migrate up`
- **Tracking**: `backend/internal/tracking/` — independent event ingestion module; mounts `POST /api/track` (no `/v1`) directly on the engine from `serve/server.go`
- **Analytics**: `backend/internal/analytics/` — Redis-backed speed layer; mounts `GET /api/v1/analytics/trending|suggestions|similar`
- Go module: `github.com/yumikokawaii/sherry-archive`

### Key Backend Conventions

- UUID v7 for all IDs: `uuid.Must(uuid.NewV7())`
- Images stored as object keys in DB (not presigned URLs); URLs resolved at read time via `urlcache.URLCache`
- `urlcache.URLCache` takes a `Signer` interface — uses `CloudFrontSigner` when `CLOUDFRONT__DOMAIN` is set, otherwise falls back to S3 presign
- Pages uploaded via zip (replace semantics, filename-sorted order)
- `backend/config.yaml` is gitignored — copy from `config.example.yaml`
- Tracking goroutines must use `context.Background()`, never `c.Request.Context()` (request context is cancelled when handler returns)
- `tracking` and `analytics` are independent modules — both use `Mount(r *gin.Engine)` and are wired from `serve/server.go`, never through `handler/router.go`
- `tracking.Enricher` interface is defined in the tracking package and implemented by `analytics.Store` to avoid import cycles
- Interest profiles are stored in `user_interests` (Postgres, source of truth); Redis `interests:{identity_id}` is a 24h cache-aside populated by `./cmd jobs`
- `ProcessEvents` only updates the `seen:{device_id}` Redis set in real-time; interest scoring is done in the batch job
- Login/register accept optional `device_id` to populate `device_user_mappings` for identity resolution in the jobs

### Analytics Config (env vars)

- `ANALYTICS__CONTRIBUTION_CAP` — max trending points a device can contribute to one manga per 24h window (default: `15`)
- `ANALYTICS__DECAY_INTERVAL` — how often trending scores decay (default: `"1h"`)

### CloudFront Config (env vars)

- `CLOUDFRONT__DOMAIN` — CloudFront distribution domain (e.g. `d1234.cloudfront.net`); leave empty for local dev
- `CLOUDFRONT__KEY_PAIR_ID` — CloudFront key pair ID
- `CLOUDFRONT__PRIVATE_KEY` — RSA private key PEM string (SecureString in SSM)

### Frontend

- **Entry**: `frontend/src/main.tsx` → `App.tsx` (React Router root)
- **Pages**: `frontend/src/pages/` — one component per route
- **Components**: `frontend/src/components/` — shared UI
- **API client**: `frontend/src/lib/` — `api.ts` (fetch wrapper + `ApiError`), `manga.ts` (mangaApi + analyticsApi), `auth.ts`
- **Tracking SDK**: `frontend/src/lib/tracking/` — `device.ts` (getDeviceId), `tracker.ts` (typed event methods, POST /api/track), `index.ts`
- **Auth**: `frontend/src/contexts/AuthContext.tsx` — JWT stored in localStorage, session restored on mount
- **Types**: `frontend/src/types/`
- Vite dev server proxies `/api/*` → `http://localhost:8080` (`vite.config.ts`)
- Tailwind v4 `@theme` in `index.css` — custom tokens: `forest-*`, `jade-*`, `mint-*`
- Animations via Framer Motion (`motion.*`, `AnimatePresence`)

### Key Frontend Conventions

- `api.get<T>` in `api.ts` returns `json.data as T` — the wrapper already unwraps the `data` field. Always type responses as `T`, never `{ data: T }`, or `.data` will be called twice and return `undefined`.
- `TagBadge` is a link to `/?tags[]=<tag>` — clicking a tag filters the homepage catalog
- Suggestions endpoint requires `device_id` query param; pass `user_id` as well when user is logged in for personalized results

## Routes

### Backend API
- `POST /api/track` — event ingestion (no `/v1` prefix)
- `GET /api/v1/analytics/trending` — top trending manga
- `GET /api/v1/analytics/suggestions?device_id=&user_id=` — personalized recommendations
- `GET /api/v1/analytics/similar?manga_id=` — content-based similar manga
- `POST /api/v1/auth/register|login|refresh|logout`
- `GET /api/v1/auth/me`
- `GET/POST/PATCH/DELETE /api/v1/mangas`
- `GET/POST/PATCH/DELETE /api/v1/mangas/:id/chapters`
- `GET/POST/PATCH/DELETE /api/v1/mangas/:id/chapters/:id/pages`
- `GET/POST /api/v1/mangas/:id/comments`
- `GET/POST/DELETE /api/v1/users/me/bookmarks`

## Infrastructure

- EC2: Docker, Nginx → Go on `:8080`
- RDS: PostgreSQL, SSL required
- ElastiCache: Valkey, TLS enabled
- S3: `sherry-archive`, `ap-southeast-1`
- CI/CD: GitHub Actions → ECR (`latest`) → SSH deploy via `scripts/deploy.sh`
- Config: SSM Parameter Store at `/sherry-archive/`; `scripts/ssm-put.sh` to write, `scripts/ssm-export.sh` to load
- Jobs: run via ECS Fargate scheduled task (EventBridge cron), same ECR image, command override `./cmd jobs`
