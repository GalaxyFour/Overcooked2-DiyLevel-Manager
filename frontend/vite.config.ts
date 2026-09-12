import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 14557,
    proxy: {
      '/api': {
        target: 'http://localhost:14556',
        changeOrigin: true,
      },
      '/swagger': {
        target: 'http://localhost:14556',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
