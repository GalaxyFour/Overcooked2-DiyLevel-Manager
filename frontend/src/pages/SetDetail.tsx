import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { useParams } from 'react-router-dom'
import { api } from '../api/client'

export default function SetDetail() {
  const { slug } = useParams<{ slug: string }>()
  const { data, isLoading } = useQuery({
    queryKey: ['set', slug],
    queryFn: () => api.getSet(slug!),
    enabled: !!slug,
  })

  const download = async (version?: string) => {
    const res = version
      ? await fetch(`/api/v1/sets/${slug}/versions/${version}/download`, { credentials: 'include' }).then((r) => r.json())
      : await api.downloadBundle(slug!)
    window.open(res.url, '_blank')
  }

  if (isLoading || !data) {
    return <div className="text-center py-20 opacity-50">加载中…</div>
  }

  const { set, versions, levels } = data
  const title = set.nameZh || set.name || set.slug

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
      <div className="card overflow-hidden mb-8">
        <div className="grid md:grid-cols-2 gap-0">
          <div className="aspect-video md:aspect-auto min-h-[200px] bg-gradient-to-br from-orange-100 to-mint/20">
            {set.coverUrl ? (
              <img src={set.coverUrl} alt={title} className="w-full h-full object-cover" />
            ) : (
              <div className="h-full flex items-center justify-center text-8xl opacity-30">🍳</div>
            )}
          </div>
          <div className="p-8 flex flex-col justify-center">
            <h1 className="text-3xl font-extrabold mb-2">{title}</h1>
            <p className="opacity-60 mb-4">作者：{set.authorName}</p>
            <p className="text-sm opacity-50 mb-6">标识：{set.slug}</p>
            <button onClick={() => download()} className="btn-primary w-fit">
              📦 下载最新安装包
            </button>
          </div>
        </div>
      </div>

      {levels.length > 0 && (
        <section className="mb-10">
          <h2 className="text-xl font-bold mb-4">关卡预览</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
            {levels.map((lv, i) => (
              <motion.div
                key={lv.id}
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: i * 0.03 }}
                className="card overflow-hidden"
              >
                <div className="aspect-video bg-orange-50 dark:bg-white/5">
                  {lv.screenshotUrl ? (
                    <img src={lv.screenshotUrl} alt={lv.levelNameZh || lv.levelName} className="w-full h-full object-cover" />
                  ) : (
                    <div className="h-full flex items-center justify-center text-3xl opacity-30">🍽️</div>
                  )}
                </div>
                <p className="p-2 text-sm font-medium truncate text-center">
                  {lv.levelNameZh || lv.levelName || lv.levelId}
                </p>
              </motion.div>
            ))}
          </div>
        </section>
      )}

      <section>
        <h2 className="text-xl font-bold mb-4">版本历史</h2>
        <div className="space-y-2">
          {versions.filter((v) => v.parseStatus === 'done').map((v) => (
            <div key={v.id} className="card p-4 flex items-center justify-between">
              <div>
                <span className="font-bold">v{v.version}</span>
                {v.isLatest && (
                  <span className="ml-2 text-xs px-2 py-0.5 rounded-full bg-mint/30 text-mint font-semibold">
                    最新
                  </span>
                )}
              </div>
              <button onClick={() => download(v.version)} className="btn-ghost text-sm">
                下载
              </button>
            </div>
          ))}
        </div>
      </section>
    </motion.div>
  )
}
