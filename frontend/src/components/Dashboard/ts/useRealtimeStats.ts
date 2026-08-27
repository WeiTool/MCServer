// ViewModel：首页实时指标（CPU/内存/JVM/GC/IO）
// 从 useDashboard 拆出的独立模块，只负责处理后端每 2 秒推送的事件并更新状态
import { ref } from "vue";
import type { GcPoint } from "../../base/GcChart/GcChart";

/** 系统资源信息（由后端每 2 秒推送） */
export interface SystemInfo {
  /** 当前资源使用率（百分比，0-100） */
  usagePercent: number;
}

/** 后端推送的 GC 统计（对应 model.GcStats） */
export interface GcStatsPayload {
  /** 年轻代 GC 次数 */
  ygc: number;
  /** Full GC 次数 */
  fgc: number;
  /** GC 总耗时（秒） */
  gct: number;
}

/** 后端推送的磁盘读写速率（对应 model.IoStats，单位字节/秒） */
export interface IoStatsPayload {
  /** 磁盘读取速率（字节/秒） */
  readBytesPerSec: number;
  /** 磁盘写入速率（字节/秒） */
  writeBytesPerSec: number;
}

/**
 * useRealtimeStats
 * 首页右侧实时指标：CPU / 内存 / JVM / GC / IO。
 * 均由后端事件驱动（memory:cpu:jvm:gc:io:update），
 * 事件注册与注销在 useDashboard 生命周期中统一处理。
 */
export function useRealtimeStats() {
  // ---------- 系统 CPU 监控 ----------
  const CPUUsagePercent = ref(0);

  function handleCPUUpdate(cpu: SystemInfo) {
    if (cpu?.usagePercent !== undefined) {
      CPUUsagePercent.value = Math.round(cpu.usagePercent);
    }
  }

  // ---------- 系统内存监控 ----------
  const memoryUsagePercent = ref(0);

  function handleMemoryUpdate(mem: SystemInfo) {
    if (mem?.usagePercent !== undefined) {
      memoryUsagePercent.value = Math.round(mem.usagePercent);
    }
  }

  // ---------- JVM GC 统计 ----------
  // 滚动窗口最多保留 60 个采样点（2 秒推送一次，约 2 分钟）
  const gcPoints = ref<GcPoint[]>([]);
  const maxGcPoints = 60;

  function handleGCUpdate(stats: GcStatsPayload) {
    if (!stats) return;
    const point: GcPoint = {
      ygc: stats.ygc || 0,
      fgc: stats.fgc || 0,
      gct: stats.gct || 0,
    };
    // 不可变更新：生成新数组替换旧引用，GcChart 的 watch 才能感知变化并刷新
    gcPoints.value = [...gcPoints.value, point].slice(-maxGcPoints);
  }

  // ---------- JVM 内存使用率 ----------
  // 后端用 GetJVMProcessMemoryUsage 采集（JVM 占系统总内存百分比）
  const jvmMemoryUsagePercent = ref(0);

  function handleJvmUpdate(info: SystemInfo) {
    if (info?.usagePercent !== undefined) {
      jvmMemoryUsagePercent.value = Math.round(info.usagePercent);
    }
  }

  // ---------- 磁盘读写速率 ----------
  // 后端推送字节/秒，前端转 MB/s 供横向柱状图显示
  const ioReadMBps = ref(0);
  const ioWriteMBps = ref(0);

  function handleIOUpdate(stats: IoStatsPayload) {
    if (!stats) return;
    ioReadMBps.value = (stats.readBytesPerSec || 0) / 1024 / 1024;
    ioWriteMBps.value = (stats.writeBytesPerSec || 0) / 1024 / 1024;
  }

  return {
    CPUUsagePercent,
    memoryUsagePercent,
    jvmMemoryUsagePercent,
    gcPoints,
    ioReadMBps,
    ioWriteMBps,
    handleCPUUpdate,
    handleMemoryUpdate,
    handleGCUpdate,
    handleJvmUpdate,
    handleIOUpdate,
  };
}
