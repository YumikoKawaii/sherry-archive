import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import type { Manga } from '../types/manga'
import type { MangaStatus } from '../types/manga'
import { MangaCard } from '../components/MangaCard'
import { Spinner } from '../components/Spinner'
import { Layout } from '../components/Layout'

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

export function HomePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [mangas, setMangas] = useState<Manga[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)

  const q = searchParams.get('q') ?? ''
  const status = searchParams.get('status') ?? ''
  const sort = searchParams.get('sort') ?? 'newest'

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await mangaApi.list({
        q: q || undefined,
        status: status || undefined,
        sort,
        page,
        limit: 24,
      })
      setMangas(res.items)
      setTotal(res.total)
    } finally {
      setLoading(false)
    }
  }, [q, status, sort, page])

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
