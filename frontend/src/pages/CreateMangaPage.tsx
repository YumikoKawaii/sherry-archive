import { useState, useRef } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { mangaApi } from '../lib/manga'
import { ApiError } from '../lib/api'
import { Layout } from '../components/Layout'
import type { MangaStatus, MangaType } from '../types/manga'

const STATUSES: { value: MangaStatus; label: string }[] = [
  { value: 'ongoing', label: 'Ongoing' },
  { value: 'completed', label: 'Completed' },
  { value: 'hiatus', label: 'Hiatus' },
]

const TYPES: { value: MangaType; label: string }[] = [
  { value: 'series', label: 'Series' },
  { value: 'oneshot', label: 'Oneshot' },
]

export function CreateMangaPage() {
  const navigate = useNavigate()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<MangaStatus>('ongoing')
  const [type, setType] = useState<MangaType>('series')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [author, setAuthor] = useState('')
  const [artist, setArtist] = useState('')
  const [category, setCategory] = useState('')
  const [coverFile, setCoverFile] = useState<File | null>(null)
  const [coverPreview, setCoverPreview] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const fileRef = useRef<HTMLInputElement>(null)

  function handleCoverChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setCoverFile(file)
    setCoverPreview(URL.createObjectURL(file))
  }

  function addTag() {
    const t = tagInput.trim().toLowerCase()
    if (t && !tags.includes(t)) setTags(prev => [...prev, t])
    setTagInput('')
  }

  function handleTagKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      addTag()
    }
  }

  function removeTag(tag: string) {
    setTags(prev => prev.filter(t => t !== tag))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const manga = await mangaApi.create({ title, description, status, type, tags, author, artist, category })
      if (coverFile) {
        await mangaApi.uploadCover(manga.id, coverFile)
      }
      navigate(`/manga/${manga.id}`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to create manga')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Layout>
      <div className="max-w-2xl mx-auto px-4 sm:px-6 py-10">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          {/* Header */}
          <div className="mb-8">
            <Link to="/" className="text-xs text-mint-200/40 hover:text-mint-200/70 transition mb-3 inline-block">
              ← Back to browse
            </Link>
            <h1 className="text-2xl font-bold text-mint-50 flex items-center gap-2">
              <span className="w-1 h-6 rounded-full bg-jade-500 inline-block" />
              New Manga
            </h1>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Cover + Title row */}
            <div className="flex gap-6">
              {/* Cover upload */}
              <div className="flex-shrink-0">
                <p className="text-xs font-medium text-mint-200/60 mb-1.5">Cover</p>
                <button
                  type="button"
                  onClick={() => fileRef.current?.click()}
                  className="w-28 aspect-[3/4] rounded-lg overflow-hidden border border-dashed border-forest-600
                             hover:border-jade-500/60 transition group relative bg-forest-800"
                >
                  {coverPreview ? (
                    <img src={coverPreview} alt="Cover preview" className="w-full h-full object-cover" />
                  ) : (
                    <div className="w-full h-full flex flex-col items-center justify-center gap-1.5
                                    text-mint-200/20 group-hover:text-mint-200/50 transition">
                      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round"
                          d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25
                             2.25 0 013.182 0l2.909 2.909M3 21h18M6.75 6.75h.008v.008H6.75V6.75z" />
                      </svg>
                      <span className="text-[10px]">Add cover</span>
                    </div>
                  )}
                  {coverPreview && (
                    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100
                                    transition flex items-center justify-center">
                      <span className="text-[10px] text-white font-medium">Change</span>
                    </div>
                  )}
                </button>
                <input
                  ref={fileRef}
                  type="file"
                  accept="image/*"
                  onChange={handleCoverChange}
                  className="hidden"
                />
              </div>

              {/* Title + Type + Status */}
              <div className="flex-1 space-y-4">
                <div>
                  <label className="block text-xs font-medium text-mint-200/60 mb-1.5">
                    Title <span className="text-jade-400">*</span>
                  </label>
                  <input
                    type="text"
                    value={title}
                    onChange={e => setTitle(e.target.value)}
                    placeholder="Manga title"
                    required
                    className="w-full h-10 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                               text-mint-50 placeholder-mint-200/20
                               focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25
                               transition"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Type</label>
                  <div className="flex gap-2">
                    {TYPES.map(t => (
                      <button
                        key={t.value}
                        type="button"
                        onClick={() => setType(t.value)}
                        className={`flex-1 h-9 rounded-lg text-xs font-medium transition border ${
                          type === t.value
                            ? 'bg-jade-500/20 border-jade-500/60 text-jade-300'
                            : 'bg-forest-800 border-forest-600 text-mint-200/50 hover:border-forest-500 hover:text-mint-200/80'
                        }`}
                      >
                        {t.label}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Status</label>
                  <div className="flex gap-2">
                    {STATUSES.map(s => (
                      <button
                        key={s.value}
                        type="button"
                        onClick={() => setStatus(s.value)}
                        className={`flex-1 h-9 rounded-lg text-xs font-medium transition border ${
                          status === s.value
                            ? 'bg-jade-500/20 border-jade-500/60 text-jade-300'
                            : 'bg-forest-800 border-forest-600 text-mint-200/50 hover:border-forest-500 hover:text-mint-200/80'
                        }`}
                      >
                        {s.label}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Author / Artist / Category */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Author</label>
                <input
                  type="text"
                  value={author}
                  onChange={e => setAuthor(e.target.value)}
                  placeholder="Author name"
                  className="w-full h-10 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                             text-mint-50 placeholder-mint-200/20
                             focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25
                             transition"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Artist</label>
                <input
                  type="text"
                  value={artist}
                  onChange={e => setArtist(e.target.value)}
                  placeholder="Artist name"
                  className="w-full h-10 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                             text-mint-50 placeholder-mint-200/20
                             focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25
                             transition"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Category</label>
                <input
                  type="text"
                  value={category}
                  onChange={e => setCategory(e.target.value)}
                  placeholder="e.g. doujinshi"
                  className="w-full h-10 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                             text-mint-50 placeholder-mint-200/20
                             focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25
                             transition"
                />
              </div>
            </div>

            {/* Description */}
            <div>
              <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Description</label>
              <textarea
                value={description}
                onChange={e => setDescription(e.target.value)}
                placeholder="Write a short synopsis…"
                rows={4}
                className="w-full px-3 py-2.5 rounded-lg text-sm bg-forest-800 border border-forest-600
                           text-mint-50 placeholder-mint-200/20 resize-none
                           focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25
                           transition"
              />
            </div>

            {/* Tags */}
            <div>
              <label className="block text-xs font-medium text-mint-200/60 mb-1.5">Tags</label>
              <div className="rounded-lg bg-forest-800 border border-forest-600
                              focus-within:border-jade-500/60 focus-within:ring-1 focus-within:ring-jade-500/25
                              transition p-2 min-h-[42px] flex flex-wrap gap-1.5 items-center">
                {tags.map(tag => (
                  <span key={tag}
                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs
                               bg-jade-500/15 text-jade-300 border border-jade-500/25">
                    {tag}
                    <button type="button" onClick={() => removeTag(tag)}
                      className="text-jade-400/60 hover:text-jade-300 transition leading-none">×</button>
                  </span>
                ))}
                <input
                  type="text"
                  value={tagInput}
                  onChange={e => setTagInput(e.target.value)}
                  onKeyDown={handleTagKeyDown}
                  onBlur={addTag}
                  placeholder={tags.length === 0 ? 'Type a tag, press Enter…' : ''}
                  className="flex-1 min-w-[120px] bg-transparent text-sm text-mint-50
                             placeholder-mint-200/20 outline-none"
                />
              </div>
              <p className="mt-1 text-[10px] text-mint-200/30">Press Enter or comma to add a tag</p>
            </div>

            {/* Error */}
            {error && (
              <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
                {error}
              </p>
            )}

            {/* Actions */}
            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={loading}
                className="px-6 h-10 rounded-lg text-sm font-semibold bg-jade-500 text-forest-950
                           hover:bg-jade-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {loading ? 'Creating…' : 'Create Manga'}
              </button>
              <Link
                to="/"
                className="px-6 h-10 rounded-lg text-sm font-medium border border-forest-600
                           text-mint-200/60 hover:text-mint-50 hover:border-forest-500 transition
                           inline-flex items-center"
              >
                Cancel
              </Link>
            </div>
          </form>
        </motion.div>
      </div>
    </Layout>
  )
}
