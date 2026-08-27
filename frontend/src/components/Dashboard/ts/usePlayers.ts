// ViewModel：首页玩家列表（轮询 + 可见性/焦点门控）
// 从 useDashboard 拆出的独立模块，只负责玩家信息的拉取与轮询调度
import { ref } from "vue";
import { GetPlayerList } from "../../../../wailsjs/go/main/App";
import { useActiveServer } from "../../../composables/useActiveServer";

/**
 * usePlayers
 * 首页玩家列表：每 10 秒轮询一次，
 * 页面隐藏或窗口失焦时暂停轮询，回到页面再恢复。
 */
export function usePlayers() {
  const { currentServer } = useActiveServer();

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

  /** 启动玩家轮询（页面可见且有焦点时调用） */
  function startPlayerPolling() {
    if (!playerTimer) {
      playerTimer = setInterval(refreshPlayerList, 10000);
    }
  }

  /** 停止玩家轮询（最小化/切走/失焦时调用） */
  function stopPlayerPolling() {
    if (playerTimer) {
      clearInterval(playerTimer);
      playerTimer = null;
    }
  }

  /**
   * 页面可见性变化处理
   */
  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      // 页面变为可见，恢复轮询
      isPollingActive.value = true;
      // 立即查询一次
      refreshPlayerList();
      // 重新启动定时器
      startPlayerPolling();
    } else {
      // 页面隐藏，暂停轮询
      isPollingActive.value = false;
      stopPlayerPolling();
    }
  }

  /**
   * 窗口获得焦点
   */
  function handleWindowFocus() {
    if (document.visibilityState === "visible") {
      isPollingActive.value = true;
      startPlayerPolling();
      refreshPlayerList();
    }
  }

  /**
   * 窗口失去焦点
   */
  function handleWindowBlur() {
    isPollingActive.value = false;
    stopPlayerPolling();
  }

  return {
    playerList,
    isLoadingPlayers,
    refreshPlayerList,
    startPlayerPolling,
    stopPlayerPolling,
    handleVisibilityChange,
    handleWindowFocus,
    handleWindowBlur,
  };
}
