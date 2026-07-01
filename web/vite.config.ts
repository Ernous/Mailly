import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  // Load .env from project root (one level up from /web).
  // loadEnv returns {} if the file is missing — safe in Docker/CI.
  const env = loadEnv(mode, '../', '')

  const apiTarget =
    env.SERVER_URL ||
    env.VITE_API_URL ||
    process.env.VITE_API_URL ||
    'http://localhost:3000'

  return {
    plugins: [vue()],
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
        },
      },
    },
    build: {
      chunkSizeWarningLimit: 600,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (id.includes('@tiptap')) return 'tiptap'
            if (id.includes('vuetify')) return 'vuetify'
            if (
              id.includes('node_modules/vue') ||
              id.includes('node_modules/vue-router')
            )
              return 'vue-core'
          },
        },
      },
    },
  }
})
