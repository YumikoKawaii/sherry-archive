import { api } from './api'
import type { Manga, Chapter, ChapterWithPages } from '../types/manga'
import type { PagedData } from '../types/api'

export interface MangaFilters {
  q?: string
  status?: string
  'tags[]'?: string[]
  sort?: string
  page?: number
  limit?: number
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

export const mangaApi = {
  list: (filters: MangaFilters = {}) =>
    api.get<PagedData<Manga>>(`/mangas${buildQuery(filters as Record<string, unknown>)}`),

  get: (id: string) =>
    api.get<Manga>(`/mangas/${id}`),

  listChapters: (mangaId: string) =>
    api.get<Chapter[]>(`/mangas/${mangaId}/chapters`),

  getChapter: (mangaId: string, chapterId: string) =>
    api.get<ChapterWithPages>(`/mangas/${mangaId}/chapters/${chapterId}`),

  listByUser: (userId: string, page = 1) =>
    api.get<PagedData<Manga>>(`/users/${userId}/mangas?page=${page}`),
}
