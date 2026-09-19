import { motion } from 'framer-motion'
import { Link } from 'react-router-dom'
import { LevelSet } from '../api/client'

export default function SetCard({ set, index }: { set: LevelSet; index: number }) {
  const title = set.nameZh || set.name || set.slug

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.35 }}
      whileHover={{ y: -4, transition: { duration: 0.2 } }}
      className="card overflow-hidden"
    >
      <Link to={`/sets/${set.slug}`} className="block">
        <div className="aspect-video bg-gradient-to-br from-orange-100 to-mint/30 dark:from-kitchen-card dark:to-tomato/10 relative overflow-hidden">
          {set.coverUrl ? (
            <img src={set.coverUrl} alt={title} className="w-full h-full object-cover" />
          ) : (
            <div className="absolute inset-0 flex items-center justify-center text-6xl opacity-40">
              🍽️
            </div>
          )}
          {set.latestVersion && (
            <span className="absolute top-3 right-3 px-2.5 py-1 rounded-full text-xs font-bold bg-white/90 dark:bg-black/50 text-tomato">
              v{set.latestVersion}
            </span>
          )}
        </div>
        <div className="p-4">
          <h3 className="font-bold text-lg truncate">{title}</h3>
          <p className="text-sm opacity-60 mt-1">
            {set.authorName} · {set.levelCount ?? 0} level
          </p>
        </div>
      </Link>
    </motion.div>
  )
}
