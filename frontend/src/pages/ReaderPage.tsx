import { useEffect, useState, useRef } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import type { ChapterWithPages, Chapter } from '../types/manga'
import { Spinner } from '../components/Spinner'
import { CommentSection } from '../components/CommentSection'
import { tracker } from '../lib/tracking'

export function ReaderPage() {
  const { mangaID, chapterID } = useParams<{ mangaID: string; chapterID: string }>()
  const navigate = useNavigate()
  const [data, setData] = useState<ChapterWithPages | null>(null)
  const [allChapters, setAllChapters] = useState<Chapter[]>([])
  const [loading, setLoading] = useState(true)
  const [headerVisible, setHeaderVisible] = useState(true)
  const [error, setError] = useState('')
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const openedAt = useRef<number>(Date.now())

  useEffect(() => {
    if (!mangaID || !chapterID) return
    setLoading(true)
    setData(null)
    Promise.all([
      mangaApi.getChapter(mangaID, chapterID),
      mangaApi.listChapters(mangaID),
    ])
      .then(([d, chs]) => {
        setData(d)
        setAllChapters([...chs].sort((a, b) => a.number - b.number))
        openedAt.current = Date.now()
        tracker.chapterOpen({
          manga_id: mangaID!,
          chapter_id: chapterID!,
          chapter_number: d.chapter.number,
        })
      })
      .catch(e => setError(e.message ?? 'Failed to load chapter'))
      .finally(() => setLoading(false))
  }, [mangaID, chapterID])

  // Auto-hide header after inactivity
  function showHeader() {
    setHeaderVisible(true)
    if (hideTimer.current) clearTimeout(hideTimer.current)
    hideTimer.current = setTimeout(() => setHeaderVisible(false), 3000)
  }

  const currentIdx = allChapters.findIndex(c => c.id === chapterID)
  const prevChapter = currentIdx > 0 ? allChapters[currentIdx - 1] : null
  const nextChapter = currentIdx < allChapters.length - 1 ? allChapters[currentIdx + 1] : null

  function goTo(ch: Chapter, direction: 'prev' | 'next') {
    if (data) {
      tracker.chapterNavigate({
        from_chapter_id: data.chapter.id,
        to_chapter_id: ch.id,
        direction,
      })
    }
    navigate(`/manga/${mangaID}/chapter/${ch.id}`)
  }

  // Keyboard navigation
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'ArrowLeft' && prevChapter) goTo(prevChapter, 'prev')
      if (e.key === 'ArrowRight' && nextChapter) goTo(nextChapter, 'next')
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [prevChapter, nextChapter])

  if (loading) return (
    <div className="min-h-screen bg-forest-950 flex items-center justify-center">
      <Spinner size="lg" />
    </div>
  )

  if (error || !data) return (
    <div className="min-h-screen bg-forest-950 flex flex-col items-center justify-center gap-4 text-mint-200/40">
      <p className="text-5xl">錯</p>
      <p>{error || 'Chapter not found'}</p>
      <Link to={`/manga/${mangaID}`} className="text-jade-400 hover:underline text-sm">← Back to manga</Link>
    </div>
  )

  return (
    <div className="min-h-screen bg-[#070707]" onMouseMove={showHeader} onClick={showHeader}>
      {/* Floating header */}
      <AnimatePresence>
        {headerVisible && (
          <motion.header
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.18 }}
            className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between
                       px-4 sm:px-6 h-12 bg-forest-950/90 backdrop-blur-md
                       border-b border-forest-700/40"
          >
            <Link to={`/manga/${mangaID}`}
              className="flex items-center gap-2 text-sm text-mint-200/60 hover:text-jade-400 transition">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
              </svg>
              Back
            </Link>

            <div className="text-center">
              <p className="text-xs font-medium text-jade-400">
                Chapter {data.chapter.number}
                {data.chapter.title && ` — ${data.chapter.title}`}
              </p>
              <p className="text-xs text-mint-200/30">{data.pages.length} pages</p>
            </div>

            <div className="flex items-center gap-2">
              {prevChapter && (
                <button onClick={() => goTo(prevChapter, 'prev')}
                  className="text-xs px-3 py-1.5 rounded border border-forest-700
                             text-mint-200/50 hover:text-mint-50 hover:border-jade-500/50 transition">
                  ← Ch.{prevChapter.number}
                </button>
              )}
              {nextChapter && (
                <button onClick={() => goTo(nextChapter, 'next')}
                  className="text-xs px-3 py-1.5 rounded border border-forest-700
                             text-mint-200/50 hover:text-mint-50 hover:border-jade-500/50 transition">
                  Ch.{nextChapter.number} →
                </button>
              )}
            </div>
          </motion.header>
        )}
      </AnimatePresence>

      {/* Pages — vertical scroll */}
      <div className="flex flex-col items-center pt-12 pb-16">
        {data.pages.map((page, i) => {
          const isLast = i === data.pages.length - 1
          return (
            <motion.img
              key={page.id}
              src={page.url}
              alt={`Page ${page.number}`}
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1, transition: { duration: 0.3 } }}
              onViewportEnter={isLast ? () => {
                tracker.chapterComplete({
                  manga_id: mangaID!,
                  chapter_id: chapterID!,
                  duration_seconds: Math.round((Date.now() - openedAt.current) / 1000),
                })
              } : undefined}
              viewport={{ once: true, margin: '200px' }}
              className="w-full max-w-2xl block"
              style={page.width && page.height
                ? { aspectRatio: `${page.width}/${page.height}` }
                : undefined}
              loading={i < 3 ? 'eager' : 'lazy'}
            />
          )
        })}
      </div>

      {/* Bottom nav */}
      <div className="flex justify-center gap-4 pb-10 pt-4">
        {prevChapter ? (
          <button onClick={() => goTo(prevChapter, 'prev')}
            className="px-5 py-2.5 rounded-lg text-sm border border-forest-700 text-mint-200
                       hover:border-jade-500/50 hover:text-mint-50 transition">
            ← Chapter {prevChapter.number}
          </button>
        ) : (
          <Link to={`/manga/${mangaID}`}
            className="px-5 py-2.5 rounded-lg text-sm border border-forest-700 text-mint-200
                       hover:border-jade-500/50 hover:text-mint-50 transition">
            ← Back to manga
          </Link>
        )}
        {nextChapter && (
          <button onClick={() => goTo(nextChapter, 'next')}
            className="px-5 py-2.5 rounded-lg text-sm bg-jade-500 text-forest-950
                       hover:bg-jade-400 font-semibold transition">
            Chapter {nextChapter.number} →
          </button>
        )}
      </div>

      {/* Chapter comments — only for series (number > 0 means it's not an oneshot chapter) */}
      {data.chapter.number > 0 && mangaID && chapterID && (
        <div className="max-w-2xl mx-auto px-4 pb-16">
          <CommentSection
            mangaId={mangaID}
            chapterId={chapterID}
            title={`Chapter ${data.chapter.number} Comments`}
          />
        </div>
      )}
    </div>
  )
}
