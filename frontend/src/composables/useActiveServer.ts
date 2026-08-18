// ViewModel 层：活动服务器共享状态
// Dashboard 和 Terminal 共用的"当前活动服务器"逻辑
import { ref, computed } from 'vue'
import { serverApi } from '../api'

// 当前活动服务器名（持久化在后端 config）
const currentServer = ref('')

// 是否有活动服务器
const hasActiveServer = computed(() => currentServer.value !== '')

// 从后端加载活动服务器
async function loadActiveServer() {
  currentServer.value = await serverApi.fetchActiveServer()
}

// 设置活动服务器
async function setActiveServer(name: string) {
  if (name === currentServer.value) return
  const ok = await serverApi.setActiveServer(name)
  if (ok) {
    currentServer.value = name
  }
}

// 供组件使用
export function useActiveServer() {
  return {
    currentServer,
    hasActiveServer,
    loadActiveServer,
    setActiveServer,
  }
}
