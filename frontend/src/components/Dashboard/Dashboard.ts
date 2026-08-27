// ViewModel：首页看板逻辑（组合式函数）
// 按功能域拆分：服务器信息(useServerInfo)、实时指标(useRealtimeStats)、
// 玩家列表(usePlayers)、进程控制(useProcessControl)，本文件只负责组装与生命周期
import { onMounted, onBeforeUnmount, ref } from "vue";
import { useRouter } from "vue-router";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
import { useActiveServer } from "../../composables/useActiveServer";
import { useServerInfo } from "./ts/useServerInfo";
import { useRealtimeStats } from "./ts/useRealtimeStats";
import { usePlayers } from "./ts/usePlayers";
import { useProcessControl } from "./ts/useProcessControl";
import { useFileDrop } from "./ts/useFileDrop";

export type { ServerInstance } from "./ts/types";

export function useDashboard() {
  // ---------- 共享状态 ----------
  const { currentServer, hasActiveServer, loadActiveServer, setActiveServer } =
    useActiveServer();

  const router = useRouter();

  /** 跳转到控制台页面 */
  function goToConsole() {
    router.push("/console");
  }

  // ---------- 各功能域模块 ----------
  const info = useServerInfo();
  const stats = useRealtimeStats();
  const players = usePlayers();
  const fileDrop = useFileDrop();
  const control = useProcessControl({
    loadExtensionsCount: info.loadExtensionsCount,
    refreshPlayerList: players.refreshPlayerList,
    clearPlayerList: () => {
      players.playerList.value = [];
    },
  });

  // ---------- 生命周期钩子 ----------
  onMounted(async () => {
    // 1. 注册后端事件监听
    EventsOn("memory:update", stats.handleMemoryUpdate);
    EventsOn("cpu:update", stats.handleCPUUpdate);
    EventsOn("server:type", info.handleTypeUpdate);
    EventsOn("gc:update", stats.handleGCUpdate);
    EventsOn("jvm:update", stats.handleJvmUpdate);
    EventsOn("io:update", stats.handleIOUpdate);

    // 2. 初始化数据
    await Promise.all([info.loadServerList(), loadActiveServer()]);

    // 3. 加载详细信息
    await Promise.all([
      info.loadExtensionsCount(),
      info.loadTypeAndVersion(),
      info.refreshUptime(),
      players.refreshPlayerList(),
    ]);

    // 4. 启动轮询
    uptimeTimer = setInterval(info.refreshUptime, 1000);
    players.startPlayerPolling();

    document.addEventListener(
      "visibilitychange",
      players.handleVisibilityChange,
    );
    window.addEventListener("focus", players.handleWindowFocus);
    window.addEventListener("blur", players.handleWindowBlur);
  });

  let uptimeTimer: ReturnType<typeof setInterval> | null = null;

  onBeforeUnmount(() => {
    EventsOff("memory:update");
    EventsOff("server:type");
    EventsOff("gc:update");
    EventsOff("jvm:update");
    EventsOff("io:update");

    if (uptimeTimer) {
      clearInterval(uptimeTimer);
      uptimeTimer = null;
    }
    players.stopPlayerPolling();

    document.removeEventListener(
      "visibilitychange",
      players.handleVisibilityChange,
    );
    window.removeEventListener("focus", players.handleWindowFocus);
    window.removeEventListener("blur", players.handleWindowBlur);
  });

  // ============================================================
  //  对外暴露 - 只保留模板需要用到的
  // ============================================================

  return {
    // 系统状态（模板中用到的）
    memoryUsagePercent: stats.memoryUsagePercent,
    CPUUsagePercent: stats.CPUUsagePercent,
    jvmMemoryUsagePercent: stats.jvmMemoryUsagePercent,
    ioReadMBps: stats.ioReadMBps,
    ioWriteMBps: stats.ioWriteMBps,

    // 服务器列表与选择（模板中用到的）
    serverList: info.serverList,
    currentServer,
    hasActiveServer,
    setActiveServer,

    // 信息面板（模板中用到的）
    infoItems: info.infoItems,

    // GC 数据（模板中用到的）
    gcPoints: stats.gcPoints,

    // 玩家数据（模板中用到的）
    playerList: players.playerList,
    isLoadingPlayers: players.isLoadingPlayers,
    refreshPlayerList: players.refreshPlayerList,

    // 拖拽状态（模板中用到的）
    isProcessing: fileDrop.isProcessing,

    // 操作方法（模板中用到的）
    goToConsole,
    loadServerList: info.loadServerList,
    handleStart: control.handleStart,
    handleStop: control.handleStop,
    handleRestart: control.handleRestart,
  };
}
