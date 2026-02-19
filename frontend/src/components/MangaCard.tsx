import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import type { Manga } from '../types/manga'
import { StatusBadge } from './StatusBadge'

interface Props {
  manga: Manga
}

export function MangaCard({ manga }: Props) {
  return (
    <motion.div
      whileHover={{ y: -4, scale: 1.02 }}
      transition={{ type: 'spring', stiffness: 320, damping: 24 }}
    >
      <Link to={`/manga/${manga.id}`} className="group block">
        <div className="relative rounded-lg overflow-hidden border border-forest-700
                        group-hover:border-jade-500/50 transition-colors duration-300
                        group-hover:shadow-[0_0_20px_rgba(34,197,94,0.12)]">
          {/* Cover */}
          <div className="aspect-[3/4] bg-forest-800 overflow-hidden">
            {manga.cover_url ? (
              <img
                src={manga.cover_url}
                alt={manga.title}
                className="w-full h-full object-cover transition-transform duration-500
                           group-hover:scale-105"
                loading="lazy"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center bg-forest-800">
                <span className="text-4xl opacity-20 select-none">漫</span>
              </div>
            )}

            {/* Gradient overlay — always present, stronger on hover */}
            <div className="absolute inset-0 bg-gradient-to-t from-forest-950/95 via-forest-950/20 to-transparent" />
          </div>

          {/* Status badge */}
          <div className="absolute top-2 right-2">
            <StatusBadge status={manga.status} />
          </div>

          {/* Title area */}
          <div className="absolute bottom-0 left-0 right-0 p-3">
            <h3 className="text-sm font-semibold text-mint-50 line-clamp-2 leading-tight">
              {manga.title}
            </h3>
            {manga.tags.length > 0 && (
              <p className="mt-1 text-xs text-jade-400/70 truncate">
                {manga.tags.slice(0, 3).join(' · ')}
              </p>
            )}
          </div>
        </div>
      </Link>
    </motion.div>
  )
}
