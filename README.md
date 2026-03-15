<div align="center">

```
  ✦ · · · · · · · · · · · · · · · · · · · · ✦
        S H E R R Y   A R C H I V E
  ✦ · · · · · · · · · · · · · · · · · · · · ✦
```

### *A quiet place to keep the things you love*

---

[![Author](https://img.shields.io/badge/Author-Yumiko%20Kawaii-ff69b4?style=for-the-badge)](https://github.com/yumikokawaii)
[![Live](https://img.shields.io/badge/Live-sherry--archive.com-brightgreen?style=for-the-badge)](https://sherry-archive.com)
[![Stack](https://img.shields.io/badge/Stack-Go%20·%20React-00ADD8?style=for-the-badge)]()

</div>

---

## About

> *"An archive is not just storage — it is care made permanent."*

**Sherry Archive** is a personal manga reading platform built to own your library without depending on anyone else.

It started from a simple frustration: the readers you love go down, drown in ads, or vanish overnight. This is the answer to that — a place that's yours, that stays up, that gets out of the way while you read.

---

## Features

### The Library
Browse your collection with cover art, tags, author, status, and category filters. Everything is organized the way you want it, not the way an algorithm decides.

### The Reader
Smooth vertical scroll, lazy-loaded pages, a UI that fades away the moment you start reading. Designed to feel like turning a page, not using software.

### Chapters & Uploads
Drop a zip file and the chapter appears — pages ordered automatically by filename, stored on S3, served fast. Supports both multi-chapter series and oneshot manga.

### Bookmarks & Progress
Track where you left off on every manga. Progress is updated as you read, chapter by chapter.

### Recommendations
A real-time analytics engine tracks engagement (views, reads, bookmarks) to surface trending titles globally and personalized suggestions per device or user. Interest profiles are built incrementally via a background aggregation job and stored in Postgres, with Redis as a cache layer.

---

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, Gin, sqlx / PostgreSQL |
| Frontend | React 19, TypeScript, Vite 7, Tailwind v4, Framer Motion |
| Storage | AWS S3 (images), AWS SQS (async zip processing) |
| Cache | AWS ElastiCache (Valkey/Redis) |
| Infra | EC2, RDS, Docker, Nginx, GitHub Actions CI/CD |

---

## Local Development

### Prerequisites
- Go 1.25+
- Node.js 20+
- PostgreSQL
- Redis
- AWS credentials (or local MinIO for S3)

### Setup

```bash
# 1. Copy and fill in config
cp backend/config.example.yaml backend/config.yaml

# 2. Run migrations
go run -C backend ./cmd migrate up

# 3. Start backend
go run -C backend ./cmd serve

# 4. Start frontend (separate terminal)
npm --prefix frontend install
npm --prefix frontend run dev
```

Backend runs at `http://localhost:8080`, frontend at `http://localhost:5173`.

### Commands

```bash
# Backend
go run -C backend ./cmd serve          # HTTP server
go run -C backend ./cmd migrate up     # apply DB migrations
go run -C backend ./cmd jobs           # run interest aggregation job
go build -C backend ./...              # build all
go test -C backend ./...               # run all tests

# Frontend
npm --prefix frontend run dev          # dev server
npm --prefix frontend run build        # production build
npm --prefix frontend run lint         # ESLint
```

---

## Architecture Overview

```
frontend/                  React SPA (Vite)
backend/
  cmd/                     CLI entrypoint (serve | migrate | jobs)
  serve/                   DI wiring + HTTP server lifecycle
  migrate/                 Standalone migration runner
  jobs/                    Background job: interest aggregation
  internal/
    handler/               Gin HTTP handlers
    service/               Business logic
    repository/postgres/   sqlx raw SQL repositories
    model/                 Domain entities (UUID v7 PKs)
    dto/                   Request/response structs
    tracking/              Event ingestion → POST /api/track
    analytics/             Trending + suggestions → GET /api/v1/analytics/*
    config/                Viper config (env vars with __ separator)
  pkg/
    storage/               S3 / CloudFront signed URL client
    token/                 JWT manager
    urlcache/              Redis-backed presigned URL cache
    pagination/            Cursor/offset helpers
  migrations/              SQL migration files (000001–000015)
```

### Key Design Decisions

- **UUID v7** for all primary keys — time-ordered, no external sequence
- **Object keys in DB, not URLs** — S3/CloudFront URLs resolved at read time via `URLCache`
- **Tracking is fire-and-forget** — `POST /api/track` returns 204 immediately; enrichment runs in a `context.Background()` goroutine
- **Analytics is independent** — `tracking` and `analytics` mount separately from `serve/server.go`, never via `handler/router.go`
- **Interest profiles in Postgres** — source of truth; Redis is cache-aside with 24h TTL; `./cmd jobs` aggregates incrementally from the `events` table
- **CloudFront-ready** — set `CLOUDFRONT__DOMAIN` to switch from S3 presigned URLs to CloudFront signed URLs; falls back to S3 if unset

---

## Infrastructure

| Resource | Detail |
|----------|--------|
| EC2 | `18.141.43.132` — app server, Docker, Nginx |
| RDS | PostgreSQL, SSL required |
| ElastiCache | Valkey (Redis-compatible), TLS |
| S3 | `sherry-archive`, `ap-southeast-1` |
| ECR | Single image, `latest` tag |
| CI/CD | GitHub Actions → ECR → SSH deploy |
| Config | SSM Parameter Store at `/sherry-archive/` |

---

## Author

<div align="center">

**~ Yumiko Kawaii ~**

*Software Engineer*

*すべての物語に、居場所があってほしい*

</div>

---

<div align="center">

*~ built with matcha and late nights ~*

</div>
