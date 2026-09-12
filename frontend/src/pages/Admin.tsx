import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'

export default function Admin() {
  const { user } = useAuth()
  const isSuper = user?.role === 'super_admin'

  const { data: adminSets = [], refetch: refetchSets } = useQuery({
    queryKey: ['admin-sets'],
    queryFn: () => api.adminSets(),
  })

  const { data: users = [], refetch: refetchUsers } = useQuery({
    queryKey: ['users'],
    queryFn: api.listUsers,
    enabled: isSuper,
  })

  const { data: invalidSets = [] } = useQuery({
    queryKey: ['invalid-sets'],
    queryFn: api.listInvalidSets,
    enabled: isSuper,
  })

  const [newUser, setNewUser] = useState({ username: '', password: '', displayName: '', role: 'author' })

  const createUser = async () => {
    await api.createUser(newUser)
    setNewUser({ username: '', password: '', displayName: '', role: 'author' })
    refetchUsers()
  }

  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-extrabold">管理后台 🛡️</h1>

      <section className="card p-6">
        <h2 className="font-bold text-lg mb-4">全部关卡集</h2>
        <div className="space-y-2">
          {adminSets.map((s) => (
            <div key={s.id} className="flex items-center justify-between p-3 rounded-xl bg-orange-50/50 dark:bg-white/5">
              <div>
                <span className="font-semibold">{s.nameZh || s.slug}</span>
                <span className="text-sm opacity-50 ml-2">by {s.authorName}</span>
              </div>
              <div className="flex gap-2">
                {s.status !== 'active' && (
                  <button
                    onClick={() => api.adminPatchStatus(s.id, 'active').then(() => refetchSets())}
                    className="btn-ghost text-xs"
                  >
                    恢复
                  </button>
                )}
                {s.status !== 'hidden' && (
                  <button
                    onClick={() => api.adminPatchStatus(s.id, 'hidden').then(() => refetchSets())}
                    className="btn-ghost text-xs"
                  >
                    隐藏
                  </button>
                )}
                {isSuper && s.status !== 'invalid' && (
                  <button
                    onClick={() => api.invalidateSet(s.id).then(() => refetchSets())}
                    className="btn-ghost text-xs text-tomato"
                  >
                    标记无效
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </section>

      {isSuper && (
        <>
          <section className="card p-6">
            <h2 className="font-bold text-lg mb-4">创建账号（申请制）</h2>
            <div className="grid sm:grid-cols-2 gap-3 mb-4">
              <input className="input" placeholder="用户名" value={newUser.username} onChange={(e) => setNewUser({ ...newUser, username: e.target.value })} />
              <input className="input" placeholder="初始密码" value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} />
              <input className="input" placeholder="显示名" value={newUser.displayName} onChange={(e) => setNewUser({ ...newUser, displayName: e.target.value })} />
              <select className="input" value={newUser.role} onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}>
                <option value="author">作者</option>
                <option value="admin">管理员</option>
                <option value="super_admin">超级管理员</option>
              </select>
            </div>
            <button onClick={createUser} className="btn-primary">创建账号</button>
          </section>

          <section className="card p-6">
            <h2 className="font-bold text-lg mb-4">用户列表</h2>
            <div className="space-y-2">
              {users.map((u) => (
                <div key={u.id} className="flex justify-between p-3 rounded-xl bg-orange-50/50 dark:bg-white/5">
                  <span>{u.displayName || u.username} <span className="opacity-50 text-sm">({u.role})</span></span>
                  <button
                    onClick={() => api.patchUser(u.id, { role: u.role, displayName: u.displayName, disabled: true }).then(() => refetchUsers())}
                    className="btn-ghost text-xs text-tomato"
                  >
                    禁用
                  </button>
                </div>
              ))}
            </div>
          </section>

          <section className="card p-6">
            <h2 className="font-bold text-lg mb-4">无效集合 ({invalidSets.length})</h2>
            {invalidSets.length === 0 ? (
              <p className="opacity-50">暂无</p>
            ) : (
              invalidSets.map((s) => (
                <div key={s.id} className="p-2 opacity-70">{s.slug} — {s.nameZh}</div>
              ))
            )}
          </section>
        </>
      )}
    </div>
  )
}
