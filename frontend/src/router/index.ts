// 路由配置
// 使用 hash 模式，桌面应用打包后以 file:// 打开也能正常跳转
// 采用懒加载：每个页面独立 chunk，按需加载，减少首屏体积
import { createRouter, createWebHashHistory } from 'vue-router'

// 路由表：点击导航即可跳转对应页面
const routes = [
    {
        // 首页（懒加载）
        path: '/',
        name: 'home',
        component: () => import('../components/Dashboard/Dashboard.vue'),
    },
    {
        // 控制台（懒加载）
        path: '/console',
        name: 'console',
        component: () => import('../components/Terminal/Terminal.vue'),
    },
]

// 创建路由实例
const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

export default router
