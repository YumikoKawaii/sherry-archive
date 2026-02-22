import { api } from './api'
import type { Manga, Chapter, ChapterWithPages, Comment } from '../types/manga'
import type { PagedData } from '../types/api'

export interface MangaFilters {
  q?: string
  status?: string
  'tags[]'?: string[]
  sort?: string
  page?: number
  limit?: number
  author?: string
  artist?: string
  category?: string
}

export function buildQuery(params: Record<string, unknown>): string {
  const parts: string[] = []
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === '' || v === null) continue
    if (Array.isArray(v)) {
      v.forEach(item => parts.push(`${encodeURIComponent(k)}=${encodeURIComponent(item)}`))
    } else {
      parts.push(`${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    }
  }
  return parts.length ? `?${parts.join('&')}` : ''
}

export interface CreateMangaPayload {
  title: string
  description?: string
  status: Manga['status']
  type: Manga['type']
  tags: string[]
  author?: string
  artist?: string
  category?: string
}

export interface ZipMetadataSuggestions {
  chapter_number?: number
  chapter_title?: string
  author?: string
  artist?: string
  tags?: string[]
  category?: string
  language?: string
}

export interface ZipUploadResult {
  pages: unknown[]
  metadata_suggestions?: ZipMetadataSuggestions
}

export interface OneshotUploadResult {
  chapter: import('../types/manga').Chapter
  pages: unknown[]
  metadata_suggestions?: ZipMetadataSuggestions
}

export const mangaApi = {
  list: (filters: MangaFilters = {}) =>
    api.get<PagedData<Manga>>(`/mangas${buildQuery(filters as Record<string, unknown>)}`),

  get: (id: string) =>
    api.get<Manga>(`/mangas/${id}`),

  create: (payload: CreateMangaPayload) =>
    api.post<Manga>('/mangas', payload),

  update: (mangaId: string, payload: Partial<Omit<CreateMangaPayload, 'title'> & { title: string }>) =>
    api.patch<Manga>(`/mangas/${mangaId}`, payload),

  uploadCover: (mangaId: string, file: File) => {
    const form = new FormData()
    form.append('cover', file)
    return api.putForm<Manga>(`/mangas/${mangaId}/cover`, form)
  },

  listChapters: (mangaId: string) =>
    api.get<Chapter[]>(`/mangas/${mangaId}/chapters`),

  getChapter: (mangaId: string, chapterId: string) =>
    api.get<ChapterWithPages>(`/mangas/${mangaId}/chapters/${chapterId}`),

  createChapter: (mangaId: string, payload: { number?: number; title?: string }) =>
    api.post<Chapter>(`/mangas/${mangaId}/chapters`, payload),

  updateChapter: (mangaId: string, chapterId: string, payload: { number?: number; title?: string }) =>
    api.patch<Chapter>(`/mangas/${mangaId}/chapters/${chapterId}`, payload),

  deleteChapter: (mangaId: string, chapterId: string) =>
    api.delete<void>(`/mangas/${mangaId}/chapters/${chapterId}`),

  uploadPagesZip: (mangaId: string, chapterId: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.postForm<ZipUploadResult>(`/mangas/${mangaId}/chapters/${chapterId}/pages/zip`, form)
  },

  oneshotUpload: (mangaId: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.postForm<OneshotUploadResult>(`/mangas/${mangaId}/oneshot/upload`, form)
  },

  listByUser: (userId: string, page = 1) =>
    api.get<PagedData<Manga>>(`/users/${userId}/mangas?page=${page}`),

  listMangaComments: (mangaId: string, page = 1) =>
    api.get<PagedData<Comment>>(`/mangas/${mangaId}/comments?page=${page}`),

  createMangaComment: (mangaId: string, content: string) =>
    api.post<Comment>(`/mangas/${mangaId}/comments`, { content }),

  listChapterComments: (mangaId: string, chapterId: string, page = 1) =>
    api.get<PagedData<Comment>>(`/mangas/${mangaId}/chapters/${chapterId}/comments?page=${page}`),

  createChapterComment: (mangaId: string, chapterId: string, content: string) =>
    api.post<Comment>(`/mangas/${mangaId}/chapters/${chapterId}/comments`, { content }),

  updateComment: (mangaId: string, commentId: string, content: string) =>
    api.patch<Comment>(`/mangas/${mangaId}/comments/${commentId}`, { content }),

  deleteComment: (mangaId: string, commentId: string) =>
    api.delete<void>(`/mangas/${mangaId}/comments/${commentId}`),
}

export interface TrendingItem extends Manga {
  trending_score: number
}

export const analyticsApi = {
  trending: (limit = 12) =>
    api.get<{ data: TrendingItem[] }>(`/analytics/trending?limit=${limit}`),

  suggestions: (deviceId: string, limit = 12) =>
    api.get<{ data: Manga[] }>(`/analytics/suggestions?device_id=${encodeURIComponent(deviceId)}&limit=${limit}`),
}
