// ViewModel：顶部导航栏逻辑（组合式函数）
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { WindowMinimise, Quit } from '../../../wailsjs/runtime/runtime';

/**
 * useMainNav
 * 顶部导航栏业务逻辑。返回模板所需的状态与交互方法。
 */
export function useMainNav() {
    // 当前路由与路由实例
    const route = useRoute()
    const router = useRouter()

    // 当前激活的导航项（根据路由自动高亮）
    const activeKey = computed(() => route.name as string)

    // 导航跳转：点击导航项切换到对应路由
    const handleNav = (key: string) => {
        router.push(key === 'home' ? '/' : '/console')
    }

    // 窗口控制方法
    const handleMinimize = () => {
        WindowMinimise()
    }

    const handleClose = () => {
        Quit()
    }

    return {
        activeKey,
        handleNav,
        handleMinimize,
        handleClose,
    }
}
