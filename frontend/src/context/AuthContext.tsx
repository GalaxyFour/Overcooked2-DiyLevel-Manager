import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { api, User } from '../api/client'

interface AuthState {
  user: User | null
  loading: boolean
  login: (username: string, password: string) => Promise<{ mustChangePassword: boolean }>
  logout: () => void
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    try {
      const u = await api.me()
      setUser(u)
    } catch {
      setUser(null)
    }
  }

  useEffect(() => {
    refresh().finally(() => setLoading(false))
  }, [])

  const login = async (username: string, password: string) => {
    const res = await api.login(username, password)
    setUser(res.user)
    return { mustChangePassword: res.mustChangePassword }
  }

  const logout = () => setUser(null)

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refresh }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth outside provider')
  return ctx
}
