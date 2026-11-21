import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@pclsgu/shared': path.resolve(__dirname, '../shared/src'),
    },
  },
  build: {
    outDir: '../../web/market',
    emptyOutDir: true,
  },
  server: {
    port: 5174,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})