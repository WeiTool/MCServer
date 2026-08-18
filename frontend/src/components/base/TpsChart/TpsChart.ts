// ViewModel：TPS 折线图（渐变面积图）的配置逻辑
// 复用 useEcharts 管理生命周期，本文件只负责构建图表配置
import { ref } from 'vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import { useEcharts } from '../../../composables/useEcharts'

// 组件 props 类型定义
export interface TpsChartProps {
    /** 折线数据 */
    data?: number[]
    /** 图表标题 */
    title?: string
    /** 折线颜色 */
    color?: string
    /** 容器宽度（px），默认自适应父容器 */
    width?: string
    /** 容器高度（px） */
    height?: string
}

/**
 * useTpsChart
 * TPS 折线图逻辑。接收 props，返回图表挂载的 DOM ref。
 * 图表生命周期由 useEcharts 统一管理。
 *
 * 注意：必须直接使用 props（from withDefaults，字段是响应式的），
 * 不要复制成普通对象，否则会丢失响应性，导致数据变化时图表不刷新。
 */
export function useTpsChart(props: TpsChartProps) {
    // 图表挂载的 DOM 元素
    const chartRef = ref<HTMLDivElement | null>(null)

    /** 构建折线图配置（可选字段用 ?? 兜底默认值） */
    function buildOption(): EChartsOption {
        // 渐变颜色（由上至下）
        const gradient = new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: props.color ?? '#4a9eff',
            },
            {
                offset: 1,
                color: 'rgba(0,0,0,0)',
            },
        ])

        return {
            color: [props.color ?? '#4a9eff'],
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross',
                    label: {
                        backgroundColor: '#6a7985',
                    },
                },
            },
            grid: {
                left: '3%',
                right: '4%',
                bottom: '3%',
                top: '8%',
                containLabel: true,
            },
            xAxis: [
                {
                    type: 'category',
                    boundaryGap: false,
                    data: (props.data ?? []).map((_, i) => `${i + 1}s`),
                },
            ],
            yAxis: [
                {
                    type: 'value',
                },
            ],
            series: [
                {
                    name: props.title ?? 'TPS',
                    type: 'line',
                    smooth: true,
                    lineStyle: {
                        width: 2,
                        color: props.color ?? '#4a9eff',
                    },
                    showSymbol: false,
                    areaStyle: {
                        opacity: 0.6,
                        color: gradient,
                    },
                    emphasis: {
                        focus: 'series',
                    },
                    data: props.data ?? [],
                },
            ],
        }
    }

    // 复用 echarts 生命周期，数据/标题/配色变化时自动刷新
    // 注意：deps 里直接用响应式的 props 字段，数据变化才会触发 watch 刷新
    useEcharts(chartRef, buildOption, () => [props.data, props.title, props.color])

    return { chartRef }
}
