// ViewModel：GC 三折线图（YGC/FGC/GCT）的配置逻辑
// 复用 useEcharts 管理生命周期，本文件只负责构建图表配置
import { ref } from "vue";
import type { EChartsOption } from "echarts";
import { useEcharts } from "../../../composables/useEcharts";

// 单次 GC 统计数据点（与后端 model.GcStats 对应）
export interface GcPoint {
  /** 年轻代 GC 次数 */
  ygc: number;
  /** Full GC 次数 */
  fgc: number;
  /** GC 总耗时（秒） */
  gct: number;
}

// 组件 props 类型定义
export interface GcChartProps {
  /** GC 统计数据点序列 */
  data?: GcPoint[];
  /** 容器宽度（px），默认自适应父容器 */
  width?: string;
  /** 容器高度（px） */
  height?: string;
}

// 三条折线的展示配置：图例用短名，tooltip 展示完整含义
const seriesConfig = [
  {
    key: "ygc" as const,
    name: "YGC",
    desc: "年轻代 GC 次数",
    color: "#4a9eff",
  },
  { key: "fgc" as const, name: "FGC", desc: "Full GC 次数", color: "#ff6b6b" },
  { key: "gct" as const, name: "GCT", desc: "GC 总耗时(秒)", color: "#ffa940" },
];

/**
 * useGcChart
 * GC 三折线图逻辑。接收 props，返回图表挂载的 DOM ref。
 * 图表生命周期由 useEcharts 统一管理。
 */
export function useGcChart(props: GcChartProps) {
  // 图表挂载的 DOM 元素
  const chartRef = ref<HTMLDivElement | null>(null);

  /** 构建折线图配置 */
  function buildOption(): EChartsOption {
    const points = props.data ?? [];

    return {
      // 数据每 2 秒更新，关闭动画避免渲染中间态导致线段/数字发虚
      animation: false,
      // 图例：标注三条折线分别是什么
      legend: {
        top: 0, 
        right: 0, 
        orient: "vertical",
        textStyle: {
          fontSize: 11,
        },
        itemGap: 10, // YGC/FGC/GCT 之间的间距
        itemWidth: 14,
        itemHeight: 10,
      },
      tooltip: {
        trigger: "axis",
        // 提示框固定在右侧垂直排列（每条线一行），不跟随光标在中间横向显示
        position: "right",
        confine: true,
        axisPointer: {
          type: "line",
        },
        textStyle: {
          fontSize: 12,
        },
        // 每条线带上完整说明（如：YGC（年轻代 GC 次数））
        formatter: (params: unknown) => {
          const list = params as Array<{
            seriesName: string;
            marker: string;
            data: number;
          }>;
          return list
            .map((p) => {
              const cfg = seriesConfig.find((c) => c.name === p.seriesName);
              const desc = cfg ? `（${cfg.desc}）` : "";
              return `${p.marker} ${p.seriesName}${desc}：${p.data}`;
            })
            .join("<br/>");
        },
      },
      grid: {
        left: "3%",
        right: "4%",
        bottom: "3%",
        top: "18%",
        containLabel: true,
      },
      xAxis: [
        {
          type: "category",
          boundaryGap: false,
          data: points.map((_, i) => `${i + 1}`),
        },
      ],
      yAxis: [
        {
          type: "value",
        },
      ],
      series: seriesConfig.map((cfg) => ({
        name: cfg.name,
        type: "line" as const,
        smooth: true,
        lineStyle: {
          width: 2,
          color: cfg.color,
        },
        itemStyle: {
          color: cfg.color,
        },
        showSymbol: false,
        emphasis: {
          focus: "series" as const,
        },
        data: points.map((p) => p[cfg.key]),
      })),
    };
  }

  // 复用 echarts 生命周期，数据变化时自动刷新
  useEcharts(chartRef, buildOption, () => [props.data]);

  return { chartRef };
}
