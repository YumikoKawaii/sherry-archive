# Sherry Archive — Technical Design

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Infrastructure](#2-infrastructure)
3. [Database Schema](#3-database-schema)
4. [Backend Architecture](#4-backend-architecture)
5. [Analytics & Recommendation System](#5-analytics--recommendation-system)
6. [Frontend Architecture](#6-frontend-architecture)
7. [Observability & Metrics](#7-observability--metrics)
8. [Configuration Reference](#8-configuration-reference)
9. [Data Flow Diagrams](#9-data-flow-diagrams)

---

## 1. System Overview

Sherry Archive is a manga reading platform. Users can browse, read, and bookmark manga. Owners upload manga and chapters (via zip). The system tracks reading behaviour to power trending, personalised recommendations, and similar manga.

**Tech stack:**

| Layer | Technology |
|---|---|
| Backend | Go 1.25, Gin, sqlx |
| Database | PostgreSQL (RDS) |
| Cache / Real-time | Redis / Valkey (ElastiCache) |
| Object storage | AWS S3 (CDN-ready via CloudFront) |
| Queue | AWS SQS (zip processing) |
| Frontend | React 19, Vite 7, TypeScript, Tailwind v4, Framer Motion |
| Hosting | EC2 (Docker), Nginx reverse proxy |
| CI/CD | GitHub Actions → ECR → SSH deploy |
| ETL job | ECS scheduled task |

---

## 2. Infrastructure

```
Internet
   │
   ▼
Nginx (EC2 :443)
   │
   ▼
Go server (:8080)
   ├── PostgreSQL RDS (ap-southeast-1)
   ├── Valkey ElastiCache (TLS, :6379)
   ├── S3 bucket: sherry-archive
   └── SQS queue (zip upload tasks)

ECS scheduled task
   ├── PostgreSQL RDS
   └── Valkey ElastiCache
```

- Go binary serves **both** the API and the React SPA static files.
- Nginx terminates TLS and proxies everything to `:8080`.
- Config is loaded from AWS SSM Parameter Store at deploy time (`/sherry-archive/*`).

### Deploy flow

1. GitHub Actions runs tests
2. Builds Docker image, pushes to ECR (`:latest` only; untagged images pruned)
3. SSH into EC2, pull latest image, reload SSM params, restart container

---

## 3. Database Schema

All primary keys are **UUID v7** (time-ordered, generated in application code via `uuid.Must(uuid.NewV7())`).

### Core tables

```
users
  id            UUID PK
  username      TEXT UNIQUE
  email         TEXT UNIQUE
  password_hash TEXT
  avatar_url    TEXT
  bio           TEXT
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ

mangas
  id          UUID PK
  owner_id    UUID → users.id
  title       TEXT
  slug        TEXT UNIQUE
  description TEXT
  cover_key   TEXT          ← S3 object key (not a URL)
  status      ENUM(ongoing, completed, hiatus)
  type        ENUM(series, oneshot)
  tags        TEXT[]        ← GIN indexed
  author      TEXT          ← B-tree indexed
  artist      TEXT
  category    TEXT          ← B-tree indexed
  created_at  TIMESTAMPTZ
  updated_at  TIMESTAMPTZ

chapters
  id         UUID PK
  manga_id   UUID → mangas.id
  number     FLOAT
  title      TEXT
  page_count INT
  created_at TIMESTAMPTZ
  updated_at TIMESTAMPTZ
  UNIQUE (manga_id, number)

pages
  id         UUID PK
  chapter_id UUID → chapters.id
  number     INT
  key        TEXT    ← S3 object key
  created_at TIMESTAMPTZ

bookmarks
  user_id          UUID → users.id
  manga_id         UUID → mangas.id
  chapter_id       UUID → chapters.id
  last_page_number INT
  updated_at       TIMESTAMPTZ
  PRIMARY KEY (user_id, manga_id)

refresh_tokens
  id         UUID PK
  user_id    UUID → users.id
  token_hash TEXT UNIQUE
  expires_at TIMESTAMPTZ
  created_at TIMESTAMPTZ

comments
  id         UUID PK
  manga_id   UUID → mangas.id
  chapter_id UUID → chapters.id (nullable — null = manga-level comment)
  user_id    UUID → users.id
  content    TEXT
  created_at TIMESTAMPTZ
  updated_at TIMESTAMPTZ

upload_tasks
  id         UUID PK
  manga_id   UUID → mangas.id
  chapter_id UUID → chapters.id (nullable — set on completion)
  status     ENUM(pending, processing, done, failed)
  error_msg  TEXT
  created_at TIMESTAMPTZ
  updated_at TIMESTAMPTZ
```

### Analytics tables

```
events                              ← raw tracking events (append-only)
  device_id  UUID
  user_id    UUID (nullable)
  event      TEXT
  properties JSONB
  referrer   TEXT
  ip_hash    TEXT
  user_agent TEXT
  created_at TIMESTAMPTZ
  INDEX BRIN(created_at), (device_id, created_at DESC), (user_id, created_at DESC)

device_user_mappings                ← links anonymous device to user on login
  device_id  UUID
  user_id    UUID → users.id
  created_at TIMESTAMPTZ
  PRIMARY KEY (device_id, user_id)

user_interests                      ← personalised interest profile (ETL output)
  identity_id UUID                  ← user_id if logged in, device_id if anonymous
  dimension   TEXT                  ← "tag:action", "author:Oda", "category:manga"
  score       FLOAT                 ← decayed cumulative score (0.98 decay per event)
  updated_at  TIMESTAMPTZ
  PRIMARY KEY (identity_id, dimension)

interest_sync_watermarks            ← ETL progress marker per identity
  identity_id   UUID PK
  last_synced_at TIMESTAMPTZ

manga_popularity                    ← lifetime popularity score per manga (ETL output)
  manga_id   UUID PK → mangas.id
  score      FLOAT                  ← cumulative, never decays
  updated_at TIMESTAMPTZ

seen_manga                          ← permanent seen history for suggestion exclusion
  identity_id UUID
  manga_id    UUID → mangas.id
  seen_at     TIMESTAMPTZ
  PRIMARY KEY (identity_id, manga_id)
  INDEX (identity_id)
```

### Image storage convention

Images (covers, pages) are stored as **S3 object keys** in the DB, never as URLs. Presigned URLs (or CloudFront signed URLs) are resolved at read time in handlers via `urlcache.URLCache`. The cache holds resolved URLs in Redis for the presign expiry duration.

---

## 4. Backend Architecture

### Entry points (`backend/cmd/`)

| Command | Description |
|---|---|
| `./cmd serve` | HTTP server |
| `./cmd migrate` | Run pending DB migrations |
| `./cmd aggregate-user-interests` | ETL job (run via ECS scheduled task) |

### Package layout

```
backend/
├── cmd/              Entry point + Cobra subcommands
├── serve/            DI wiring + HTTP server lifecycle
├── migrate/          Standalone migration runner
├── jobs/             ETL job implementation
├── internal/
│   ├── config/       Viper config (Application struct)
│   ├── model/        Domain structs (UUID v7 PKs)
│   ├── dto/          All request/response structs
│   ├── repository/   Interfaces + postgres implementations
│   ├── service/      Business logic + ownership enforcement
│   ├── handler/      Gin handlers + router
│   ├── tracking/     Event ingestion module
│   ├── analytics/    Redis speed layer (trending + suggestions)
│   └── apperror/     Sentinel errors
└── pkg/
    ├── storage/      S3 client + CloudFront signer
    ├── urlcache/     Presigned URL cache (Redis)
    ├── token/        JWT (access + refresh)
    ├── password/     bcrypt helpers
    ├── slug/         URL slug generation
    ├── pagination/   Cursor/offset helpers
    └── queue/        SQS client
```

### Request lifecycle (API)

```
HTTP Request
  → Gin router (handler/router.go)
  → Auth middleware (JWT validation, optional)
  → Handler (handler/*.go)          ← validates input, calls service
  → Service (service/*.go)          ← business logic, ownership check
  → Repository (repository/postgres/*.go)  ← raw SQL via sqlx
  → Response: { "data": ... } or { "error": "..." }
```

### API routes

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
GET    /api/v1/auth/me

GET    /api/v1/mangas
POST   /api/v1/mangas
GET    /api/v1/mangas/:id
PATCH  /api/v1/mangas/:id
DELETE /api/v1/mangas/:id
PUT    /api/v1/mangas/:id/cover

GET    /api/v1/mangas/:id/chapters
POST   /api/v1/mangas/:id/chapters
GET    /api/v1/mangas/:id/chapters/:chId
PATCH  /api/v1/mangas/:id/chapters/:chId
DELETE /api/v1/mangas/:id/chapters/:chId
POST   /api/v1/mangas/:id/chapters/:chId/pages/zip
POST   /api/v1/mangas/:id/oneshot/upload

GET    /api/v1/mangas/:id/comments
POST   /api/v1/mangas/:id/comments
GET    /api/v1/mangas/:id/chapters/:chId/comments
POST   /api/v1/mangas/:id/chapters/:chId/comments
PATCH  /api/v1/mangas/:id/comments/:cmId
DELETE /api/v1/mangas/:id/comments/:cmId

GET    /api/v1/users/:id/mangas
PUT    /api/v1/users/me/bookmarks/:mangaId
GET    /api/v1/users/me/bookmarks/:mangaId
GET    /api/v1/users/me/bookmarks
DELETE /api/v1/users/me/bookmarks/:mangaId

GET    /api/v1/analytics/trending
GET    /api/v1/analytics/suggestions
GET    /api/v1/analytics/similar

POST   /api/track          ← no /v1 prefix, no auth required
```

### Zip upload flow

1. Client POSTs zip to `/pages/zip` → server queues an `upload_task` in DB and sends task ID to SQS
2. SQS worker (same process, background goroutine) receives task, downloads zip from S3, extracts pages, uploads each page to S3, updates chapter
3. `ClaimProcessing` uses `UPDATE ... WHERE status='pending' RETURNING id` — atomic, prevents duplicate processing on redelivery

---

## 5. Analytics & Recommendation System

This is the most complex part of the system. It has three independent components that work together.

### 5.1 Event tracking (real-time)

`POST /api/track` accepts batches of up to 50 events. No authentication required — `user_id` is extracted from the Bearer token if present.

**Event types and point values:**

| Event | Trending pts | Interest pts | Triggers seen |
|---|---|---|---|
| `manga_view` | 1 | 1 | yes |
| `chapter_open` | 3 | 3 | yes |
| `chapter_complete` | 5 | 5 | yes |
| `comment_post` | — | 4 | yes |
| `bookmark_add` | — | 8 | yes |
| `bookmark_remove` | — | -3 | yes |

After storing events to DB, `analytics.Store.ProcessEvents` is called in a **fire-and-forget goroutine** (uses `context.Background()`, never the request context):

1. **Trending**: increments Redis sorted set `trending:{manga_id}`, capped per device per manga per 24h window via atomic Lua script. Cap default: 15 pts (configurable via `ANALYTICS__CONTRIBUTION_CAP`).
2. **Seen**: inserts `seen_manga(identity_id, manga_id)` — uses `user_id` if the user is logged in, otherwise `device_id`.

### 5.2 ETL job (scheduled, runs on ECS)

Command: `./cmd aggregate-user-interests`

Runs incrementally — only processes events since the last watermark per device.

**Algorithm per device:**

```
for each device with unprocessed events:
  resolve identity_id = user_id if mapped, else device_id
  fetch events since watermark
  fetch manga metadata in batch (tags, author, category)

  for each event:
    popularity delta += capped(trending_pts, 15/device/manga/day)
    interest scores:
      tags   → pts / len(active_tags)  [split evenly, stop tags excluded]
      author → full pts
      category → full pts
    apply 0.98 decay to existing scores before adding

  upsert user_interests to DB
  upsert manga_popularity to DB (incremental add, never decays)
  write interests:{identity_id} to Redis (24h TTL)
  advance watermark
```

**Stop tags**: configured via `ANALYTICS__STOP_TAGS` (comma-separated). Tags like "oneshot" that describe format rather than genre are excluded from interest dimensions so they don't pollute recommendations.

**Popularity vs Trending**:
- **Trending** (Redis ZSET): real-time, decays by 0.9 every `ANALYTICS__DECAY_INTERVAL` (default 1h). Reflects what's hot *right now*.
- **Popularity** (`manga_popularity.score`): lifetime cumulative score, never decays. Reflects how well-received a manga is overall. Used to rank suggestion candidates.

### 5.3 Suggestion query (3 phases)

`GET /api/v1/analytics/suggestions?device_id=&user_id=&manga_id=&limit=`

**Phase 0 — Load interest profile**
- Try Redis `interests:{identity_id}` (24h TTL, cache-aside)
- On miss: query `user_interests` DB, repopulate Redis
- If user profile empty → fallback to device profile
- Filter stop tags
- Sort by score → pick top 5 tags, 3 authors, 3 categories

**Context boost** (if `manga_id` provided):
- Load current manga's metadata
- Extend tag pool to 8, author/category pool to 5

**Cold-start** (no interest profile at all):
- If `manga_id` context → run similarity query on current manga's metadata
- Otherwise → top popularity-ranked unseen manga

**Phase 1 — Seen exclusion**
```sql
SELECT manga_id FROM seen_manga WHERE identity_id = $identity_id
```

**Phase 2 — Candidate retrieval (UNION of separate index scans)**
```sql
SELECT id FROM mangas WHERE id != ALL($seen) AND tags && $tags
UNION
SELECT id FROM mangas WHERE id != ALL($seen) AND author != '' AND author = ANY($authors)
UNION
SELECT id FROM mangas WHERE id != ALL($seen) AND category != '' AND category = ANY($categories)
```
Each branch gets its own index: GIN on `tags`, B-tree on `author`, B-tree on `category`.
`OR` in a single query would prevent the planner from using all three indexes.
UNION deduplicates candidate IDs.

**Phase 3 — Ranking**
```sql
SELECT m.* FROM mangas m
LEFT JOIN manga_popularity p ON p.manga_id = m.id
WHERE m.id = ANY($candidate_ids)
ORDER BY COALESCE(p.score, 0) DESC
LIMIT $n
```

### 5.4 Identity model

```
Anonymous user:   identity_id = device_id (UUID stored in localStorage)
Logged-in user:   identity_id = user_id

On login/register with device_id:
  1. Upsert device_user_mappings(device_id → user_id)
  2. COPY seen_manga FROM device → user  (ON CONFLICT DO NOTHING)
  3. MERGE user_interests FROM device → user  (GREATEST score on conflict)
  4. DELETE Redis interests:{user_id}  (force reload of merged profile)

While logged in:
  Tracking events write seen_manga under user_id directly.
```

### 5.5 Redis key inventory

| Key pattern | Type | TTL | Purpose |
|---|---|---|---|
| `trending` | ZSET | no TTL (decayed) | Real-time trending scores |
| `contributed:{device_id}:{manga_id}` | STRING | 24h | Per-device contribution cap |
| `interests:{identity_id}` | HASH | 24h | Interest profile cache |
| `manga:meta:{manga_id}` | HASH | 1h | Manga metadata cache (tags/author/category) |
| `urlcache:{object_key}` | STRING | presign expiry | Presigned URL cache |

### 5.6 Similar manga

`GET /api/v1/analytics/similar?manga_id=&limit=`

Same query as suggestions phase 2+3 but using the source manga's own metadata. No seen exclusion, no interest profile. Only excludes the source manga itself. Used in "More Like This" on the detail page.

---

## 6. Frontend Architecture

### Structure

```
frontend/src/
├── main.tsx          Entry point
├── App.tsx           React Router root
├── pages/            One component per route
│   ├── HomePage.tsx
│   ├── MangaDetailPage.tsx
│   ├── ReaderPage.tsx
│   ├── BookmarksPage.tsx
│   ├── LoginPage.tsx
│   ├── CreateMangaPage.tsx
│   └── ManageChaptersPage.tsx
├── components/       Shared UI
│   ├── Layout.tsx
│   ├── Navbar.tsx
│   ├── MangaCard.tsx
│   ├── CommentSection.tsx
│   ├── TagBadge.tsx    ← links to /?tags[]=<tag>
│   ├── StatusBadge.tsx
│   └── Spinner.tsx
├── contexts/
│   └── AuthContext.tsx  ← JWT in localStorage, session restored on mount
├── lib/
│   ├── api.ts          ← fetch wrapper; unwraps .data field automatically
│   ├── manga.ts        ← mangaApi, bookmarkApi, analyticsApi
│   ├── auth.ts         ← authApi (login/register pass device_id)
│   └── tracking/
│       ├── device.ts   ← getDeviceId() with crypto.randomUUID fallback
│       ├── tracker.ts  ← typed event methods, POST /api/track
│       └── index.ts
└── types/
    ├── manga.ts
    ├── user.ts
    └── api.ts
```

### API client convention

`api.get<T>` in `api.ts` already unwraps the `{ "data": ... }` envelope. Always type responses as `T`, never `{ data: T }` — doing otherwise causes `.data` to be called twice and returns `undefined`.

### Analytics integration per page

**HomePage**
- On mount: fetch `trending` (once)
- On `user.id` change (login/logout): re-fetch `suggestions(deviceId, { userId })`
- Displays: "Trending" shelf (with score badge), "For You" shelf

**MangaDetailPage**
- On manga load: fetch `similar(mangaId)` + `suggestions(deviceId, { userId, mangaId })`
- Fires `tracker.mangaView` on load
- Displays: "More Like This" shelf, "For You" shelf (context-boosted)

**ReaderPage**
- Fires `tracker.chapterOpen` on load
- Fires `tracker.chapterComplete` when user reaches last page

### Auth flow

```
Login/Register:
  authApi.login(email, password, getDeviceId())
  → stores access_token + refresh_token in localStorage
  → sets user in AuthContext
  → backend merges device data into user profile

Session restore:
  On mount: if access_token in localStorage → authApi.me()
  On 401: token refresh via authApi.refresh()

Logout:
  Revokes refresh token, clears localStorage
```

---

## 7. Observability & Metrics

### Overview

Metrics are pushed directly to **AWS CloudWatch** using `aws-sdk-go-v2/service/cloudwatch`. The app accumulates counters and duration statistics in memory and calls `PutMetricData` every **10 seconds** — no scraper, no sidecar, no exposed HTTP endpoint required.

All metrics land in the **`SherryArchive`** CloudWatch namespace. Build a CloudWatch Dashboard from there (`Metrics → Custom namespaces → SherryArchive`).

At startup, `metrics.Init` loads AWS credentials from the EC2 instance role via IMDS. If CloudWatch is unavailable (e.g. local dev), it logs and all `Record*` calls become no-ops — the app continues normally.

### IAM requirement

The EC2 instance role must have the following permission:

```json
{
  "Effect": "Allow",
  "Action": "cloudwatch:PutMetricData",
  "Resource": "*",
  "Condition": {
    "StringEquals": { "cloudwatch:namespace": "SherryArchive" }
  }
}
```

### Metrics reference

#### HTTP layer

| Metric name | Dimensions | Unit | Description |
|---|---|---|---|
| `HTTPRequestCount` | `Method`, `Route`, `StatusCode` | Count | Request count per route template (e.g. `/api/v1/mangas/:mangaID`), not raw URL — keeps dimension cardinality low |
| `HTTPRequestDuration` | `Method`, `Route` | Seconds | StatisticSet (min/max/sum/count) of request latency per route |

#### Database connection pool

Read from `db.Stats()` at each flush interval.

| Metric name | Unit | Description |
|---|---|---|
| `DBOpenConnections` | Count | Total open connections (in-use + idle) |
| `DBInUseConnections` | Count | Connections currently executing a query |
| `DBIdleConnections` | Count | Connections sitting idle in the pool |

#### Business metrics

| Metric name | Dimensions | Unit | Description |
|---|---|---|---|
| `TrackingEvents` | `EventType` | Count | Events ingested via `POST /api/track` (`manga_view`, `chapter_open`, `chapter_complete`, …) |
| `AnalyticsRequests` | `Endpoint` | Count | Requests to analytics endpoints (`trending`, `suggestions`, `similar`) |

### Implementation notes

- `backend/internal/metrics/` — `metrics.go` (publisher + flush logic), `middleware.go` (Gin middleware)
- HTTP middleware uses `c.FullPath()` (Gin route template) for the `Route` dimension — prevents high-cardinality explosion from path parameters.
- `metrics.Init(ctx, region, namespace, db)` is called from `serve/server.go` after the background context is created, so the flush goroutine is cancelled on graceful shutdown.
- The flush goroutine snapshots and resets accumulators atomically under a mutex before each `PutMetricData` call to avoid data races and double-counting.
- `PutMetricData` batches up to 1000 data points per API call (CloudWatch limit).
- Each data point is stamped with `windowStart` (the time accumulation began, not the flush time) so CloudWatch places the counts in the correct 10s bucket — enabling accurate `Sum / 10 = req/s` calculations via CloudWatch Metric Math.
- All metrics use `StorageResolution: 1` (high-resolution), which allows dashboard granularity down to 10s instead of the default 1-minute minimum.
- Flush interval is 10s to limit data loss on deploys/crashes to at most one 10s window.

---

## 8. Configuration Reference

Config is loaded from `config.yaml` (searched in `.` then `..`) or environment variables using `__` as separator.

| Env var | Default | Description |
|---|---|---|
| `DB__HOST` | — | PostgreSQL host |
| `DB__PORT` | 5432 | PostgreSQL port |
| `DB__USER` | — | DB user |
| `DB__PASSWORD` | — | DB password |
| `DB__NAME` | — | DB name |
| `DB__SSL_MODE` | require | SSL mode |
| `JWT__ACCESS_SECRET` | — | Access token signing key |
| `JWT__REFRESH_SECRET` | — | Refresh token signing key |
| `JWT__ACCESS_TOKEN_EXPIRY` | 15m | Access token TTL |
| `JWT__REFRESH_TOKEN_EXPIRY` | 168h | Refresh token TTL |
| `REDIS__ADDR` | — | Redis/Valkey address |
| `REDIS__PASSWORD` | — | Redis password |
| `REDIS__TLS` | true | Enable TLS (set false for local dev) |
| `S3__BUCKET` | — | S3 bucket name |
| `S3__REGION` | ap-southeast-1 | AWS region |
| `S3__ENDPOINT` | — | Custom endpoint (local MinIO only) |
| `S3__PRESIGN_EXPIRY` | 1h | Presigned URL TTL |
| `CLOUDFRONT__DOMAIN` | — | CloudFront domain (empty = use S3 presign) |
| `CLOUDFRONT__KEY_PAIR_ID` | — | CloudFront key pair ID |
| `CLOUDFRONT__PRIVATE_KEY` | — | RSA private key PEM |
| `SQS__QUEUE_URL` | — | SQS queue URL for upload tasks |
| `ANALYTICS__CONTRIBUTION_CAP` | 15 | Max trending pts per device per manga per 24h |
| `ANALYTICS__DECAY_INTERVAL` | 1h | How often trending scores decay |
| `ANALYTICS__STOP_TAGS` | oneshot | Comma-separated tags excluded from interest dims |
| `SERVER__PORT` | 8080 | HTTP listen port |

---

## 9. Data Flow Diagrams

### Reading a chapter (tracking flow)

```
Browser
  │  POST /api/track  { event: "chapter_open", manga_id: "...", device_id: "..." }
  ▼
tracking.Handler
  │  extract user_id from Bearer token (optional)
  │  goroutine: Insert(events) → events table
  │  goroutine: analytics.Store.ProcessEvents(events)
  │               ├─ ZADD trending (capped via Lua)
  │               └─ INSERT seen_manga(user_id or device_id, manga_id)
  ▼
204 No Content  (response immediate, goroutine continues)
```

### ETL job (interest aggregation)

```
ECS Scheduled Task (every N minutes)
  │
  ├─ SELECT distinct device_ids with events newer than watermark
  │
  │  for each device:
  ├─ resolve identity_id (user or device)
  ├─ SELECT events since watermark
  ├─ SELECT manga metadata (tags/author/category) in batch
  ├─ compute interest deltas (0.98 decay, stop tag filter)
  ├─ compute popularity deltas (capped per device/manga/day)
  ├─ UPSERT user_interests
  ├─ UPSERT manga_popularity (score += delta)
  ├─ SET interests:{identity_id} in Redis (24h TTL)
  └─ UPSERT interest_sync_watermarks
```

### Suggestion request

```
GET /api/v1/analytics/suggestions?device_id=X&user_id=Y&manga_id=Z
  │
  ├─ HGETALL interests:{user_id}     ← Redis cache hit?
  │    └─ miss: SELECT user_interests WHERE identity_id = user_id
  │              → repopulate Redis
  │    └─ empty: fallback to interests:{device_id}
  │    └─ still empty: cold-start path
  │
  ├─ filter stop tags, sort by score, pick top 5 tags / 3 authors / 3 categories
  ├─ if manga_id context: extend pool with context manga's metadata
  │
  ├─ SELECT manga_id FROM seen_manga WHERE identity_id = user_id  ← Phase 1
  │
  └─ SELECT m.* FROM mangas m                                     ← Phase 2+3
       LEFT JOIN manga_popularity p ON p.manga_id = m.id
       WHERE m.id != ALL($seen)
         AND (tags && $tags OR author = ANY($authors) OR ...)
       ORDER BY COALESCE(p.score, 0) DESC
       LIMIT N
```

### Login / device merge

```
POST /api/v1/auth/login  { email, password, device_id }
  │
  ├─ verify credentials
  ├─ issue token pair
  ├─ UPSERT device_user_mappings(device_id → user_id)
  ├─ INSERT seen_manga SELECT FROM seen_manga WHERE identity_id = device_id
  │    → ON CONFLICT (user_id, manga_id) DO NOTHING
  ├─ INSERT user_interests SELECT FROM user_interests WHERE identity_id = device_id
  │    → ON CONFLICT DO UPDATE SET score = GREATEST(...)
  └─ DEL interests:{user_id}  ← invalidate stale cache
```
