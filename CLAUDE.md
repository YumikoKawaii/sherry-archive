# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Sherry Archive** — a manga reading platform with a Go backend and a React frontend.

## Monorepo Structure

```
sherry-archive/
├── frontend/   React + Vite + TypeScript + Framer Motion + Tailwind CSS
└── backend/    Go + Gin
```

## Commands

### Frontend (`frontend/`)

```bash
npm --prefix frontend install          # install dependencies
npm --prefix frontend run dev          # dev server (http://localhost:5173)
npm --prefix frontend run build        # production build
npm --prefix frontend run lint         # ESLint
npm --prefix frontend run preview      # preview production build
```

### Backend (`backend/`)

```bash
go run -C backend ./cmd/server         # run dev server (http://localhost:8080)
go build -C backend ./...              # build all
go test -C backend ./...               # run all tests
go test -C backend ./internal/handler/ # run a specific package's tests
```

## Architecture

### Frontend

- **Entry**: `frontend/src/main.tsx` → `App.tsx` (React Router root)
- **Pages**: `frontend/src/pages/` — one component per route
- **Components**: `frontend/src/components/` — shared UI components
- **Hooks**: `frontend/src/hooks/` — custom React hooks
- **API client**: `frontend/src/lib/` — fetch wrappers for backend calls
- **Types**: `frontend/src/types/` — shared TypeScript interfaces
- Vite dev server proxies `/api/*` → `http://localhost:8080`
- Animations use **Framer Motion** (`motion.*` components, `AnimatePresence`)

### Backend

- **Entry**: `backend/cmd/server/main.go` — sets up Gin router and starts HTTP server on `:8080`
- **Handlers**: `backend/internal/handler/` — HTTP handler functions wired to routes
- **Services**: `backend/internal/service/` — business logic
- **Repositories**: `backend/internal/repository/` — data access layer
- **Models**: `backend/internal/model/` — domain types
- **Shared utilities**: `backend/pkg/`
- Go module: `github.com/yumikokawaii/sherry-archive`
