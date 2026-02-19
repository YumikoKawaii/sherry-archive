import { useEffect, useRef, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import type { ZipMetadataSuggestions, OneshotUploadResult } from '../lib/manga'
import { ApiError } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { Layout } from '../components/Layout'
import { Spinner } from '../components/Spinner'
import type { Manga, Chapter } from '../types/manga'

export function ManageChaptersPage() {
  const { mangaID } = useParams<{ mangaID: string }>()
  const { user } = useAuth()
  const navigate = useNavigate()

  const [manga, setManga] = useState<Manga | null>(null)
  const [chapters, setChapters] = useState<Chapter[]>([])
  const [loading, setLoading] = useState(true)
  const [pageError, setPageError] = useState('')

  // New chapter form
  const [newNumber, setNewNumber] = useState('')
  const [newTitle, setNewTitle] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  // ZIP metadata suggestions (shown after a zip upload on any chapter row)
  const [suggestions, setSuggestions] = useState<ZipMetadataSuggestions | null>(null)
  const [applyingMeta, setApplyingMeta] = useState(false)
  const [applyMetaError, setApplyMetaError] = useState('')
  const [applyMetaSuccess, setApplyMetaSuccess] = useState(false)

  useEffect(() => {
    if (!mangaID) return
    Promise.all([mangaApi.get(mangaID), mangaApi.listChapters(mangaID)])
      .then(([m, chs]) => {
        setManga(m)
        setChapters([...chs].sort((a, b) => a.number - b.number))
      })
      .catch(e => setPageError(e.message ?? 'Failed to load'))
      .finally(() => setLoading(false))
  }, [mangaID])

  // Redirect non-owners
  useEffect(() => {
    if (!loading && manga && user && manga.owner_id !== user.id) {
      navigate(`/manga/${mangaID}`)
    }
  }, [loading, manga, user, mangaID, navigate])

  const isOneshot = manga?.type === 'oneshot'
  const oneshotHasChapter = isOneshot && chapters.length > 0

  // Oneshot direct upload
  const oneshotFileRef = useRef<HTMLInputElement>(null)
  const [oneshotUploading, setOneshotUploading] = useState(false)
  const [oneshotUploadError, setOneshotUploadError] = useState('')

  async function handleOneshotUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file || !mangaID) return
    setOneshotUploading(true)
    setOneshotUploadError('')
    try {
      const result: OneshotUploadResult = await mangaApi.oneshotUpload(mangaID, file)
      setChapters([result.chapter])
      if (result.metadata_suggestions) {
        setSuggestions(result.metadata_suggestions)
        setApplyMetaSuccess(false)
      }
    } catch (err) {
      setOneshotUploadError(err instanceof ApiError ? err.message : 'Upload failed')
    } finally {
      setOneshotUploading(false)
      if (oneshotFileRef.current) oneshotFileRef.current.value = ''
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!mangaID) return
    setCreateError('')
    setCreating(true)
    try {
      const ch = await mangaApi.createChapter(mangaID, {
        number: parseFloat(newNumber),
        title: newTitle,
      })
      setChapters(prev => [...prev, ch].sort((a, b) => a.number - b.number))
      setNewNumber('')
      setNewTitle('')
    } catch (err) {
      setCreateError(err instanceof ApiError ? err.message : 'Failed to create chapter')
    } finally {
      setCreating(false)
    }
  }

  async function handleApplySuggestions() {
    if (!mangaID || !suggestions) return
    setApplyingMeta(true)
    setApplyMetaError('')
    setApplyMetaSuccess(false)
    try {
      const patch: Record<string, unknown> = {}
      if (suggestions.author) patch.author = suggestions.author
      if (suggestions.artist) patch.artist = suggestions.artist
      if (suggestions.category) patch.category = suggestions.category
      if (suggestions.tags && suggestions.tags.length > 0) patch.tags = suggestions.tags
      await mangaApi.update(mangaID, patch)
      setApplyMetaSuccess(true)
      setSuggestions(null)
    } catch (err) {
      setApplyMetaError(err instanceof ApiError ? err.message : 'Failed to apply')
    } finally {
      setApplyingMeta(false)
    }
  }

  if (loading) return <Layout><div className="flex justify-center py-32"><Spinner size="lg" /></div></Layout>
  if (pageError || !manga) return (
    <Layout>
      <div className="flex flex-col items-center justify-center py-32 text-mint-200/40">
        <p className="text-5xl mb-4">錯</p>
        <p>{pageError || 'Not found'}</p>
        <Link to="/" className="mt-6 text-jade-400 hover:underline text-sm">← Back to browse</Link>
      </div>
    </Layout>
  )

  return (
    <Layout>
      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10">
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.3 }}>

          {/* Header */}
          <div className="mb-8">
            <Link to={`/manga/${manga.id}`}
              className="text-xs text-mint-200/40 hover:text-mint-200/70 transition mb-3 inline-block">
              ← {manga.title}
            </Link>
            <h1 className="text-2xl font-bold text-mint-50 flex items-center gap-2">
              <span className="w-1 h-6 rounded-full bg-jade-500 inline-block" />
              Manage Chapters
              {isOneshot && (
                <span className="text-xs font-normal px-2 py-0.5 rounded bg-forest-800 border border-forest-600 text-mint-200/40 ml-1">
                  Oneshot
                </span>
              )}
            </h1>
          </div>

          {/* Add chapter / oneshot upload panel */}
          {isOneshot ? (
            !oneshotHasChapter && (
              <div className="bg-forest-900 border border-forest-700 rounded-xl p-5 mb-6">
                <h2 className="text-sm font-semibold text-mint-200 mb-1">Upload Oneshot</h2>
                <p className="text-xs text-mint-200/40 mb-4">
                  Upload a ZIP of images. A chapter will be created automatically.
                  Include a <code className="text-jade-400/80">metadata.json</code> in the ZIP to pre-fill author, tags and more.
                </p>
                <div className="flex items-center gap-3">
                  <input
                    ref={oneshotFileRef}
                    type="file"
                    accept=".zip,application/zip"
                    onChange={handleOneshotUpload}
                    disabled={oneshotUploading}
                    className="hidden"
                  />
                  <button
                    onClick={() => oneshotFileRef.current?.click()}
                    disabled={oneshotUploading}
                    className="h-9 px-4 rounded-lg text-sm font-medium bg-jade-500 text-forest-950
                               hover:bg-jade-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {oneshotUploading ? 'Uploading…' : 'Choose ZIP'}
                  </button>
                  {oneshotUploading && <Spinner size="sm" />}
                </div>
                {oneshotUploadError && (
                  <p className="mt-3 text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
                    {oneshotUploadError}
                  </p>
                )}
              </div>
            )
          ) : (
            <div className="bg-forest-900 border border-forest-700 rounded-xl p-5 mb-6">
              <h2 className="text-sm font-semibold text-mint-200 mb-4">Add Chapter</h2>
              <form onSubmit={handleCreate} className="flex gap-3 flex-wrap">
                <div>
                  <label className="block text-xs text-mint-200/50 mb-1">Number *</label>
                  <input
                    type="number"
                    step="0.1"
                    min="0"
                    value={newNumber}
                    onChange={e => setNewNumber(e.target.value)}
                    placeholder="1"
                    required
                    className="w-24 h-9 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 placeholder-mint-200/20
                               focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
                  />
                </div>
                <div className="flex-1 min-w-40">
                  <label className="block text-xs text-mint-200/50 mb-1">Title</label>
                  <input
                    type="text"
                    value={newTitle}
                    onChange={e => setNewTitle(e.target.value)}
                    placeholder="Optional title"
                    className="w-full h-9 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 placeholder-mint-200/20
                               focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25 transition"
                  />
                </div>
                <div className="flex items-end">
                  <button
                    type="submit"
                    disabled={creating}
                    className="h-9 px-4 rounded-lg text-sm font-medium bg-jade-500 text-forest-950
                               hover:bg-jade-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {creating ? 'Adding…' : 'Add'}
                  </button>
                </div>
              </form>
              {createError && (
                <p className="mt-3 text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
                  {createError}
                </p>
              )}
            </div>
          )}

          {/* Metadata suggestions card */}
          <AnimatePresence>
            {suggestions && (
              <motion.div
                key="suggestions"
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.2 }}
                className="bg-jade-500/5 border border-jade-500/20 rounded-xl p-5 mb-6"
              >
                <div className="flex items-start justify-between gap-3 mb-3">
                  <h2 className="text-sm font-semibold text-jade-300">Metadata detected in ZIP</h2>
                  <button
                    onClick={() => setSuggestions(null)}
                    className="text-mint-200/30 hover:text-mint-200/60 transition text-lg leading-none"
                  >×</button>
                </div>
                <div className="space-y-1 text-xs text-mint-200/60 mb-4">
                  {suggestions.author && <p><span className="text-mint-200/40">Author:</span> {suggestions.author}</p>}
                  {suggestions.artist && <p><span className="text-mint-200/40">Artist:</span> {suggestions.artist}</p>}
                  {suggestions.category && <p><span className="text-mint-200/40">Category:</span> {suggestions.category}</p>}
                  {suggestions.tags && suggestions.tags.length > 0 && (
                    <p><span className="text-mint-200/40">Tags:</span> {suggestions.tags.join(', ')}</p>
                  )}
                  {suggestions.language && <p><span className="text-mint-200/40">Language:</span> {suggestions.language}</p>}
                  {suggestions.chapter_title && <p><span className="text-mint-200/40">Chapter title:</span> {suggestions.chapter_title}</p>}
                </div>
                <div className="flex items-center gap-3">
                  <button
                    onClick={handleApplySuggestions}
                    disabled={applyingMeta}
                    className="h-8 px-4 rounded-lg text-xs font-medium bg-jade-500 text-forest-950
                               hover:bg-jade-400 disabled:opacity-50 transition-colors"
                  >
                    {applyingMeta ? 'Applying…' : 'Apply to manga'}
                  </button>
                  <button
                    onClick={() => setSuggestions(null)}
                    className="h-8 px-3 rounded-lg text-xs border border-forest-600
                               text-mint-200/50 hover:text-mint-50 transition"
                  >
                    Dismiss
                  </button>
                </div>
                {applyMetaError && (
                  <p className="mt-2 text-xs text-red-400">{applyMetaError}</p>
                )}
              </motion.div>
            )}
            {applyMetaSuccess && (
              <motion.div
                key="apply-success"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="mb-4 text-xs text-jade-300 bg-jade-500/10 border border-jade-500/20 rounded-lg px-3 py-2"
              >
                Manga metadata updated from ZIP suggestions.
              </motion.div>
            )}
          </AnimatePresence>

          {/* Chapter list */}
          <div>
            <h2 className="text-sm font-semibold text-mint-200/60 mb-3 flex items-center gap-2">
              {chapters.length} chapter{chapters.length !== 1 ? 's' : ''}
            </h2>

            {chapters.length === 0 ? (
              <p className="text-mint-200/30 text-sm py-8 text-center border border-forest-700 rounded-lg">
                No chapters yet — add one above.
              </p>
            ) : (
              <div className="border border-forest-700 rounded-xl overflow-hidden divide-y divide-forest-700/60">
                <AnimatePresence initial={false}>
                  {chapters.map(ch => (
                    <ChapterRow
                      key={ch.id}
                      chapter={ch}
                      mangaId={manga.id}
                      isOneshot={!!isOneshot}
                      onUpdate={updated => setChapters(prev =>
                        [...prev.map(c => c.id === updated.id ? updated : c)].sort((a, b) => a.number - b.number)
                      )}
                      onDelete={id => setChapters(prev => prev.filter(c => c.id !== id))}
                      onSuggestions={meta => { setSuggestions(meta); setApplyMetaSuccess(false) }}
                    />
                  ))}
                </AnimatePresence>
              </div>
            )}
          </div>
        </motion.div>
      </div>
    </Layout>
  )
}

