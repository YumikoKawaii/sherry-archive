import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import type { Manga, Chapter } from '../types/manga'
import type { MangaStatus, MangaType } from '../types/manga'
import { Layout } from '../components/Layout'
import { StatusBadge } from '../components/StatusBadge'
import { TagBadge } from '../components/TagBadge'
import { Spinner } from '../components/Spinner'
import { useAuth } from '../contexts/AuthContext'
import { ApiError } from '../lib/api'
import { CommentSection } from '../components/CommentSection'
import { tracker } from '../lib/tracking'

const STATUSES: { value: MangaStatus; label: string }[] = [
  { value: 'ongoing', label: 'Ongoing' },
  { value: 'completed', label: 'Completed' },
  { value: 'hiatus', label: 'Hiatus' },
]

const TYPES: { value: MangaType; label: string }[] = [
  { value: 'series', label: 'Series' },
  { value: 'oneshot', label: 'Oneshot' },
]

export function MangaDetailPage() {
  const { mangaID } = useParams<{ mangaID: string }>()
  const { user } = useAuth()
  const [manga, setManga] = useState<Manga | null>(null)
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Edit panel state
  const [editing, setEditing] = useState(false)
  const [editStatus, setEditStatus] = useState<MangaStatus>('ongoing')
  const [editType, setEditType] = useState<MangaType>('series')
  const [editAuthor, setEditAuthor] = useState('')
  const [editArtist, setEditArtist] = useState('')
  const [editCategory, setEditCategory] = useState('')
  const [editTags, setEditTags] = useState<string[]>([])
  const [editTagInput, setEditTagInput] = useState('')
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  useEffect(() => {
    if (!mangaID) return
    setLoading(true)
    Promise.all([
      mangaApi.get(mangaID),
      mangaApi.listChapters(mangaID),
    ])
      .then(([m, chs]) => {
        setManga(m)
        setChapters([...chs].sort((a, b) => b.number - a.number))
        tracker.mangaView({ manga_id: m.id, manga_type: m.type })
      })
      .catch(e => setError(e.message ?? 'Failed to load'))
      .finally(() => setLoading(false))
  }, [mangaID])

  function openEdit() {
    if (!manga) return
    setEditStatus(manga.status)
    setEditType(manga.type)
    setEditAuthor(manga.author)
    setEditArtist(manga.artist)
    setEditCategory(manga.category)
    setEditTags([...manga.tags])
    setEditTagInput('')
    setSaveError('')
    setEditing(true)
  }

  function addEditTag() {
    const t = editTagInput.trim().toLowerCase()
    if (t && !editTags.includes(t)) setEditTags(prev => [...prev, t])
    setEditTagInput('')
  }

  function handleEditTagKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      addEditTag()
    }
  }

  async function handleSave() {
    if (!mangaID || !manga) return
    setSaving(true)
    setSaveError('')
    try {
      const updated = await mangaApi.update(mangaID, {
        status: editStatus,
        type: editType,
        author: editAuthor,
        artist: editArtist,
        category: editCategory,
        tags: editTags,
      })
      setManga(updated)
      setEditing(false)
    } catch (err) {
      setSaveError(err instanceof ApiError ? err.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

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

  const isOwner = user && manga.owner_id === user.id
  const isOneshot = manga.type === 'oneshot'
  const firstChapter = chapters[chapters.length - 1]
  const lastChapter = chapters[0]
  const oneshotChapter = isOneshot ? chapters[0] : null

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
              {isOneshot && (
                <span className="px-2 py-0.5 rounded text-xs font-medium bg-forest-800 border border-forest-600 text-mint-200/50">
                  Oneshot
                </span>
              )}
            </div>

            {(manga.author || manga.artist) && (
              <div className="mt-2 flex flex-wrap gap-x-4 gap-y-0.5 text-xs text-mint-200/50">
                {manga.author && <span>by <span className="text-mint-200/80">{manga.author}</span></span>}
                {manga.artist && manga.artist !== manga.author && (
                  <span>art <span className="text-mint-200/80">{manga.artist}</span></span>
                )}
                {manga.category && <span className="text-mint-200/40">{manga.category}</span>}
              </div>
            )}

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

            {!isOneshot && (
              <div className="mt-4 text-xs text-mint-200/30">
                <p>{chapters.length} chapter{chapters.length !== 1 ? 's' : ''}</p>
              </div>
            )}

            {/* CTA buttons */}
            <div className="mt-6 flex flex-wrap gap-3">
              {isOneshot ? (
                oneshotChapter ? (
                  <Link
                    to={`/manga/${manga.id}/chapter/${oneshotChapter.id}`}
                    className="px-5 py-2.5 rounded-lg text-sm font-semibold
                               bg-jade-500 text-forest-950 hover:bg-jade-400 transition-colors"
                  >
                    Read Oneshot →
                  </Link>
                ) : (
                  <span className="text-sm text-mint-200/30 italic">No chapter yet.</span>
                )
              ) : (
                <>
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
                </>
              )}

              {/* Owner actions */}
              {isOwner && (
                <button
                  onClick={editing ? () => setEditing(false) : openEdit}
                  className={`px-4 py-2.5 rounded-lg text-sm font-medium border transition ${
                    editing
                      ? 'bg-jade-500/20 border-jade-500/60 text-jade-300'
                      : 'border-forest-600 text-mint-200/50 hover:border-jade-500/40 hover:text-jade-300'
                  }`}
                >
                  {editing ? 'Cancel edit' : 'Edit'}
                </button>
              )}
            </div>
          </motion.div>
        </div>

        {/* Inline edit panel */}
        <AnimatePresence>
          {editing && (
            <motion.div
              key="edit-panel"
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.2 }}
              className="mt-6 bg-forest-900 border border-forest-700 rounded-xl p-6 space-y-5"
            >
              {/* Status */}
              <div>
                <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Status</label>
                <div className="flex gap-2">
                  {STATUSES.map(s => (
                    <button
                      key={s.value}
                      type="button"
                      onClick={() => setEditStatus(s.value)}
                      className={`flex-1 h-9 rounded-lg text-xs font-medium transition border ${
                        editStatus === s.value
                          ? 'bg-jade-500/20 border-jade-500/60 text-jade-300'
                          : 'bg-forest-800 border-forest-600 text-mint-200/50 hover:border-forest-500 hover:text-mint-200/80'
                      }`}
                    >
                      {s.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Type */}
              <div>
                <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Type</label>
                <div className="flex gap-2 max-w-xs">
                  {TYPES.map(t => (
                    <button
                      key={t.value}
                      type="button"
                      onClick={() => setEditType(t.value)}
                      className={`flex-1 h-9 rounded-lg text-xs font-medium transition border ${
                        editType === t.value
                          ? 'bg-jade-500/20 border-jade-500/60 text-jade-300'
                          : 'bg-forest-800 border-forest-600 text-mint-200/50 hover:border-forest-500 hover:text-mint-200/80'
                      }`}
                    >
                      {t.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Author / Artist / Category */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div>
                  <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Author</label>
                  <input
                    type="text"
                    value={editAuthor}
                    onChange={e => setEditAuthor(e.target.value)}
                    placeholder="Author name"
                    className="w-full h-9 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 placeholder-mint-200/20
                               focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Artist</label>
                  <input
                    type="text"
                    value={editArtist}
                    onChange={e => setEditArtist(e.target.value)}
                    placeholder="Artist name"
                    className="w-full h-9 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 placeholder-mint-200/20
                               focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Category</label>
                  <input
                    type="text"
                    value={editCategory}
                    onChange={e => setEditCategory(e.target.value)}
                    placeholder="e.g. doujinshi"
                    className="w-full h-9 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 placeholder-mint-200/20
                               focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
                  />
                </div>
              </div>

              {/* Tags */}
              <div>
                <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Tags</label>
                <div className="rounded-lg bg-forest-800 border border-forest-600
                                focus-within:border-jade-500/60 focus-within:ring-1 focus-within:ring-jade-500/25
                                transition p-2 min-h-[42px] flex flex-wrap gap-1.5 items-center">
                  {editTags.map(tag => (
                    <span key={tag}
                      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs
                                 bg-jade-500/15 text-jade-300 border border-jade-500/25">
                      {tag}
                      <button type="button" onClick={() => setEditTags(prev => prev.filter(t => t !== tag))}
                        className="text-jade-400/60 hover:text-jade-300 transition leading-none">×</button>
                    </span>
                  ))}
                  <input
                    type="text"
                    value={editTagInput}
                    onChange={e => setEditTagInput(e.target.value)}
                    onKeyDown={handleEditTagKeyDown}
                    onBlur={addEditTag}
                    placeholder={editTags.length === 0 ? 'Type a tag, press Enter…' : ''}
                    className="flex-1 min-w-[120px] bg-transparent text-sm text-mint-50
                               placeholder-mint-200/20 outline-none"
                  />
                </div>
              </div>

              {saveError && (
                <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
                  {saveError}
                </p>
              )}

              <div className="flex gap-3">
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="px-5 h-9 rounded-lg text-sm font-semibold bg-jade-500 text-forest-950
                             hover:bg-jade-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {saving ? 'Saving…' : 'Save'}
                </button>
                <button
                  onClick={() => setEditing(false)}
                  className="px-5 h-9 rounded-lg text-sm font-medium border border-forest-600
                             text-mint-200/50 hover:text-mint-50 hover:border-forest-500 transition"
                >
                  Cancel
                </button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Chapter list — only shown for series */}
        {!isOneshot && (
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.2 }}
            className="mt-10"
          >
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-bold text-mint-50 flex items-center gap-2">
                <span className="w-1 h-5 rounded-full bg-jade-500 inline-block" />
                Chapters
              </h2>
              {isOwner && (
                <Link
                  to={`/manga/${manga.id}/manage`}
                  className="text-xs px-3 py-1.5 rounded border border-forest-600
                             text-mint-200/50 hover:border-jade-500/40 hover:text-jade-300 transition"
                >
                  Manage chapters
                </Link>
              )}
            </div>

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
        )}

        {/* Manage link for oneshot owners */}
        {isOneshot && isOwner && (
          <div className="mt-6">
            <Link
              to={`/manga/${manga.id}/manage`}
              className="text-xs px-3 py-1.5 rounded border border-forest-600
                         text-mint-200/50 hover:border-jade-500/40 hover:text-jade-300 transition"
            >
              Manage chapters
            </Link>
          </div>
        )}

        {/* Comments — manga-level, shown for all types */}
        <CommentSection mangaId={manga.id} />
      </div>
    </Layout>
  )
}
