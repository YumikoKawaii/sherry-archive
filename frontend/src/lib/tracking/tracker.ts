import { getDeviceId } from './device'

type Properties = Record<string, unknown>

class Tracker {
  private send(event: string, properties: Properties = {}): void {
    const token = localStorage.getItem('access_token')
    const payload = {
      events: [{
        device_id: getDeviceId(),
        event,
        properties,
        referrer: document.referrer,
      }],
    }
    fetch('/api/track', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(payload),
    }).catch(() => {}) // silent — tracking must never break the app
  }

  // --- Discovery ---

  mangaView(props: { manga_id: string; manga_type: string }): void {
    this.send('manga_view', props)
  }

  search(props: { query: string; filters: Properties; result_count: number }): void {
    this.send('search', props)
  }

  tagClick(props: { tag: string }): void {
    this.send('tag_click', props)
  }

  // --- Reading ---

  chapterOpen(props: { manga_id: string; chapter_id: string; chapter_number: number }): void {
    this.send('chapter_open', props)
  }

  chapterComplete(props: { manga_id: string; chapter_id: string; duration_seconds: number }): void {
    this.send('chapter_complete', props)
  }

  chapterNavigate(props: { from_chapter_id: string; to_chapter_id: string; direction: 'prev' | 'next' }): void {
    this.send('chapter_navigate', props)
  }

  // --- Social ---

  commentPost(props: { manga_id: string; chapter_id?: string; content_length: number }): void {
    this.send('comment_post', props)
  }

  // --- Library ---

  bookmarkAdd(props: { manga_id: string }): void {
    this.send('bookmark_add', props)
  }

  bookmarkRemove(props: { manga_id: string }): void {
    this.send('bookmark_remove', props)
  }

  // --- Auth ---

  signup(): void { this.send('signup') }
  login(): void  { this.send('login') }
}

export const tracker = new Tracker()
