const API = '/api/v1'

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API}${path}`, {
    credentials: 'include',
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'request failed')
  }
  return res.json()
}

export interface User {
  id: number
  username: string
  role: string
  mustChangePassword: boolean
  displayName: string
}

export interface LevelSet {
  id: number
  slug: string
  authorId: number
  name: string
  nameZh: string
  status: string
  coverUrl?: string
  authorName?: string
  latestVersion?: string
  levelCount?: number
  createdAt: string
}

export interface LevelEntry {
  id: number
  levelId: string
  levelName: string
  levelNameZh: string
  sceneName: string
  screenshotUrl?: string
  sortOrder: number
}

export interface Version {
  id: number
  version: string
  uid: string
  parseStatus: string
  isLatest: boolean
  createdAt: string
}

export interface ParseJob {
  id: number
  versionId: number
  status: string
  phase: string
  progress: number
  message: string
  error?: string
}

export interface PresignResult {
  url: string
  expiresAt: string
}

export const api = {
  login: (username: string, password: string) =>
    request<{ user: User; mustChangePassword: boolean }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  me: () => request<User>('/auth/me'),

  changePassword: (newPassword: string) =>
    request('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ newPassword }),
    }),

  listSets: () => request<LevelSet[]>('/sets'),

  getSet: (slug: string) =>
    request<{ set: LevelSet; versions: Version[]; levels: LevelEntry[] }>(`/sets/${slug}`),

  downloadLatest: (slug: string) => request<PresignResult>(`/sets/${slug}/latest/download`),
  downloadBundle: (slug: string) => request<PresignResult>(`/sets/${slug}/bundle`),

  mySets: () => request<LevelSet[]>('/me/sets'),

  createSet: (slug: string, name: string, nameZh: string) =>
    request<LevelSet>('/me/sets', {
      method: 'POST',
      body: JSON.stringify({ slug, name, nameZh }),
    }),

  uploadZip: async (slug: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${API}/me/sets/${slug}/upload`, {
      method: 'POST',
      credentials: 'include',
      body: form,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(err.error || 'upload failed')
    }
    return res.json() as Promise<ParseJob>
  },

  getParseJob: (id: number) => request<ParseJob>(`/me/parse-jobs/${id}`),

  adminSets: (status?: string) =>
    request<LevelSet[]>(`/admin/sets${status ? `?status=${status}` : ''}`),

  adminPatchStatus: (id: number, status: string) =>
    request(`/admin/sets/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),

  listUsers: () => request<User[]>('/super/users'),

  createUser: (data: { username: string; password: string; displayName: string; role: string }) =>
    request<User>('/super/users', { method: 'POST', body: JSON.stringify(data) }),

  patchUser: (id: number, data: { role: string; displayName: string; disabled: boolean }) =>
    request(`/super/users/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),

  invalidateSet: (id: number) =>
    request(`/super/sets/${id}/invalidate`, { method: 'POST' }),

  listInvalidSets: () => request<LevelSet[]>('/super/invalid-sets'),
}
