import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { HomePage } from './pages/HomePage'
import { MangaDetailPage } from './pages/MangaDetailPage'
import { ReaderPage } from './pages/ReaderPage'
import { LoginPage } from './pages/LoginPage'
import { NotFoundPage } from './pages/NotFoundPage'
import { CreateMangaPage } from './pages/CreateMangaPage'
import { ManageChaptersPage } from './pages/ManageChaptersPage'

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/manga/new" element={<CreateMangaPage />} />
          <Route path="/manga/:mangaID" element={<MangaDetailPage />} />
          <Route path="/manga/:mangaID/manage" element={<ManageChaptersPage />} />
          <Route path="/manga/:mangaID/chapter/:chapterID" element={<ReaderPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
