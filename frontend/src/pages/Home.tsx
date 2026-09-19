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
          <span className="text-tomato">Overcooked 2</span> Level Kustom
        </h1>
        <p className="text-lg opacity-70 max-w-xl mx-auto">
          Temukan set level menarik buatan komunitas, unduh dan pasang ke BepInEx/OC2DIYLevel dengan sekali klik 🎮
        </p>
        <button onClick={() => refetch()} className="btn-ghost mt-4 text-sm">
          🔄 Segarkan Daftar
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
        <div className="text-center py-20 opacity-50">Memuat… 🍳</div>
      ) : sets.length === 0 ? (
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="card p-12 text-center"
        >
          <div className="text-6xl mb-4">🥘</div>
          <p className="text-lg opacity-70">Belum ada set level yang dipublikasikan, jadilah penulis pertama!</p>
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
