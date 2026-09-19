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
      <h1 className="text-3xl font-extrabold">Panel Admin 🛡️</h1>

      <section className="card p-6">
        <h2 className="font-bold text-lg mb-4">Semua Set Level</h2>
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
                    Pulihkan
                  </button>
                )}
                {s.status !== 'hidden' && (
                  <button
                    onClick={() => api.adminPatchStatus(s.id, 'hidden').then(() => refetchSets())}
                    className="btn-ghost text-xs"
                  >
                    Sembunyikan
                  </button>
                )}
                {isSuper && s.status !== 'invalid' && (
                  <button
                    onClick={() => api.invalidateSet(s.id).then(() => refetchSets())}
                    className="btn-ghost text-xs text-tomato"
                  >
                    Tandai Tidak Valid
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
            <h2 className="font-bold text-lg mb-4">Buat Akun (Sistem Pengajuan)</h2>
            <div className="grid sm:grid-cols-2 gap-3 mb-4">
              <input className="input" placeholder="Nama pengguna" value={newUser.username} onChange={(e) => setNewUser({ ...newUser, username: e.target.value })} />
              <input className="input" placeholder="Kata sandi awal" value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} />
              <input className="input" placeholder="Nama tampilan" value={newUser.displayName} onChange={(e) => setNewUser({ ...newUser, displayName: e.target.value })} />
              <select className="input" value={newUser.role} onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}>
                <option value="author">Penulis</option>
                <option value="admin">Admin</option>
                <option value="super_admin">Super Admin</option>
              </select>
            </div>
            <button onClick={createUser} className="btn-primary">Buat Akun</button>
          </section>

          <section className="card p-6">
            <h2 className="font-bold text-lg mb-4">Daftar Pengguna</h2>
            <div className="space-y-2">
              {users.map((u) => (
                <div key={u.id} className="flex justify-between p-3 rounded-xl bg-orange-50/50 dark:bg-white/5">
                  <span>{u.displayName || u.username} <span className="opacity-50 text-sm">({u.role})</span></span>
                  <button
                    onClick={() => api.patchUser(u.id, { role: u.role, displayName: u.displayName, disabled: true }).then(() => refetchUsers())}
                    className="btn-ghost text-xs text-tomato"
                  >
                    Nonaktifkan
                  </button>
                </div>
              ))}
            </div>
          </section>

          <section className="card p-6">
            <h2 className="font-bold text-lg mb-4">Set Tidak Valid ({invalidSets.length})</h2>
            {invalidSets.length === 0 ? (
              <p className="opacity-50">Belum ada</p>
            ) : (
              invalidSets.map((s) => (
                <div key={s.id} className="p-2 opacity-70">{s.slug} - {s.nameZh}</div>
              ))
            )}
          </section>
        </>
      )}
    </div>
  )
}
