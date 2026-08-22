// ViewModel：磁盘读写横向柱状图（读取/写入 2 条）的配置逻辑
// 复用 useEcharts 管理生命周期，本文件只负责构建图表配置
import { ref } from 'vue'
import type { EChartsOption } from 'echarts'
import { useEcharts } from '../../../composables/useEcharts'

// 组件 props 类型定义
export interface IoChartProps {
  /** 磁盘读取速率（MB/s） */
  read?: number
  /** 磁盘写入速率（MB/s） */
  write?: number
  /** 容器宽度（px），默认自适应父容器 */
  width?: string
  /** 容器高度（px） */
  height?: string
}

// 两条柱的展示配置（读取/写入）
const bars = [
  { label: '读取', color: '#4a9eff' },
  { label: '写入', color: '#ffa940' },
]

/**
 * useIoChart
 * 磁盘读写横向柱状图逻辑。接收 props，返回图表挂载的 DOM ref。
 * 图表生命周期由 useEcharts 统一管理。
 */
export function useIoChart(props: IoChartProps) {
    // 图表挂载的 DOM 元素
    const chartRef = ref<HTMLDivElement | null>(null)

    /** 构建柱状图配置 */
    function buildOption(): EChartsOption {
        const values = [props.read ?? 0, props.write ?? 0]

        return {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow',
                },
                // 每根柱带上单位
                formatter: (params: unknown) => {
                    const items = params as Array<{ dataIndex: number; value: number }>
                    return items
                        .map((p) => `${bars[p.dataIndex]?.label ?? ''}：${p.value.toFixed(2)} MB/s`)
                        .join('<br/>')
                },
            },
            grid: {
                left: '3%',
                right: '3%',
                bottom: '3%',
                top: '12%',
                containLabel: true,
            },
            // 竖向柱状图：类别在横轴，数值在纵轴，窄容器内不会横向溢出
            xAxis: [
                {
                    type: 'category',
                    data: bars.map((b) => b.label),
                },
            ],
            yAxis: [
                {
                    type: 'value',
                },
            ],
            series: [
                {
                    type: 'bar',
                    barWidth: 26,
                    itemStyle: {
                        borderRadius: [6, 6, 0, 0],
                    },
                    label: {
                        show: true,
                        position: 'top',
                        fontSize: 11,
                        formatter: (p: unknown) => `${(p as { value: number }).value.toFixed(2)} MB/s`,
                    },
                    data: values.map((v, i) => ({
                        value: v,
                        itemStyle: {
                            color: bars[i]?.color,
                        },
                    })),
                },
            ],
        }
    }

    // 复用 echarts 生命周期，数值变化时自动刷新
    useEcharts(chartRef, buildOption, () => [props.read, props.write])

    return { chartRef }
}
