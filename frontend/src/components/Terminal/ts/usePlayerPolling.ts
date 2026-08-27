// ViewModel：控制台玩家信息轮询
// 从 useTerminal 拆出的独立模块：在线/最大玩家数查询 + 可见性/焦点门控
import { ref } from "vue";
import { GetOnlinePlayers, GetMaxPlayers } from "../../../../wailsjs/go/main/App";
import { useActiveServer } from "../../../composables/useActiveServer";

/**
 * usePlayerPolling
 * 在线/最大玩家数：每 10 秒轮询一次，
 * 页面隐藏或窗口失焦时暂停轮询，回到页面再恢复。
 */
export function usePlayerPolling() {
  const { hasActiveServer } = useActiveServer();

  // 后端 GetOnlinePlayers / GetMaxPlayers 查询当前活动服务器（无需参数）
  // 数字直接赋值，不做加载动画
  const onlinePlayers = ref(0);
  const maxPlayers = ref(0);

  // 轮询是否激活（本页面可见且有焦点）
  const isPollingActive = ref(true);
  let playerTimer: ReturnType<typeof setInterval> | null = null;

  /** 查询一次在线/最大玩家数，直接更新数字 */
  async function refreshPlayerInfo() {
    if (!isPollingActive.value) return;
    if (!hasActiveServer.value) return;

    try {
      const [online, max] = await Promise.all([
        GetOnlinePlayers(),
        GetMaxPlayers(),
      ]);
      onlinePlayers.value = online || 0;
      maxPlayers.value = max || 0;
    } catch {
      // 服务器未运行或 Query 未启用：数字清零
      onlinePlayers.value = 0;
      maxPlayers.value = 0;
    }
  }

  /** 启动玩家轮询（页面可见且有焦点时调用） */
  function startPlayerPolling() {
    if (!playerTimer) {
      refreshPlayerInfo();
      playerTimer = setInterval(refreshPlayerInfo, 10000);
    }
  }

  /** 停止玩家轮询（最小化/切走/失焦时调用） */
  function stopPlayerPolling() {
    if (playerTimer) {
      clearInterval(playerTimer);
      playerTimer = null;
    }
  }

  /** 页面可见性变化：切回可见恢复轮询，隐藏（最小化/切走）停止 */
  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      isPollingActive.value = true;
      startPlayerPolling();
    } else {
      isPollingActive.value = false;
      stopPlayerPolling();
    }
  }

  /** 窗口获得焦点：恢复轮询 */
  function handleWindowFocus() {
    if (document.visibilityState === "visible") {
      isPollingActive.value = true;
      startPlayerPolling();
    }
  }

  /** 窗口失去焦点：停止轮询 */
  function handleWindowBlur() {
    isPollingActive.value = false;
    stopPlayerPolling();
  }

  return {
    onlinePlayers,
    maxPlayers,
    refreshPlayerInfo,
    startPlayerPolling,
    stopPlayerPolling,
    handleVisibilityChange,
    handleWindowFocus,
    handleWindowBlur,
  };
}
