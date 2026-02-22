import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { mangaApi, analyticsApi } from '../lib/manga'
import type { Manga, MangaStatus } from '../types/manga'
import type { TrendingItem } from '../lib/manga'
import { MangaCard } from '../components/MangaCard'
import { Spinner } from '../components/Spinner'
import { Layout } from '../components/Layout'
import { getDeviceId } from '../lib/tracking'

const STATUSES: { value: MangaStatus | ''; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'ongoing', label: 'Ongoing' },
  { value: 'completed', label: 'Completed' },
  { value: 'hiatus', label: 'Hiatus' },
]

const SORTS = [
  { value: 'newest', label: 'Newest' },
  { value: 'oldest', label: 'Oldest' },
  { value: 'title', label: 'Title A–Z' },
]

function MangaShelf({ title, items, badge }: {
  title: string
  items: Manga[]
  badge?: (m: Manga) => string | undefined
}) {
  if (items.length === 0) return null
  return (
    <div>
      <h2 className="text-sm font-semibold tracking-[0.2em] text-jade-500 uppercase mb-4">{title}</h2>
      <div className="flex gap-4 overflow-x-auto pb-2 -mx-1 px-1 scrollbar-none snap-x">
        {items.map(manga => (
          <div key={manga.id} className="flex-none w-36 sm:w-40 snap-start relative">
            <MangaCard manga={manga} />
            {badge && badge(manga) && (
              <div className="absolute top-2 left-2 bg-jade-500/90 text-forest-950 text-[10px]
                              font-bold px-1.5 py-0.5 rounded leading-none">
                {badge(manga)}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export function HomePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [mangas, setMangas] = useState<Manga[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)

  const [trending, setTrending] = useState<TrendingItem[]>([])
  const [trendingLoaded, setTrendingLoaded] = useState(false)
  const [suggestions, setSuggestions] = useState<Manga[]>([])

  const q = searchParams.get('q') ?? ''
  const status = searchParams.get('status') ?? ''
  const sort = searchParams.get('sort') ?? 'newest'
  const author = searchParams.get('author') ?? ''
  const category = searchParams.get('category') ?? ''

  // Load trending & suggestions once on mount
  useEffect(() => {
    analyticsApi.trending(12)
      .then(res => setTrending(res.data))
      .catch(() => {})
      .finally(() => setTrendingLoaded(true))
    const deviceId = getDeviceId()
    analyticsApi.suggestions(deviceId, 12).then(res => setSuggestions(res.data)).catch(() => {})
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await mangaApi.list({
        q: q || undefined,
        status: status || undefined,
        sort,
        page,
        limit: 24,
        author: author || undefined,
        category: category || undefined,
      })
      setMangas(res.items)
      setTotal(res.total)
    } finally {
      setLoading(false)
    }
  }, [q, status, sort, page, author, category])

  useEffect(() => { load() }, [load])

  function setFilter(key: string, value: string) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (value) next.set(key, value)
      else next.delete(key)
      return next
    })
    setPage(1)
  }

  const totalPages = Math.ceil(total / 24)

  return (
    <Layout>
      {/* Hero */}
      <section className="relative overflow-hidden bg-grid border-b border-forest-700/40">
        <div className="absolute inset-0 bg-gradient-to-b from-jade-500/5 to-transparent pointer-events-none" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-16 sm:py-24">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="max-w-xl"
          >
            <p className="text-xs font-semibold tracking-[0.3em] text-jade-500 mb-3 uppercase">
              Manga Archive
            </p>
            <h1 className="text-4xl sm:text-5xl font-black text-mint-50 leading-tight">
              Read what
              <span className="text-jade-400"> moves </span>
              you.
            </h1>
            <p className="mt-4 text-mint-200/60 text-lg leading-relaxed">
              Browse thousands of manga, track your progress, and pick up exactly where you left off.
            </p>
          </motion.div>
        </div>
      </section>

      {/* Trending & Suggestions shelves */}
      {(trendingLoaded || suggestions.length > 0) && (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 pt-8 pb-2 space-y-8">
          {trending.length > 0 ? (
            <MangaShelf
              title="Trending"
              items={trending}
              badge={m => {
                const score = (m as TrendingItem).trending_score
                return score > 0 ? `↑${Math.round(score)}` : undefined
              }}
            />
          ) : (
            // Fallback to newest when trending ZSET is still cold
            !loading && mangas.length > 0 && (
              <MangaShelf title="New Arrivals" items={mangas.slice(0, 12)} />
            )
          )}
          <MangaShelf title="For You" items={suggestions} />
        </div>
      )}

      {/* Filters */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 py-5">
        <div className="flex flex-wrap items-center gap-3">
          {/* Status filter */}
          <div className="flex items-center gap-1.5 bg-forest-900 border border-forest-700 rounded-lg p-1">
            {STATUSES.map(s => (
              <button
                key={s.value}
                onClick={() => setFilter('status', s.value)}
                className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${
                  status === s.value
                    ? 'bg-jade-500 text-forest-950'
                    : 'text-mint-200/60 hover:text-mint-50'
                }`}
              >
                {s.label}
              </button>
            ))}
          </div>

          {/* Author filter */}
          <input
            type="text"
            value={author}
            onChange={e => setFilter('author', e.target.value)}
            placeholder="Author…"
            className="h-9 px-3 rounded-lg text-sm bg-forest-900 border border-forest-700
                       text-mint-200 placeholder-mint-200/30
                       focus:outline-none focus:border-jade-500/60 transition w-32"
          />

          {/* Category filter */}
          <input
            type="text"
            value={category}
            onChange={e => setFilter('category', e.target.value)}
            placeholder="Category…"
            className="h-9 px-3 rounded-lg text-sm bg-forest-900 border border-forest-700
                       text-mint-200 placeholder-mint-200/30
                       focus:outline-none focus:border-jade-500/60 transition w-32"
          />

          {/* Sort */}
          <select
            value={sort}
            onChange={e => setFilter('sort', e.target.value)}
            className="h-9 px-3 rounded-lg text-sm bg-forest-900 border border-forest-700
                       text-mint-200 focus:outline-none focus:border-jade-500/60 cursor-pointer"
          >
            {SORTS.map(s => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>

          {/* Result count */}
          <span className="ml-auto text-sm text-mint-200/40">
            {total} {total === 1 ? 'title' : 'titles'}
          </span>
        </div>
      </div>

      {/* Grid */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 pb-16">
        {loading ? (
          <div className="flex justify-center py-24"><Spinner size="lg" /></div>
        ) : mangas.length === 0 ? (
          <div className="text-center py-24 text-mint-200/30">
            <p className="text-5xl mb-4">空</p>
            <p className="text-lg">No manga found.</p>
          </div>
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.3 }}
            className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4"
          >
            {mangas.map(manga => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </motion.div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center items-center gap-2 mt-10">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-4 py-2 rounded-md text-sm border border-forest-700 text-mint-200/60
                         hover:border-jade-500/50 hover:text-mint-50 disabled:opacity-30
                         disabled:cursor-not-allowed transition"
            >
              ← Prev
            </button>
            <span className="text-sm text-mint-200/40">
              {page} / {totalPages}
            </span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-4 py-2 rounded-md text-sm border border-forest-700 text-mint-200/60
                         hover:border-jade-500/50 hover:text-mint-50 disabled:opacity-30
                         disabled:cursor-not-allowed transition"
            >
              Next →
            </button>
          </div>
        )}
      </div>
    </Layout>
  )
}
