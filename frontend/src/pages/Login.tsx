import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [mustChange, setMustChange] = useState(false)
  const [error, setError] = useState('')
  const { login, refresh } = useAuth()
  const navigate = useNavigate()

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const res = await login(username, password)
      if (res.mustChangePassword) {
        setMustChange(true)
        return
      }
      navigate('/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal masuk')
    }
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await api.changePassword(newPassword)
      await refresh()
      navigate('/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal mengubah kata sandi')
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-md mx-auto card p-8"
    >
      <div className="text-center mb-8">
        <div className="text-5xl mb-3">👨‍🍳</div>
        <h1 className="text-2xl font-extrabold">{mustChange ? 'Atur Kata Sandi Baru' : 'Masuk Penulis'}</h1>
      </div>

      {mustChange ? (
        <form onSubmit={handleChangePassword} className="space-y-4">
          <p className="text-sm opacity-70 text-center mb-4">Silakan atur kata sandi untuk login pertama kali (minimal 6 karakter)</p>
          <input
            type="password"
            className="input"
            placeholder="Kata sandi baru"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={6}
            required
          />
          <button type="submit" className="btn-primary w-full">Konfirmasi</button>
        </form>
      ) : (
        <form onSubmit={handleLogin} className="space-y-4">
          <input
            className="input"
            placeholder="Nama pengguna"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <input
            type="password"
            className="input"
            placeholder="Kata sandi"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <button type="submit" className="btn-primary w-full">Masuk</button>
        </form>
      )}

      {error && <p className="text-tomato text-sm text-center mt-4">{error}</p>}
    </motion.div>
  )
}
