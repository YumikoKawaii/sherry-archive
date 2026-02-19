import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import type { Manga, Chapter } from '../types/manga'
import { Layout } from '../components/Layout'
import { StatusBadge } from '../components/StatusBadge'
import { TagBadge } from '../components/TagBadge'
import { Spinner } from '../components/Spinner'

export function MangaDetailPage() {
  const { mangaID } = useParams<{ mangaID: string }>()
  const [manga, setManga] = useState<Manga | null>(null)
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!mangaID) return
    setLoading(true)
    Promise.all([
      mangaApi.get(mangaID),
      mangaApi.listChapters(mangaID),
    ])
      .then(([m, chs]) => {
        setManga(m)
        // Sort descending by chapter number (newest first)
        setChapters([...chs].sort((a, b) => b.number - a.number))
      })
      .catch(e => setError(e.message ?? 'Failed to load'))
      .finally(() => setLoading(false))
  }, [mangaID])

  if (loading) return <Layout><div className="flex justify-center py-32"><Spinner size="lg" /></div></Layout>
  if (error || !manga) return (
    <Layout>
      <div className="flex flex-col items-center justify-center py-32 text-mint-200/40">
        <p className="text-5xl mb-4">錯</p>
        <p>{error || 'Manga not found'}</p>
        <Link to="/" className="mt-6 text-jade-400 hover:underline text-sm">← Back to browse</Link>
      </div>
    </Layout>
  )

  const firstChapter = chapters[chapters.length - 1]
  const lastChapter = chapters[0]

  return (
    <Layout>
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
        <div className="flex flex-col sm:flex-row gap-8">
          {/* Cover */}
          <motion.div
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.35 }}
            className="flex-shrink-0 w-full sm:w-52"
          >
            <div className="aspect-[3/4] rounded-lg overflow-hidden border border-forest-700
                            shadow-[0_0_30px_rgba(34,197,94,0.08)]">
              {manga.cover_url ? (
                <img src={manga.cover_url} alt={manga.title}
                  className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full bg-forest-800 flex items-center justify-center">
                  <span className="text-6xl opacity-10">漫</span>
                </div>
              )}
            </div>
          </motion.div>

          {/* Info */}
          <motion.div
            initial={{ opacity: 0, x: 12 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.35, delay: 0.1 }}
            className="flex-1 min-w-0"
          >
            <div className="flex items-start gap-3 flex-wrap">
              <h1 className="text-2xl sm:text-3xl font-bold text-mint-50 leading-tight flex-1">
                {manga.title}
              </h1>
              <StatusBadge status={manga.status} />
            </div>

            {manga.tags.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mt-3">
                {manga.tags.map(tag => <TagBadge key={tag} tag={tag} />)}
              </div>
            )}

            {manga.description && (
              <p className="mt-4 text-mint-200/70 text-sm leading-relaxed line-clamp-3">
                {manga.description}
              </p>
            )}

            <div className="mt-4 text-xs text-mint-200/30 space-y-0.5">
              <p>{chapters.length} chapter{chapters.length !== 1 ? 's' : ''}</p>
            </div>

            {/* CTA buttons */}
            <div className="mt-6 flex flex-wrap gap-3">
              {firstChapter && (
                <Link
                  to={`/manga/${manga.id}/chapter/${firstChapter.id}`}
                  className="px-5 py-2.5 rounded-lg text-sm font-semibold
                             bg-jade-500 text-forest-950 hover:bg-jade-400 transition-colors"
                >
                  Start Reading →
                </Link>
              )}
              {lastChapter && lastChapter.id !== firstChapter?.id && (
                <Link
                  to={`/manga/${manga.id}/chapter/${lastChapter.id}`}
                  className="px-5 py-2.5 rounded-lg text-sm font-medium border border-forest-600
                             text-mint-200 hover:border-jade-500/50 hover:text-mint-50 transition"
                >
                  Latest Chapter
                </Link>
              )}
            </div>
          </motion.div>
        </div>

        {/* Chapter list */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
          className="mt-10"
        >
          <h2 className="text-lg font-bold text-mint-50 mb-4 flex items-center gap-2">
            <span className="w-1 h-5 rounded-full bg-jade-500 inline-block" />
            Chapters
          </h2>

          {chapters.length === 0 ? (
            <p className="text-mint-200/30 text-sm py-8 text-center border border-forest-700 rounded-lg">
              No chapters yet.
            </p>
          ) : (
            <div className="border border-forest-700 rounded-lg overflow-hidden divide-y divide-forest-700/60">
              {chapters.map(ch => (
                <Link
                  key={ch.id}
                  to={`/manga/${manga.id}/chapter/${ch.id}`}
                  className="flex items-center justify-between px-4 py-3.5 group
                             hover:bg-forest-800/60 transition-colors"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <span className="text-xs font-mono text-jade-400 flex-shrink-0 w-16">
                      Ch. {ch.number}
                    </span>
                    <span className="text-sm text-mint-200 group-hover:text-mint-50 truncate transition-colors">
                      {ch.title || `Chapter ${ch.number}`}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 flex-shrink-0 ml-4">
                    <span className="text-xs text-mint-200/30">{ch.page_count}p</span>
                    <span className="text-xs text-mint-200/20">
                      {new Date(ch.created_at).toLocaleDateString()}
                    </span>
                    <svg className="w-4 h-4 text-mint-200/20 group-hover:text-jade-400 transition-colors"
                      fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </motion.div>
      </div>
    </Layout>
  )
}
