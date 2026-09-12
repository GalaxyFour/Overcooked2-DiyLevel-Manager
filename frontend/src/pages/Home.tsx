import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { api } from '../api/client'
import SetCard from '../components/SetCard'

export default function Home() {
  const { data: sets = [], isLoading, refetch } = useQuery({
    queryKey: ['sets'],
    queryFn: api.listSets,
    refetchInterval: 10000,
  })

  const authors = [...new Set(sets.map((s) => s.authorName).filter(Boolean))]

  return (
    <div>
      <motion.section
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center mb-12"
      >
        <h1 className="text-4xl sm:text-5xl font-extrabold mb-4">
          <span className="text-tomato">胡闹厨房</span> 自定义关卡
        </h1>
        <p className="text-lg opacity-70 max-w-xl mx-auto">
          发现社区作者创作的精彩关卡集，一键下载安装到 BepInEx/OC2DIYLevel 🎮
        </p>
        <button onClick={() => refetch()} className="btn-ghost mt-4 text-sm">
          🔄 刷新列表
        </button>
      </motion.section>

      {authors.length > 0 && (
        <div className="flex flex-wrap gap-2 justify-center mb-8">
          {authors.map((a) => (
            <span
              key={a}
              className="px-4 py-1.5 rounded-full bg-mint/20 dark:bg-mint/10 text-sm font-semibold"
            >
              👨‍🍳 {a}
            </span>
          ))}
        </div>
      )}

      {isLoading ? (
        <div className="text-center py-20 opacity-50">加载中… 🍳</div>
      ) : sets.length === 0 ? (
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="card p-12 text-center"
        >
          <div className="text-6xl mb-4">🥘</div>
          <p className="text-lg opacity-70">还没有发布的关卡集，成为第一个作者吧！</p>
        </motion.div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {sets.map((set, i) => (
            <SetCard key={set.id} set={set} index={i} />
          ))}
        </div>
      )}
    </div>
  )
}
