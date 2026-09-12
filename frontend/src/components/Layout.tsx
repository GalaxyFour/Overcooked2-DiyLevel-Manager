import { Link, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuth } from '../context/AuthContext'
import { useTheme } from '../context/ThemeContext'

export default function Layout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth()
  const { theme, toggle } = useTheme()
  const loc = useLocation()

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-50 backdrop-blur-md bg-cream/80 dark:bg-kitchen-dark/80 border-b border-orange-100/50 dark:border-white/5">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between gap-4">
          <Link to="/" className="flex items-center gap-2 group">
            <motion.span
              className="text-3xl"
              whileHover={{ rotate: [0, -10, 10, 0] }}
              transition={{ duration: 0.4 }}
            >
              🍳
            </motion.span>
            <span className="font-extrabold text-lg text-tomato group-hover:text-tomato/80 transition">
              OC2 关卡管理器
            </span>
          </Link>

          <nav className="flex items-center gap-1 sm:gap-2">
            <NavLink to="/" active={loc.pathname === '/'}>首页</NavLink>
            {user && (user.role === 'author' || user.role === 'super_admin' || user.role === 'admin') && (
              <NavLink to="/dashboard" active={loc.pathname.startsWith('/dashboard')}>工作台</NavLink>
            )}
            {(user?.role === 'admin' || user?.role === 'super_admin') && (
              <NavLink to="/admin" active={loc.pathname.startsWith('/admin')}>管理</NavLink>
            )}
            <button onClick={toggle} className="btn-ghost text-lg" title="切换主题">
              {theme === 'light' ? '🌙' : '☀️'}
            </button>
            {user ? (
              <button onClick={logout} className="btn-ghost text-sm">
                {user.displayName || user.username} · 退出
              </button>
            ) : (
              <Link to="/login" className="btn-primary text-sm">登录</Link>
            )}
          </nav>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">{children}</main>

      <footer className="text-center py-8 text-sm opacity-60">
        <p>🥘 Overcooked 2 DIY Level Manager · 胡闹厨房自定义关卡</p>
      </footer>
    </div>
  )
}

function NavLink({ to, children, active }: { to: string; children: React.ReactNode; active: boolean }) {
  return (
    <Link
      to={to}
      className={`px-3 py-1.5 rounded-full text-sm font-semibold transition ${
        active
          ? 'bg-tomato/15 text-tomato dark:text-tomato'
          : 'text-[var(--muted)] hover:bg-orange-50 dark:hover:bg-white/5'
      }`}
    >
      {children}
    </Link>
  )
}
