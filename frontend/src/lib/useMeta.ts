import { useEffect } from 'react'

interface MetaOptions {
  title?: string
  description?: string
  ogImage?: string
  ogUrl?: string
  ogType?: string
}

const SITE_NAME = 'Sherry Archive'
const DEFAULT_DESC = 'Discover and read manga on Sherry Archive. Browse trending titles, get personalized recommendations, and dive into your next favorite story.'

export function useMeta({ title, description, ogImage, ogUrl, ogType = 'website' }: MetaOptions) {
  useEffect(() => {
    const fullTitle = title ? `${title} — ${SITE_NAME}` : SITE_NAME
    const desc = description || DEFAULT_DESC

    document.title = fullTitle
    setMeta('name', 'description', desc)
    setMeta('property', 'og:site_name', SITE_NAME)
    setMeta('property', 'og:type', ogType)
    setMeta('property', 'og:title', fullTitle)
    setMeta('property', 'og:description', desc)
    setMeta('property', 'og:url', ogUrl ?? window.location.href)
    if (ogImage) setMeta('property', 'og:image', ogImage)

    return () => {
      document.title = SITE_NAME
    }
  }, [title, description, ogImage, ogUrl, ogType])
}

function setMeta(attrKey: string, attrValue: string, content: string) {
  let el = document.querySelector(`meta[${attrKey}="${attrValue}"]`) as HTMLMetaElement | null
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attrKey, attrValue)
    document.head.appendChild(el)
  }
  el.content = content
}
