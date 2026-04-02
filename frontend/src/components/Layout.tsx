import { type ReactNode } from 'react'
import { motion } from 'framer-motion'
import { Navbar } from './Navbar'

export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col bg-forest-950">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:z-[100] focus:top-2 focus:left-2
                   focus:px-4 focus:py-2 focus:bg-jade-500 focus:text-forest-950 focus:rounded-md
                   focus:text-sm focus:font-medium"
      >
        Skip to main content
      </a>
      <Navbar />
      <motion.main
        id="main-content"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25 }}
        className="flex-1"
      >
        {children}
      </motion.main>
    </div>
  )
}
