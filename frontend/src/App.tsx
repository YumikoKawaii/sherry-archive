import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { HomePage } from './pages/HomePage'
import { FullPageSpinner } from './components/Spinner'

// All non-landing routes are lazy-loaded so the initial bundle only includes
// what's needed to render the homepage.
const MangaDetailPage   = lazy(() => import('./pages/MangaDetailPage').then(m => ({ default: m.MangaDetailPage })))
const ReaderPage        = lazy(() => import('./pages/ReaderPage').then(m => ({ default: m.ReaderPage })))
const LoginPage         = lazy(() => import('./pages/LoginPage').then(m => ({ default: m.LoginPage })))
const CreateMangaPage   = lazy(() => import('./pages/CreateMangaPage').then(m => ({ default: m.CreateMangaPage })))
const ManageChaptersPage = lazy(() => import('./pages/ManageChaptersPage').then(m => ({ default: m.ManageChaptersPage })))
const BookmarksPage     = lazy(() => import('./pages/BookmarksPage').then(m => ({ default: m.BookmarksPage })))
const NotFoundPage      = lazy(() => import('./pages/NotFoundPage').then(m => ({ default: m.NotFoundPage })))

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Suspense fallback={<FullPageSpinner />}>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/manga/new" element={<CreateMangaPage />} />
            <Route path="/manga/:mangaID" element={<MangaDetailPage />} />
            <Route path="/manga/:mangaID/manage" element={<ManageChaptersPage />} />
            <Route path="/manga/:mangaID/chapter/:chapterID" element={<ReaderPage />} />
            <Route path="/me" element={<BookmarksPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Suspense>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
