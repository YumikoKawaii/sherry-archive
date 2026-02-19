import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuth } from '../contexts/AuthContext'

export function Navbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [search, setSearch] = useState('')
  const [menuOpen, setMenuOpen] = useState(false)

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    if (search.trim()) navigate(`/?q=${encodeURIComponent(search.trim())}`)
  }

  function handleLogout() {
    logout()
    navigate('/')
  }

  return (
    <header className="sticky top-0 z-50 border-b border-forest-700/60
                       bg-forest-950/80 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center gap-4">
        {/* Logo */}
        <Link to="/" className="flex-shrink-0 flex items-baseline gap-1.5 group">
          <span className="text-lg font-black tracking-widest text-jade-400
                           group-hover:text-jade-300 transition-colors">
            SHERRY
          </span>
          <span className="text-xs font-medium tracking-[0.2em] text-mint-200/50
                           group-hover:text-mint-200/80 transition-colors">
            ARCHIVE
          </span>
        </Link>

        {/* Search */}
        <form onSubmit={handleSearch} className="flex-1 max-w-md">
          <div className="relative">
            <input
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Search manga…"
              className="w-full h-8 pl-3 pr-9 rounded-md text-sm bg-forest-800
                         border border-forest-600 text-mint-50 placeholder-mint-200/30
                         focus:outline-none focus:border-jade-500/60 focus:ring-1
                         focus:ring-jade-500/30 transition"
            />
            <button type="submit"
              className="absolute right-2 top-1/2 -translate-y-1/2 text-mint-200/40
                         hover:text-jade-400 transition-colors">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round"
                  d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
              </svg>
            </button>
          </div>
        </form>

        {/* Nav links — desktop */}
        <nav className="hidden sm:flex items-center gap-1 text-sm">
          <NavLink to="/" active={location.pathname === '/'}>Browse</NavLink>
          {user && <NavLink to="/me" active={location.pathname === '/me'}>Bookmarks</NavLink>}
          {user && <NavLink to="/manga/new" active={location.pathname === '/manga/new'}>+ New</NavLink>}
        </nav>

        {/* Auth */}
        <div className="hidden sm:flex items-center gap-2 ml-auto flex-shrink-0">
          {user ? (
            <div className="flex items-center gap-3">
              <span className="text-sm text-jade-300 font-medium">{user.username}</span>
              <button onClick={handleLogout}
                className="text-xs px-3 py-1.5 rounded border border-forest-600
                           text-mint-200/60 hover:text-mint-50 hover:border-forest-500 transition">
                Sign out
              </button>
            </div>
          ) : (
            <Link to="/login"
              className="text-sm px-4 py-1.5 rounded-md font-medium
                         bg-jade-500 text-forest-950 hover:bg-jade-400 transition-colors">
              Sign in
            </Link>
          )}
        </div>

        {/* Mobile menu button */}
        <button onClick={() => setMenuOpen(v => !v)}
          className="sm:hidden ml-auto p-1.5 rounded text-mint-200/60 hover:text-mint-50 transition">
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            {menuOpen
              ? <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              : <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />}
          </svg>
        </button>
      </div>

      {/* Mobile menu */}
      <AnimatePresence>
        {menuOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.2 }}
            className="sm:hidden border-t border-forest-700/60 bg-forest-900/95 px-4 py-3 space-y-2"
          >
            <Link to="/" onClick={() => setMenuOpen(false)}
              className="block text-sm py-2 text-mint-200 hover:text-jade-400 transition">Browse</Link>
            {user
              ? <>
                  <Link to="/me" onClick={() => setMenuOpen(false)}
                    className="block text-sm py-2 text-mint-200 hover:text-jade-400 transition">Bookmarks</Link>
                  <Link to="/manga/new" onClick={() => setMenuOpen(false)}
                    className="block text-sm py-2 text-mint-200 hover:text-jade-400 transition">+ New Manga</Link>
                  <button onClick={() => { handleLogout(); setMenuOpen(false) }}
                    className="block text-sm py-2 text-mint-200/60 hover:text-mint-50 transition">Sign out</button>
                </>
              : <Link to="/login" onClick={() => setMenuOpen(false)}
                  className="block text-sm py-2 text-jade-400 font-medium">Sign in</Link>
            }
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  )
}

function NavLink({ to, active, children }: { to: string; active: boolean; children: React.ReactNode }) {
  return (
    <Link to={to}
      className={`px-3 py-1.5 rounded-md text-sm transition-colors ${
        active
          ? 'text-jade-400 bg-jade-500/10'
          : 'text-mint-200/60 hover:text-mint-50 hover:bg-forest-800'
      }`}>
      {children}
    </Link>
  )
}
