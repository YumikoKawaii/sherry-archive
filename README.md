# 本棚 · Sherry Archive

> A self-hosted manga reading platform built for speed and simplicity.

---

## What is this?

**Sherry Archive** is a clean, minimal manga reader you host yourself. Upload your collection, read from any device, pick up exactly where you left off.

No ads. No tracking. Just your manga.

---

## Features

- **Library** — browse your collection with cover art, tags, and status filters
- **Reader** — smooth vertical scroll, lazy-loaded pages, auto-hiding UI
- **Chapters** — upload via zip, pages ordered automatically by filename
- **Bookmarks** — per-manga progress tracking, resume from last page
- **Auth** — JWT with refresh token rotation, private by default

---

## Stack

```
Backend   Go · Gin · PostgreSQL · MinIO
Frontend  React · TypeScript · Tailwind CSS · Framer Motion
```

---

## Preview

```
┌─────────────────────────────────────────────────────┐
│  sherry archive                          [Login]     │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐      │
│  │      │ │      │ │      │ │      │ │      │       │
│  │cover │ │cover │ │cover │ │cover │ │cover │       │
│  │      │ │      │ │      │ │      │ │      │       │
│  │      │ │      │ │      │ │      │ │      │       │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘      │
│  Title    Title    Title    Title    Title           │
│  ongoing  completed hiatus  ongoing  completed       │
│                                                      │
└─────────────────────────────────────────────────────┘
```

---

*Built with 抹茶 and late nights.*
