import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
// naive-ui 按需自动引入
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // naive-ui 按需自动引入组件与样式
    Components({
      resolvers: [NaiveUiResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  // 代码分隔：拆分依赖 chunk，利用浏览器缓存
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // 框架
          'vue-vendor': ['vue', 'vue-router'],
          // UI 库
          'naive-ui': ['naive-ui'],
          // 图表
          'echarts': ['echarts'],
        },
      },
    },
  },
})
