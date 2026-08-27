// ViewModel：控制台（终端）业务逻辑（组合式函数）
// 按功能域拆分：玩家轮询(usePlayerPolling)、Java/内存配置(useJavaConfig)、
// 日志操作(useLogActions)、服务器操作(useServerActions)、统计(useTerminalStats)。
// 本文件只负责组装各模块 + 生命周期（事件注册/定时器）。
// 跨路由保留的状态走 Pinia store（stores/terminal.ts），组件卸载不清空 store
import { onMounted, onBeforeUnmount } from "vue";
import { storeToRefs } from "pinia";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
import { useActiveServer } from "../../composables/useActiveServer";
import { useTerminalStore } from "../../stores/terminal";
import { usePlayerPolling } from "./ts/usePlayerPolling";
import { useJavaConfig } from "./ts/useJavaConfig";
import { useLogActions } from "./ts/useLogActions";
import { useServerActions } from "./ts/useServerActions";
import { useTerminalStats } from "./ts/useTerminalStats";

/**
 * useTerminal
 * 控制台（终端）业务逻辑。返回模板所需的所有状态、数据与交互方法。
 * 跨路由状态由 useTerminalStore 持有，路由切换不清空。
 */
export function useTerminal() {
  // ---------- 共享状态 ----------
  const { currentServer, hasActiveServer, loadActiveServer } =
    useActiveServer();
  const store = useTerminalStore();
  // 用 storeToRefs 解构以保持响应式（模板绑定 xmxGB/xmsGB）
  const { xmxGB, xmsGB } = storeToRefs(store);

  // ---------- 各功能域模块 ----------
  // 玩家信息：在线/最大玩家数轮询（页面可见 + 有焦点时才轮询）
  const players = usePlayerPolling();
  // Java/内存配置：扫描列表、当前服务器选择、自定义添加、内存读取
  const java = useJavaConfig();
  // 日志：追加/滚动/导出/清除
  const logs = useLogActions();
  // 统计：时长/模组插件/信息面板/CPU内存推送
  const stats = useTerminalStats({
    onlinePlayers: players.onlinePlayers,
    maxPlayers: players.maxPlayers,
  });
  // 服务器操作：启停重启/命令/按钮组（依赖日志回显与统计刷新）
  const actions = useServerActions({
    appendLog: logs.appendLog,
    loadModCount: stats.loadModCount,
    handleExportLog: logs.handleExportLog,
    handleExportErrorLog: logs.handleExportErrorLog,
    handleClearLogs: logs.handleClearLogs,
  });

  let uptimeTimer: ReturnType<typeof setInterval> | null = null;

  // ---------- 生命周期钩子 ----------
  onMounted(async () => {
    // 1. 注册后端事件监听（每次挂载都注册，卸载时取消）
    EventsOn("server:log", logs.handleServerLog);
    EventsOn("memory:update", stats.handleMemoryUpdate);
    EventsOn("cpu:update", stats.handleCPUUpdate);

    // 2. 仅首次进入时跑完整初始化链路；路由切回时跳过，复用 store 中已有数据
    if (!store.isInitialized()) {
      await loadActiveServer();
      await Promise.all([
        java.loadJavaList(),
        java.loadServerJava(),
        java.loadServerMemory(),
        stats.loadModCount(),
        stats.refreshUptime(),
      ]);
      store.markInitialized();
    }

    // 3. 启动运行时长轮询（路由切回也会重启）
    uptimeTimer = setInterval(stats.refreshUptime, 1000);

    // 4. 玩家信息轮询：仅在本页面可见且窗口聚焦时运行
    //    最小化/切走/失焦即停止，回到页面再恢复
    players.refreshPlayerInfo();
    players.startPlayerPolling();
    document.addEventListener("visibilitychange", players.handleVisibilityChange);
    window.addEventListener("focus", players.handleWindowFocus);
    window.addEventListener("blur", players.handleWindowBlur);
  });

  onBeforeUnmount(() => {
    // 清理事件监听（store 状态保留，不清理）
    EventsOff("server:log");
    EventsOff("memory:update");
    EventsOff("cpu:update");

    // 清理定时器
    if (uptimeTimer) {
      clearInterval(uptimeTimer);
      uptimeTimer = null;
    }
    players.stopPlayerPolling();

    document.removeEventListener("visibilitychange", players.handleVisibilityChange);
    window.removeEventListener("focus", players.handleWindowFocus);
    window.removeEventListener("blur", players.handleWindowBlur);
  });

  // ============================================================
  //  对外暴露
  // ============================================================

  return {
    // 信息面板
    infoItems: stats.infoItems,

    // 终端
    terminalLines: logs.terminalLines,
    terminalScreenRef: logs.terminalScreenRef,
    commandInput: actions.commandInput,

    // Java 配置
    selectedJavaDisplay: java.selectedJavaDisplay,
    javaDropdownOptions: java.javaDropdownOptions,

    // 内存配置
    xmxGB,
    xmsGB,

    // 状态
    currentServer,
    hasActiveServer,

    // 方法
    sendCommand: actions.sendCommand,
    handleAction: actions.handleAction,
    handleJavaSelect: java.handleJavaSelect,
    actionButtons: actions.actionButtons,
  };
}
