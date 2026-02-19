import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuth } from '../contexts/AuthContext'
import { ApiError } from '../lib/api'

type Tab = 'login' | 'register'

export function LoginPage() {
  const [tab, setTab] = useState<Tab>('login')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login, register } = useAuth()
  const navigate = useNavigate()

  // Login form state
  const [loginEmail, setLoginEmail] = useState('')
  const [loginPassword, setLoginPassword] = useState('')

  // Register form state
  const [regUsername, setRegUsername] = useState('')
  const [regEmail, setRegEmail] = useState('')
  const [regPassword, setRegPassword] = useState('')

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(loginEmail, loginPassword)
      navigate('/')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await register(regUsername, regEmail, regPassword)
      navigate('/')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-forest-950 bg-grid px-4">
      <div className="absolute inset-0 bg-gradient-to-b from-jade-500/5 to-transparent pointer-events-none" />

      <motion.div
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35 }}
        className="relative w-full max-w-sm"
      >
        {/* Logo */}
        <Link to="/" className="flex justify-center mb-8">
          <span className="text-2xl font-black tracking-widest text-jade-400">SHERRY</span>
          <span className="ml-1.5 self-end text-xs font-medium tracking-[0.2em] text-mint-200/40">ARCHIVE</span>
        </Link>

        <div className="bg-forest-900 border border-forest-700 rounded-xl overflow-hidden
                        shadow-[0_0_40px_rgba(34,197,94,0.06)]">
          {/* Tabs */}
          <div className="flex border-b border-forest-700">
            {(['login', 'register'] as Tab[]).map(t => (
              <button
                key={t}
                onClick={() => { setTab(t); setError('') }}
                className={`flex-1 py-3.5 text-sm font-medium capitalize transition-colors ${
                  tab === t
                    ? 'text-jade-400 border-b-2 border-jade-500 -mb-px'
                    : 'text-mint-200/40 hover:text-mint-200/70'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          <div className="p-6">
            <AnimatePresence mode="wait">
              {tab === 'login' ? (
                <motion.form
                  key="login"
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: 10 }}
                  transition={{ duration: 0.18 }}
                  onSubmit={handleLogin}
                  className="space-y-4"
                >
                  <Field label="Email" type="email" value={loginEmail}
                    onChange={setLoginEmail} placeholder="you@example.com" />
                  <Field label="Password" type="password" value={loginPassword}
                    onChange={setLoginPassword} placeholder="••••••••" />
                  {error && <ErrorMsg msg={error} />}
                  <SubmitBtn loading={loading}>Sign in</SubmitBtn>
                </motion.form>
              ) : (
                <motion.form
                  key="register"
                  initial={{ opacity: 0, x: 10 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -10 }}
                  transition={{ duration: 0.18 }}
                  onSubmit={handleRegister}
                  className="space-y-4"
                >
                  <Field label="Username" type="text" value={regUsername}
                    onChange={setRegUsername} placeholder="your_username" />
                  <Field label="Email" type="email" value={regEmail}
                    onChange={setRegEmail} placeholder="you@example.com" />
                  <Field label="Password" type="password" value={regPassword}
                    onChange={setRegPassword} placeholder="min 8 characters" />
                  {error && <ErrorMsg msg={error} />}
                  <SubmitBtn loading={loading}>Create account</SubmitBtn>
                </motion.form>
              )}
            </AnimatePresence>
          </div>
        </div>
      </motion.div>
    </div>
  )
}

function Field({ label, type, value, onChange, placeholder }: {
  label: string; type: string; value: string
  onChange: (v: string) => void; placeholder: string
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-mint-200/60 mb-1.5">{label}</label>
      <input
        type={type}
        value={value}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder}
        required
        className="w-full h-10 px-3 rounded-lg text-sm bg-forest-800 border border-forest-600
                   text-mint-50 placeholder-mint-200/20
                   focus:outline-none focus:border-jade-500/60 focus:ring-1 focus:ring-jade-500/25
                   transition"
      />
    </div>
  )
}

function ErrorMsg({ msg }: { msg: string }) {
  return (
    <p className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
      {msg}
    </p>
  )
}

function SubmitBtn({ loading, children }: { loading: boolean; children: React.ReactNode }) {
  return (
    <button
      type="submit"
      disabled={loading}
      className="w-full h-10 rounded-lg text-sm font-semibold bg-jade-500 text-forest-950
                 hover:bg-jade-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors mt-2"
    >
      {loading ? 'Loading…' : children}
    </button>
  )
}
