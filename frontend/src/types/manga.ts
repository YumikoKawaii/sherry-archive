export type MangaStatus = 'ongoing' | 'completed' | 'hiatus'

export interface Manga {
  id: string
  owner_id: string
  title: string
  slug: string
  description: string
  cover_url: string
  status: MangaStatus
  tags: string[]
  created_at: string
  updated_at: string
}

export interface Chapter {
  id: string
  manga_id: string
  number: number
  title: string
  page_count: number
  created_at: string
  updated_at: string
}

export interface PageItem {
  id: string
  number: number
  url: string
  width: number
  height: number
}

export interface ChapterWithPages {
  chapter: Chapter
  pages: PageItem[]
}

export interface Bookmark {
  id: string
  manga_id: string
  chapter_id: string
  last_page_number: number
  created_at: string
  updated_at: string
}
