// ViewModel：首页服务器进程控制（启动/停止/重启）
// 从 useDashboard 拆出的独立模块，只负责进程控制动作与反馈
import { useMessage } from "naive-ui";
import { useServerControl } from "../../../composables/useServerControl";

/**
 * useProcessControl
 * 首页启停/重启动作：调用公用 useServerControl 执行进程操作，
 * 成功后主动刷新模组/插件数量与玩家列表（后端已重新扫描）。
 * 需要外部提供 refresh 回调（由 useDashboard 注入，避免循环依赖）。
 */
export function useProcessControl(refresh: {
  /** 重新拉取模组/插件数量 */
  loadExtensionsCount: () => Promise<void>;
  /** 重新拉取玩家列表 */
  refreshPlayerList: () => Promise<void>;
  /** 清空玩家列表（停止成功后） */
  clearPlayerList: () => void;
}) {
  const message = useMessage();
  const { startServer, stopServer, restartServer } = useServerControl();

  /** 启动当前服务器，完成后主动重新拉取 mod/插件数量 */
  async function handleStart() {
    const result = await startServer();
    message[result.ok ? "success" : "error"](
      result.ok ? "服务器已启动" : `启动失败：${result.message}`,
    );
    if (result.ok) {
      await refresh.loadExtensionsCount();
      await refresh.refreshPlayerList(); //  启动后刷新玩家信息
    }
  }

  /** 停止当前服务器，完成后主动重新拉取 mod/插件数量 */
  async function handleStop() {
    const ok = await stopServer();
    message[ok ? "success" : "error"](ok ? "服务器已停止" : "停止失败");
    if (ok) {
      await refresh.loadExtensionsCount();
      // 停止后清空玩家列表
      refresh.clearPlayerList();
    }
  }

  /** 重启当前服务器 */
  async function handleRestart() {
    const result = await restartServer();
    message[result.ok ? "success" : "error"](
      result.ok ? "服务器已重启" : `重启失败：${result.message}`,
    );
    if (result.ok) {
      await refresh.loadExtensionsCount();
      await refresh.refreshPlayerList(); //  重启后刷新玩家信息
    }
  }

  return {
    handleStart,
    handleStop,
    handleRestart,
  };
}
