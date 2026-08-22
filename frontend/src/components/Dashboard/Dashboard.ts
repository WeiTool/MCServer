// ViewModel：首页看板逻辑（组合式函数）
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { useMessage } from "naive-ui";
import { useRouter } from "vue-router";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
import { useActiveServer } from "../../composables/useActiveServer";
import { useServerControl } from "../../composables/useServerControl";
import { serverApi, processApi } from "../../api";
import { formatUptime } from "../../utils/formatUptime.js";
import { GetPlayerList } from "../../../wailsjs/go/main/App";
import type { GcPoint } from "../base/GcChart/GcChart";

// ============================================================
//  类型定义
// ============================================================

/** 系统内存信息（由后端每 2 秒推送） */
export interface SystemInfo {
  /** 当前内存使用率（百分比，0-100） */
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

/** 服务器实例（首页卡片展示的数据结构） */
export type ServerInstance = {
  /** 服务器显示名称 */
  name: string;
  /** 服务器文件夹绝对路径 */
  path: string;
  /** 是否包含有效的 jar 文件 */
  hasJar: boolean;
  /** jar 文件数量 */
  jarCount: number;
  /** jar 文件名列表 */
  jarFiles: string[];
};

// ============================================================
//  核心业务逻辑
// ============================================================

export function useDashboard() {
  // ---------- 生命周期钩子 ----------
  onMounted(async () => {
    // 1. 注册后端事件监听
    EventsOn("memory:update", handleMemoryUpdate);
    EventsOn("cpu:update", handleCPUUpdate);
    EventsOn("server:type", handleTypeUpdate);
    EventsOn("gc:update", handleGCUpdate);
    EventsOn("jvm:update", handleJvmUpdate);
    EventsOn("io:update", handleIOUpdate);

    // 2. 初始化数据（并行加载，提升性能）
    await Promise.all([loadServerList(), loadActiveServer()]);

    // 3. 活动服务器加载完成后，加载其详细信息
    await Promise.all([
      loadExtensionsCount(),
      loadTypeAndVersion(),
      refreshUptime(),
      // 加载玩家信息
      refreshPlayerList(),
    ]);

    // 4. 启动运行时长轮询（每秒刷新）
    uptimeTimer = setInterval(refreshUptime, 1000);

    // 启动玩家信息轮询（每10秒刷新）
    playerTimer = setInterval(refreshPlayerList, 10000);

    // 监听窗口可见性变化
    document.addEventListener('visibilitychange', handleVisibilityChange);
    // 监听窗口焦点变化
    window.addEventListener('focus', handleWindowFocus);
    window.addEventListener('blur', handleWindowBlur);
  });

  onBeforeUnmount(() => {
    // 清理事件监听，防止内存泄漏
    EventsOff("memory:update");
    EventsOff("server:type");
    EventsOff("gc:update");
    EventsOff("jvm:update");
    EventsOff("io:update");

    // 清理定时器
    if (uptimeTimer) {
      clearInterval(uptimeTimer);
      uptimeTimer = null;
    }
    if (playerTimer) {
      clearInterval(playerTimer);
      playerTimer = null;
    }

    //  清理事件监听
    document.removeEventListener('visibilitychange', handleVisibilityChange);
    window.removeEventListener('focus', handleWindowFocus);
    window.removeEventListener('blur', handleWindowBlur);
  });

  // ---------- 共享状态 ----------
  const { currentServer, hasActiveServer, loadActiveServer, setActiveServer } =
    useActiveServer();

  const router = useRouter();

  /** 跳转到控制台页面 */
  function goToConsole() {
    router.push("/console");
  }

  // ---------- 服务器进程控制 ----------
  const message = useMessage();
  const { startServer, stopServer, restartServer } = useServerControl();

  /** 启动当前服务器，完成后主动重新拉取 mod/插件数量 */
  async function handleStart() {
    const result = await startServer();
    message[result.ok ? "success" : "error"](
      result.ok ? "服务器已启动" : `启动失败：${result.message}`,
    );
    if (result.ok) {
      await loadExtensionsCount();
      await refreshPlayerList(); //  启动后刷新玩家信息
    }
  }

  /** 停止当前服务器，完成后主动重新拉取 mod/插件数量 */
  async function handleStop() {
    const ok = await stopServer();
    message[ok ? "success" : "error"](ok ? "服务器已停止" : "停止失败");
    if (ok) {
      await loadExtensionsCount();
      // 停止后清空玩家列表
      playerList.value = [];
    }
  }

  /** 重启当前服务器 */
  async function handleRestart() {
    const result = await restartServer();
    message[result.ok ? "success" : "error"](
      result.ok ? "服务器已重启" : `重启失败：${result.message}`,
    );
    if (result.ok) {
      await loadExtensionsCount();
      await refreshPlayerList(); //  重启后刷新玩家信息
    }
  }

  // ---------- 服务器类型与版本 ----------
  const currentType = ref("");
  const currentVersion = ref("");

  async function loadTypeAndVersion() {
    if (!currentServer.value) {
      currentType.value = "";
      currentVersion.value = "";
      return;
    }
    const [type, version] = await Promise.all([
      serverApi.fetchServerType(currentServer.value),
      serverApi.fetchServerVersion(currentServer.value),
    ]);
    currentType.value = type;
    currentVersion.value = version;
  }

  function handleTypeUpdate(serverType: string) {
    currentType.value = serverType;
  }

  // ---------- 服务器运行时长 ----------
  const uptime = ref(0);
  let uptimeTimer: ReturnType<typeof setInterval> | null = null;

  async function refreshUptime() {
    uptime.value = await processApi.fetchServerUptime();
  }

  // ============================================================
  //  玩家信息相关
  // ============================================================

  /** 完整玩家列表（字符串数组） */
  const playerList = ref<string[]>([]);
  /** 是否正在加载 */
  const isLoadingPlayers = ref(false);

  let playerTimer: ReturnType<typeof setInterval> | null = null;
  // 轮询是否激活（页面可见且有焦点时）
  const isPollingActive = ref(true);

  /**
   * 从后端获取玩家信息（使用 GetPlayerList）
   */
  async function refreshPlayerList() {
    // 如果页面不可见或无焦点，跳过查询
    if (!isPollingActive.value) return;
    // 没有活动服务器，跳过
    if (!currentServer.value) return;

    isLoadingPlayers.value = true;
    try {
      const players = await GetPlayerList();
      playerList.value = players || [];
    } catch {
      // 查询失败保持原列表，轮询下轮自动恢复
    } finally {
      isLoadingPlayers.value = false;
    }
  }

  /**
   * 页面可见性变化处理
   */
  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      // 页面变为可见，恢复轮询
      isPollingActive.value = true;
      // 立即查询一次
      refreshPlayerList();
      // 重新启动定时器
      if (!playerTimer) {
        playerTimer = setInterval(refreshPlayerList, 10000);
      }
    } else {
      // 页面隐藏，暂停轮询
      isPollingActive.value = false;
      if (playerTimer) {
        clearInterval(playerTimer);
        playerTimer = null;
      }
    }
  }

  /**
   * 窗口获得焦点
   */
  function handleWindowFocus() {
    if (document.visibilityState === 'visible') {
      isPollingActive.value = true;
      if (!playerTimer) {
        refreshPlayerList();
        playerTimer = setInterval(refreshPlayerList, 10000);
      }
    }
  }

  /**
   * 窗口失去焦点
   */
  function handleWindowBlur() {
    isPollingActive.value = false;
    if (playerTimer) {
      clearInterval(playerTimer);
      playerTimer = null;
    }
  }

  // ---------- 模组与插件统计 ----------
  const modCount = ref(0);
  const pluginCount = ref(0);

  async function loadExtensionsCount() {
    if (!currentServer.value) {
      modCount.value = 0;
      pluginCount.value = 0;
      return;
    }
    const [mods, plugins] = await Promise.all([
      serverApi.fetchServerModCount(currentServer.value),
      serverApi.fetchServerPluginCount(currentServer.value),
    ]);
    modCount.value = mods;
    pluginCount.value = plugins;
  }

  // ---------- 服务器列表 ----------
  const serverList = ref<ServerInstance[]>([]);

  async function loadServerList() {
    const res = await serverApi.fetchServerList();
    serverList.value = res?.servers || [];
  }

  // ---------- 系统内存监控 ----------
  const CPUUsagePercent = ref(0);

  function handleCPUUpdate(cpu: SystemInfo) {
    if (cpu?.usagePercent !== undefined) {
      CPUUsagePercent.value = Math.round(cpu.usagePercent);
    }
  }

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

  // ---------- 左侧信息面板 ----------
  const infoItems = computed(() => [
    {
      icon: "server",
      label: "类型",
      value: currentType.value || "未知",
      color: "#4a9eff",
    },
    {
      icon: "gamepad",
      label: "版本",
      value: currentVersion.value || "未知",
      color: "#36cfc9",
    },
    {
      icon: "clock",
      label: "运行时长",
      value: formatUptime(uptime.value),
      color: "#73d13d",
    },
    {
      icon: "puzzle",
      label: "模组数量",
      value: String(modCount.value),
      color: "#ff6b6b",
    },
    {
      icon: "blocks",
      label: "插件数量",
      value: String(pluginCount.value),
      color: "#b37feb",
    },
  ]);

  // ============================================================
  //  对外暴露
  // ============================================================

  return {
    // 系统状态
    memoryUsagePercent,
    CPUUsagePercent,
    jvmMemoryUsagePercent,
    ioReadMBps,
    ioWriteMBps,

    // 服务器列表与选择
    serverList,
    currentServer,
    hasActiveServer,
    setActiveServer,

    // 信息面板（统一数据驱动）
    infoItems,

    // GC 折线图数据
    gcPoints,

    playerList,
    isLoadingPlayers,
    refreshPlayerList,

    // 操作方法
    goToConsole,
    loadServerList,
    handleStart,
    handleStop,
    handleRestart,
  };
}