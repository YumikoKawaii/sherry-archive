import { useEffect, useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import type { Comment } from '../types/manga'
import type { PagedData } from '../types/api'
import { useAuth } from '../contexts/AuthContext'
import { ApiError } from '../lib/api'

interface Props {
  mangaId: string
  chapterId?: string   // if provided → chapter comments; otherwise → manga comments
  title?: string
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  if (d < 30) return `${d}d ago`
  return new Date(iso).toLocaleDateString()
}

export function CommentSection({ mangaId, chapterId, title = 'Comments' }: Props) {
  const { user } = useAuth()
  const [data, setData] = useState<PagedData<Comment> | null>(null)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [input, setInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editInput, setEditInput] = useState('')
  const [editError, setEditError] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  function load(p: number) {
    setLoading(true)
    const req = chapterId
      ? mangaApi.listChapterComments(mangaId, chapterId, p)
      : mangaApi.listMangaComments(mangaId, p)
    req
      .then(d => { setData(d); setPage(p) })
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load(1) }, [mangaId, chapterId])

  async function handleSubmit() {
    const content = input.trim()
    if (!content) return
    setSubmitting(true)
    setSubmitError('')
    try {
      const comment = chapterId
        ? await mangaApi.createChapterComment(mangaId, chapterId, content)
        : await mangaApi.createMangaComment(mangaId, content)
      setInput('')
      setData(prev => prev
        ? { ...prev, items: [comment, ...prev.items], total: prev.total + 1 }
        : { items: [comment], total: 1, page: 1, limit: 20 })
    } catch (e) {
      setSubmitError(e instanceof ApiError ? e.message : 'Failed to post comment')
    } finally {
      setSubmitting(false)
    }
  }

  function startEdit(c: Comment) {
    setEditingId(c.id)
    setEditInput(c.content)
    setEditError('')
  }

  async function handleEdit(commentId: string) {
    const content = editInput.trim()
    if (!content) return
    setEditError('')
    try {
      const updated = await mangaApi.updateComment(mangaId, commentId, content)
      setData(prev => prev
        ? { ...prev, items: prev.items.map(c => c.id === commentId ? updated : c) }
        : prev)
      setEditingId(null)
    } catch (e) {
      setEditError(e instanceof ApiError ? e.message : 'Failed to update')
    }
  }

  async function handleDelete(commentId: string) {
    try {
      await mangaApi.deleteComment(mangaId, commentId)
      setData(prev => prev
        ? { ...prev, items: prev.items.filter(c => c.id !== commentId), total: prev.total - 1 }
        : prev)
    } catch {}
  }

  const totalPages = data ? Math.ceil(data.total / (data.limit || 20)) : 0

  return (
    <div className="mt-10">
      <h2 className="text-lg font-bold text-mint-50 flex items-center gap-2 mb-4">
        <span className="w-1 h-5 rounded-full bg-jade-500 inline-block" />
        {title}
        {data && data.total > 0 && (
          <span className="text-xs font-normal text-mint-200/30 ml-1">({data.total})</span>
        )}
      </h2>

      {/* Post box */}
      {user && (
        <div className="mb-6">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder="Write a comment…"
            rows={3}
            maxLength={2000}
            className="w-full rounded-lg px-3 py-2.5 text-sm bg-forest-900 border border-forest-700
                       text-mint-50 placeholder-mint-200/25 resize-none
                       focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
          />
          {submitError && (
            <p className="mt-1 text-xs text-red-400">{submitError}</p>
          )}
          <div className="mt-2 flex justify-end">
            <button
              onClick={handleSubmit}
              disabled={submitting || !input.trim()}
              className="px-4 h-8 rounded-lg text-xs font-semibold bg-jade-500 text-forest-950
                         hover:bg-jade-400 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? 'Posting…' : 'Post'}
            </button>
          </div>
        </div>
      )}

      {/* List */}
      {loading ? (
        <div className="text-center py-8 text-mint-200/30 text-sm">Loading…</div>
      ) : !data || data.items.length === 0 ? (
        <p className="text-sm text-mint-200/30 py-8 text-center border border-forest-700 rounded-lg">
          No comments yet.{user ? ' Be the first!' : ''}
        </p>
      ) : (
        <div className="space-y-3">
          <AnimatePresence initial={false}>
            {data.items.map(comment => (
              <motion.div
                key={comment.id}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.97 }}
                transition={{ duration: 0.15 }}
                className="bg-forest-900 border border-forest-700 rounded-lg px-4 py-3"
              >
                <div className="flex items-center justify-between gap-2 mb-2">
                  <div className="flex items-center gap-2 min-w-0">
                    {comment.author.avatar_url ? (
                      <img src={comment.author.avatar_url} alt=""
                        className="w-6 h-6 rounded-full object-cover flex-shrink-0" />
                    ) : (
                      <div className="w-6 h-6 rounded-full bg-forest-700 flex items-center justify-center flex-shrink-0">
                        <span className="text-[10px] text-mint-200/50">
                          {comment.author.username[0]?.toUpperCase()}
                        </span>
                      </div>
                    )}
                    <span className="text-xs font-medium text-mint-200/80 truncate">
                      {comment.author.username}
                    </span>
                    <span className="text-xs text-mint-200/30 flex-shrink-0">
                      {timeAgo(comment.created_at)}
                    </span>
                    {comment.edited && (
                      <span className="text-[10px] text-mint-200/20 flex-shrink-0">(edited)</span>
                    )}
                  </div>

                  {user && user.id === comment.author.id && (
                    <div className="flex items-center gap-2 flex-shrink-0">
                      {editingId !== comment.id && (
                        <button
                          onClick={() => startEdit(comment)}
                          className="text-xs text-mint-200/30 hover:text-jade-400 transition"
                        >
                          Edit
                        </button>
                      )}
                      <button
                        onClick={() => handleDelete(comment.id)}
                        className="text-xs text-mint-200/30 hover:text-red-400 transition"
                      >
                        Delete
                      </button>
                    </div>
                  )}
                </div>

                {editingId === comment.id ? (
                  <div>
                    <textarea
                      value={editInput}
                      onChange={e => setEditInput(e.target.value)}
                      rows={3}
                      maxLength={2000}
                      className="w-full rounded-lg px-3 py-2 text-sm bg-forest-800 border border-forest-600
                                 text-mint-50 resize-none
                                 focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
                    />
                    {editError && <p className="mt-1 text-xs text-red-400">{editError}</p>}
                    <div className="mt-2 flex gap-2 justify-end">
                      <button
                        onClick={() => setEditingId(null)}
                        className="text-xs text-mint-200/40 hover:text-mint-200 transition"
                      >
                        Cancel
                      </button>
                      <button
                        onClick={() => handleEdit(comment.id)}
                        disabled={!editInput.trim()}
                        className="px-3 h-7 rounded text-xs font-semibold bg-jade-500 text-forest-950
                                   hover:bg-jade-400 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                      >
                        Save
                      </button>
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-mint-200/80 whitespace-pre-wrap break-words">
                    {comment.content}
                  </p>
                )}
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-4">
          <button
            disabled={page <= 1}
            onClick={() => load(page - 1)}
            className="px-3 h-8 rounded text-xs border border-forest-700 text-mint-200/50
                       hover:border-jade-500/40 hover:text-mint-200 disabled:opacity-30 transition"
          >
            ← Prev
          </button>
          <span className="flex items-center text-xs text-mint-200/30">
            {page} / {totalPages}
          </span>
          <button
            disabled={page >= totalPages}
            onClick={() => load(page + 1)}
            className="px-3 h-8 rounded text-xs border border-forest-700 text-mint-200/50
                       hover:border-jade-500/40 hover:text-mint-200 disabled:opacity-30 transition"
          >
            Next →
          </button>
        </div>
      )}
    </div>
  )
}
