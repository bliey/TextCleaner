import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  // GitHub Pages 项目站点 URL: https://bliey.github.io/TextCleaner/
  // 部署到 GitHub Pages 时由 CI 设置 VITE_BASE=/TextCleaner/；
  // Wails 本地构建（嵌入 exe）不设置该变量，默认 base=/ 保证资源可加载。
  base: process.env.VITE_BASE || '/',
  plugins: [vue()],
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
  },
})
