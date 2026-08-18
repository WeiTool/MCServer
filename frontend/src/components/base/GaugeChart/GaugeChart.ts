// ViewModel：仪表盘（Gauge）图表的配置逻辑
// 复用 useEcharts 管理生命周期，本文件只负责构建图表配置
import { ref } from 'vue'
import type { EChartsOption } from 'echarts'
import { useEcharts } from '../../../composables/useEcharts'

// 组件 props 类型定义
export interface GaugeChartProps {
    /** 当前数值 */
    value: number
    /** 最大值（表盘量程上限） */
    max?: number
    /** 标题（默认不显示） */
    title?: string
    /** 仪表盘配色 */
    color?: string
    /** 容器宽度（px），默认自适应父容器 */
    width?: string
    /** 容器高度（px） */
    height?: string
}

/**
 * useGaugeChart
 * 仪表盘图表逻辑。接收 props，返回图表挂载的 DOM ref。
 * 图表生命周期由 useEcharts 统一管理。
 *
 * 注意：必须直接使用 props（from withDefaults，字段是响应式的），
 * 不要复制成普通对象，否则会丢失响应性，导致数值变化时图表不刷新。
 */
export function useGaugeChart(props: GaugeChartProps) {
    // 图表挂载的 DOM 元素
    const chartRef = ref<HTMLDivElement | null>(null)

    /** 构建仪表盘配置（可选字段用 ?? 兜底默认值） */
    function buildOption(): EChartsOption {
        return {
            series: [
                {
                    type: 'gauge',
                    // 起始/结束角度
                    startAngle: 220,
                    endAngle: -40,
                    min: 0,
                    max: props.max ?? 100,
                    // 半径和中心位置
                    radius: '70%',
                    center: ['50%', '50%'],
                    // 进度条（当前值彩色弧）
                    progress: {
                        show: true,
                        width: 14,
                        roundCap: true,
                        itemStyle: {
                            color: props.color ?? '#4a9eff',
                        },
                    },
                    // 表盘底色环
                    axisLine: {
                        roundCap: true,
                        lineStyle: {
                            width: 14,
                            color: [[1, '#eaeaea']],
                        },
                    },
                    // 隐藏小刻度线
                    axisTick: { show: false },
                    // 隐藏大刻度分隔线
                    splitLine: { show: false },
                    // 隐藏刻度数字，避免杂乱重叠
                    axisLabel: { show: false },
                    // 中心圆点锚点
                    anchor: {
                        show: true,
                        showAbove: true,
                        size: 18,
                        itemStyle: {
                            color: props.color ?? '#4a9eff',
                            borderColor: '#ffffff',
                            borderWidth: 4,
                        },
                    },
                    // 标题（默认隐藏，由外部 HTML 渲染更灵活）
                    title: { show: false },
                    // 中央大数值
                    detail: {
                        valueAnimation: true,
                        formatter: '{value}',
                        fontSize: 20,
                        fontWeight: 'normal',
                        offsetCenter: [0, '60%'],
                        color: '#333333',
                    },
                    data: [
                        {
                            value: props.value,
                            name: props.title ?? '',
                        },
                    ],
                },
            ],
        }
    }

    // 复用 echarts 生命周期，数值/配色变化时自动刷新
    // 注意：deps 里直接用响应式的 props 字段，数值变化才会触发 watch 刷新
    useEcharts(chartRef, buildOption, () => [props.value, props.max, props.title, props.color])

    return { chartRef }
}
