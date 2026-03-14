import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { bookmarkApi, mangaApi } from '../lib/manga'
import type { Bookmark, Manga } from '../types/manga'
import { Layout } from '../components/Layout'
import { Spinner } from '../components/Spinner'
import { useAuth } from '../contexts/AuthContext'

interface BookmarkEntry {
  bookmark: Bookmark
  manga: Manga
}

export function BookmarksPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [entries, setEntries] = useState<BookmarkEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!user) {
      navigate('/login')
      return
    }
    bookmarkApi.list()
      .then(bookmarks => Promise.all(
        bookmarks.map(b => mangaApi.get(b.manga_id).then(manga => ({ bookmark: b, manga })))
      ))
      .then(setEntries)
      .finally(() => setLoading(false))
  }, [user])

  if (loading) return (
    <Layout>
      <div className="flex justify-center py-32"><Spinner size="lg" /></div>
    </Layout>
  )

  return (
    <Layout>
      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10">
        <h1 className="text-2xl font-bold text-mint-50 mb-8 flex items-center gap-2">
          <span className="w-1 h-6 rounded-full bg-jade-500 inline-block" />
          Bookmarks
        </h1>

        {entries.length === 0 ? (
          <div className="text-center py-24 text-mint-200/30">
            <p className="text-4xl mb-4">栞</p>
            <p className="text-sm">No bookmarks yet.</p>
            <Link to="/" className="mt-4 inline-block text-jade-400 hover:underline text-sm">
              Browse manga →
            </Link>
          </div>
        ) : (
          <div className="space-y-3">
            {entries.map(({ bookmark, manga }, i) => (
              <motion.div
                key={bookmark.id}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.25, delay: i * 0.04 }}
                className="flex items-center gap-4 bg-forest-900 border border-forest-700
                           rounded-xl p-3 hover:border-forest-600 transition-colors group"
              >
                {/* Cover */}
                <Link to={`/manga/${manga.id}`} className="flex-shrink-0">
                  <div className="w-12 h-16 rounded-lg overflow-hidden border border-forest-700">
                    {manga.cover_url ? (
                      <img src={manga.cover_url} alt={manga.title}
                        className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full bg-forest-800 flex items-center justify-center">
                        <span className="text-lg opacity-20">漫</span>
                      </div>
                    )}
                  </div>
                </Link>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <Link to={`/manga/${manga.id}`}
                    className="text-sm font-semibold text-mint-50 group-hover:text-jade-300
                               transition-colors truncate block">
                    {manga.title}
                  </Link>
                  <p className="text-xs text-mint-200/40 mt-0.5">
                    Page {bookmark.last_page_number}
                  </p>
                </div>

                {/* Continue */}
                <Link
                  to={`/manga/${manga.id}/chapter/${bookmark.chapter_id}`}
                  className="flex-shrink-0 px-4 py-2 rounded-lg text-xs font-semibold
                             bg-jade-500/15 text-jade-300 border border-jade-500/30
                             hover:bg-jade-500/25 hover:border-jade-500/50 transition"
                >
                  Continue →
                </Link>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </Layout>
  )
}
