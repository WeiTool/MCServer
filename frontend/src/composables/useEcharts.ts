// ViewModel 层：可复用的 echarts 生命周期管理
// 封装图表初始化、自适应、销毁与数据更新逻辑，
// 供 GaugeChart / TpsChart 等所有 echarts 图表组件复用
import { onMounted, onBeforeUnmount, watch, nextTick, type Ref } from 'vue'
import * as echarts from 'echarts'
import type { ECharts, EChartsOption } from 'echarts'

/**
 * useEcharts
 * 管理一个 echarts 图表的完整生命周期。
 *
 * @param containerRef 挂载图表的 DOM 元素 ref（组件内用 ref<HTMLDivElement|null>(null) 提供）
 * @param buildOption   生成图表配置的函数（返回 EChartsOption）
 * @param deps          需要触发刷新的响应式依赖数组（默认仅构建一次）
 *
 * 用法：
 *   const chartRef = ref<HTMLDivElement | null>(null)
 *   useEcharts(chartRef, () => ({ ... }), () => [props.value])
 */
export function useEcharts(
    containerRef: Ref<HTMLDivElement | null>,
    buildOption: () => EChartsOption,
    deps: () => unknown[] = () => [],
) {
    // echarts 实例
    let chart: ECharts | null = null
    // 组件是否已卸载
    let isDisposed = false
    // 容器尺寸监听器（等布局完成后初始化）
    let resizeObserver: ResizeObserver | null = null

    /** 初始化图表（容器尺寸为 0 时跳过，等待布局完成后由 ResizeObserver 触发） */
    function initChart() {
        // DOM 未挂载或组件已卸载时直接返回
        if (!containerRef.value || isDisposed) return
        const el = containerRef.value
        // 容器还没有实际尺寸时跳过（WebView 加载初期布局未完成）
        if (el.clientWidth === 0 || el.clientHeight === 0) return
        // 初始化图表
        // 高 DPI 缩放下强制至少 2x 分辨率渲染（对应 window 包的 DPI 缩放逻辑），
        // 避免窗口被系统缩放后 canvas 内容发虚、数字与线段模糊
        chart = echarts.init(el, null, {
            devicePixelRatio: Math.max(window.devicePixelRatio || 1, 2),
        })
        // 写入图表配置
        chart.setOption(buildOption())
    }

    /** 更新图表（数值或配置变化时） */
    function updateChart() {
        if (!chart) return
        chart.setOption(buildOption(), true)
    }

    /** 自适应容器/窗口尺寸变化 */
    function handleResize() {
        chart?.resize()
    }

    onMounted(async () => {
        await nextTick()
        // 尝试立即初始化（容器已有尺寸时直接成功）
        initChart()
        // 监听容器尺寸：布局完成后自动初始化 / 尺寸变化自动自适应
        if (typeof ResizeObserver !== 'undefined' && containerRef.value) {
            resizeObserver = new ResizeObserver(() => {
                // 尚未初始化时尝试初始化
                if (!chart) {
                    initChart()
                } else {
                    // 已初始化时自适应尺寸
                    handleResize()
                }
            })
            resizeObserver.observe(containerRef.value)
        }
        // 监听窗口变化，保持图表自适应
        window.addEventListener('resize', handleResize)
    })

    onBeforeUnmount(() => {
        // 标记组件已卸载
        isDisposed = true
        // 停止监听容器尺寸
        resizeObserver?.disconnect()
        resizeObserver = null
        // 移除窗口监听
        window.removeEventListener('resize', handleResize)
        // 销毁图表实例
        chart?.dispose()
        chart = null
    })

    // 监听依赖变化，动态刷新图表
    watch(deps, updateChart)
}
