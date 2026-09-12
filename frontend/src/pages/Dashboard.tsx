import { useState, useRef } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { api, ParseJob } from '../api/client'

export default function Dashboard() {
  const qc = useQueryClient()
  const { data: sets = [], refetch } = useQuery({
    queryKey: ['my-sets'],
    queryFn: api.mySets,
  })

  const [slug, setSlug] = useState('')
  const [name, setName] = useState('')
  const [nameZh, setNameZh] = useState('')
  const [uploadSlug, setUploadSlug] = useState('')
  const [job, setJob] = useState<ParseJob | null>(null)
  const [error, setError] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  const createSet = async () => {
    setError('')
    try {
      await api.createSet(slug, name, nameZh)
      setSlug('')
      setName('')
      setNameZh('')
      refetch()
    } catch (e) {
      setError(e instanceof Error ? e.message : '创建失败')
    }
  }

  const handleUpload = async (file: File) => {
    if (!uploadSlug) {
      setError('请选择关卡集')
      return
    }
    setError('')
    try {
      const j = await api.uploadZip(uploadSlug, file)
      setJob(j)
      pollJob(j.id)
    } catch (e) {
      setError(e instanceof Error ? e.message : '上传失败')
    }
  }

  const pollJob = (id: number) => {
    const timer = setInterval(async () => {
      try {
        const j = await api.getParseJob(id)
        setJob(j)
        if (j.status === 'done' || j.status === 'failed') {
          clearInterval(timer)
          refetch()
          qc.invalidateQueries({ queryKey: ['sets'] })
        }
      } catch {
        clearInterval(timer)
      }
    }, 2000)
  }

  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-extrabold">作者工作台 🧑‍🍳</h1>

      <section className="card p-6">
        <h2 className="font-bold text-lg mb-4">创建关卡集（最多 20 个）</h2>
        <div className="grid sm:grid-cols-3 gap-3 mb-4">
          <input className="input" placeholder="标识 slug (如 jia_carnival)" value={slug} onChange={(e) => setSlug(e.target.value)} />
          <input className="input" placeholder="英文名" value={name} onChange={(e) => setName(e.target.value)} />
          <input className="input" placeholder="中文名" value={nameZh} onChange={(e) => setNameZh(e.target.value)} />
        </div>
        <button onClick={createSet} className="btn-primary">创建</button>
      </section>

      <section className="card p-6">
        <h2 className="font-bold text-lg mb-4">上传关卡包</h2>
        <p className="text-sm opacity-60 mb-4">
          上传 Level Editor 导出的 zip（命名格式：slug_v版本_日期.zip）
        </p>
        <select
          className="input mb-4 max-w-md"
          value={uploadSlug}
          onChange={(e) => setUploadSlug(e.target.value)}
        >
          <option value="">选择关卡集…</option>
          {sets.map((s) => (
            <option key={s.id} value={s.slug}>{s.nameZh || s.name || s.slug}</option>
          ))}
        </select>

        <div
          className="border-2 border-dashed border-tomato/30 rounded-2xl p-10 text-center cursor-pointer hover:border-tomato/60 transition"
          onClick={() => fileRef.current?.click()}
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => {
            e.preventDefault()
            const f = e.dataTransfer.files[0]
            if (f) handleUpload(f)
          }}
        >
          <div className="text-4xl mb-2">📦</div>
          <p>拖拽 zip 到此处，或点击选择文件</p>
          <input
            ref={fileRef}
            type="file"
            accept=".zip"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) handleUpload(f)
            }}
          />
        </div>

        {job && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="mt-6 p-4 rounded-xl bg-orange-50 dark:bg-white/5"
          >
            <div className="flex justify-between text-sm mb-2">
              <span>{job.phase || job.status}</span>
              <span>{job.progress}%</span>
            </div>
            <div className="h-2 rounded-full bg-orange-100 dark:bg-white/10 overflow-hidden">
              <motion.div
                className="h-full bg-tomato rounded-full"
                initial={{ width: 0 }}
                animate={{ width: `${job.progress}%` }}
                transition={{ duration: 0.3 }}
              />
            </div>
            <p className="text-sm mt-2 opacity-70">{job.message}</p>
            {job.error && <p className="text-tomato text-sm mt-1">{job.error}</p>}
            {job.status === 'done' && (
              <p className="text-mint font-semibold mt-2">✅ 解析完成！首页已可看到最新版本</p>
            )}
          </motion.div>
        )}
      </section>

      <section>
        <h2 className="font-bold text-lg mb-4">我的关卡集</h2>
        <div className="grid sm:grid-cols-2 gap-4">
          {sets.map((s) => (
            <div key={s.id} className="card p-4 flex justify-between items-center">
              <div>
                <p className="font-bold">{s.nameZh || s.name || s.slug}</p>
                <p className="text-sm opacity-60">{s.slug} · v{s.latestVersion || '—'}</p>
              </div>
              <span className={`text-xs px-2 py-1 rounded-full ${
                s.status === 'active' ? 'bg-mint/20 text-mint' : 'bg-orange-100 text-orange-600'
              }`}>
                {s.status}
              </span>
            </div>
          ))}
        </div>
      </section>

      {error && <p className="text-tomato text-center">{error}</p>}
    </div>
  )
}
