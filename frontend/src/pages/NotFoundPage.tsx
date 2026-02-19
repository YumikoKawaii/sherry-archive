import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'

export function NotFoundPage() {
  return (
    <div className="min-h-screen bg-forest-950 bg-grid flex flex-col items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="text-center"
      >
        <p className="text-8xl font-black text-jade-500/10 mb-2 select-none">404</p>
        <p className="text-6xl mb-6 select-none">迷</p>
        <h1 className="text-2xl font-bold text-mint-50 mb-2">Page not found</h1>
        <p className="text-mint-200/40 mb-8 text-sm">The page you're looking for doesn't exist.</p>
        <Link to="/"
          className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-semibold
                     bg-jade-500 text-forest-950 hover:bg-jade-400 transition-colors">
          ← Go home
        </Link>
      </motion.div>
    </div>
  )
}