// ─── Chapter Row ─────────────────────────────────────────────────────────────

interface ChapterRowProps {
  chapter: Chapter
  mangaId: string
  isOneshot: boolean
  onUpdate: (ch: Chapter) => void
  onDelete: (id: string) => void
  onSuggestions: (meta: ZipMetadataSuggestions) => void
}

type RowMode = 'idle' | 'edit' | 'delete' | 'upload'

function ChapterRow({ chapter, mangaId, isOneshot, onUpdate, onDelete, onSuggestions }: ChapterRowProps) {
  const [mode, setMode] = useState<RowMode>('idle')
  const [busy, setBusy] = useState(false)
  const [rowError, setRowError] = useState('')

  // Edit state
  const [editNumber, setEditNumber] = useState(String(chapter.number))
  const [editTitle, setEditTitle] = useState(chapter.title)

  // Upload state
  const fileRef = useRef<HTMLInputElement>(null)
  const [uploadName, setUploadName] = useState('')

  function openEdit() {
    setEditNumber(String(chapter.number))
    setEditTitle(chapter.title)
    setRowError('')
    setMode('edit')
  }

  async function handleSave() {
    setBusy(true)
    setRowError('')
    try {
      const updated = await mangaApi.updateChapter(mangaId, chapter.id, {
        number: isOneshot ? undefined : parseFloat(editNumber),
        title: editTitle,
      })
      onUpdate(updated)
      setMode('idle')
    } catch (err) {
      setRowError(err instanceof ApiError ? err.message : 'Update failed')
    } finally {
      setBusy(false)
    }
  }

  async function handleDelete() {
    setBusy(true)
    setRowError('')
    try {
      await mangaApi.deleteChapter(mangaId, chapter.id)
      onDelete(chapter.id)
    } catch (err) {
      setRowError(err instanceof ApiError ? err.message : 'Delete failed')
      setBusy(false)
    }
  }

  async function handleZipUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setUploadName(file.name)
    setBusy(true)
    setRowError('')
    try {
      const result = await mangaApi.uploadPagesZip(mangaId, chapter.id, file)
      // Refresh chapter to get updated page_count
      const chs = await mangaApi.listChapters(mangaId)
      const updated = chs.find(c => c.id === chapter.id)
      if (updated) onUpdate(updated)
      setMode('idle')
      setUploadName('')
      if (result.metadata_suggestions) {
        onSuggestions(result.metadata_suggestions)
      }
    } catch (err) {
      setRowError(err instanceof ApiError ? err.message : 'Upload failed')
    } finally {
      setBusy(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  return (
    <motion.div
      layout
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0, height: 0 }}
      transition={{ duration: 0.18 }}
    >
      {/* Main row */}
      <div className="flex items-center justify-between px-4 py-3.5 gap-3">
        <div className="flex items-center gap-3 min-w-0">
          {!isOneshot && (
            <span className="text-xs font-mono text-jade-400 flex-shrink-0 w-16">
              Ch. {chapter.number}
            </span>
          )}
          <span className="text-sm text-mint-200 truncate">
            {chapter.title || (isOneshot ? 'Oneshot' : <span className="text-mint-200/30 italic">No title</span>)}
          </span>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          <span className="text-xs text-mint-200/30 hidden sm:block">{chapter.page_count}p</span>
          <ActionBtn
            active={mode === 'upload'}
            onClick={() => { setRowError(''); setMode(mode === 'upload' ? 'idle' : 'upload') }}
          >
            Upload
          </ActionBtn>
          <ActionBtn
            active={mode === 'edit'}
            onClick={() => mode === 'edit' ? setMode('idle') : openEdit()}
          >
            Edit
          </ActionBtn>
          <button
            onClick={() => { setRowError(''); setMode(mode === 'delete' ? 'idle' : 'delete') }}
            className={`text-xs px-2.5 py-1 rounded border transition ${
              mode === 'delete'
                ? 'bg-red-500/20 border-red-500/50 text-red-300'
                : 'border-forest-600 text-mint-200/40 hover:border-red-500/40 hover:text-red-400'
            }`}
          >
            Delete
          </button>
        </div>
      </div>

      {/* Expanded panels */}
      <AnimatePresence>
        {mode === 'edit' && (
          <motion.div
            key="edit"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.18 }}
            className="overflow-hidden border-t border-forest-700/60"
          >
            <div className="px-4 py-3 bg-forest-800/40 flex gap-3 flex-wrap items-end">
              {!isOneshot && (
                <div>
                  <label className="block text-xs text-mint-200/50 mb-1">Number</label>
                  <input
                    type="number"
                    step="0.1"
                    min="0"
                    value={editNumber}
                    onChange={e => setEditNumber(e.target.value)}
                    className="w-24 h-8 px-2.5 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 focus:outline-none focus:border-jade-500/60 transition"
                  />
                </div>
              )}
              <div className="flex-1 min-w-40">
                <label className="block text-xs text-mint-200/50 mb-1">Title</label>
                <input
                  type="text"
                  value={editTitle}
                  onChange={e => setEditTitle(e.target.value)}
                  className="w-full h-8 px-2.5 rounded-lg text-sm bg-forest-800 border border-forest-600
                             text-mint-50 focus:outline-none focus:border-jade-500/60 transition"
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleSave}
                  disabled={busy}
                  className="h-8 px-3 rounded-lg text-xs font-medium bg-jade-500 text-forest-950
                             hover:bg-jade-400 disabled:opacity-50 transition-colors"
                >
                  {busy ? 'Saving…' : 'Save'}
                </button>
                <button
                  onClick={() => setMode('idle')}
                  className="h-8 px-3 rounded-lg text-xs border border-forest-600
                             text-mint-200/50 hover:text-mint-50 transition"
                >
                  Cancel
                </button>
              </div>
            </div>
            {rowError && <ErrorBar msg={rowError} />}
          </motion.div>
        )}

        {mode === 'delete' && (
          <motion.div
            key="delete"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.18 }}
            className="overflow-hidden border-t border-forest-700/60"
          >
            <div className="px-4 py-3 bg-red-500/5 flex items-center justify-between gap-3">
              <p className="text-sm text-mint-200/70">
                Delete{isOneshot ? ' the oneshot chapter' : <> <span className="text-mint-50 font-medium">Ch. {chapter.number}</span></>}? This removes all pages too.
              </p>
              <div className="flex gap-2 flex-shrink-0">
                <button
                  onClick={handleDelete}
                  disabled={busy}
                  className="h-8 px-3 rounded-lg text-xs font-medium bg-red-500 text-white
                             hover:bg-red-400 disabled:opacity-50 transition-colors"
                >
                  {busy ? 'Deleting…' : 'Yes, delete'}
                </button>
                <button
                  onClick={() => setMode('idle')}
                  className="h-8 px-3 rounded-lg text-xs border border-forest-600
                             text-mint-200/50 hover:text-mint-50 transition"
                >
                  Cancel
                </button>
              </div>
            </div>
            {rowError && <ErrorBar msg={rowError} />}
          </motion.div>
        )}

        {mode === 'upload' && (
          <motion.div
            key="upload"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.18 }}
            className="overflow-hidden border-t border-forest-700/60"
          >
            <div className="px-4 py-3 bg-forest-800/40 flex items-center gap-3 flex-wrap">
              <p className="text-xs text-mint-200/50">
                Upload a ZIP of images — replaces all existing pages. Files sorted by filename.
              </p>
              <input
                ref={fileRef}
                type="file"
                accept=".zip,application/zip"
                onChange={handleZipUpload}
                disabled={busy}
                className="hidden"
              />
              <button
                onClick={() => fileRef.current?.click()}
                disabled={busy}
                className="h-8 px-3 rounded-lg text-xs font-medium border border-jade-500/40
                           text-jade-300 hover:bg-jade-500/10 disabled:opacity-50 transition flex-shrink-0"
              >
                {busy ? `Uploading ${uploadName}…` : 'Choose ZIP'}
              </button>
              {busy && <Spinner size="sm" />}
            </div>
            {rowError && <ErrorBar msg={rowError} />}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}

function ActionBtn({ active, onClick, children }: {
  active: boolean; onClick: () => void; children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      className={`text-xs px-2.5 py-1 rounded border transition ${
        active
          ? 'bg-jade-500/20 border-jade-500/50 text-jade-300'
          : 'border-forest-600 text-mint-200/40 hover:border-jade-500/40 hover:text-jade-300'
      }`}
    >
      {children}
    </button>
  )
}

function ErrorBar({ msg }: { msg: string }) {
  return (
    <p className="px-4 py-2 text-xs text-red-400 bg-red-500/10 border-t border-red-500/20">
      {msg}
    </p>
  )
}
