// ViewModel 层：服务器进程控制（启动/停止/重启）
// Dashboard 和 Terminal 共用的服务器控制逻辑
import { useActiveServer } from "./useActiveServer";
import { processApi } from "../api";
import type { StartResult } from "../api/processApi";

/**
 * useServerControl
 * 服务器进程控制公用函数。基于 useActiveServer 的当前活动服务器，
 * 封装启动 / 停止 / 重启三个动作，仅返回结果，由调用方决定如何反馈
 * （终端回显日志 / 消息提示等）。
 */
export function useServerControl() {
  const { currentServer } = useActiveServer();

  /**
   * 启动当前活动服务器（使用该服务器已配置的 Java 和内存）
   * @returns 启动结果；失败时 message 含具体原因
   */
  async function startServer(): Promise<StartResult> {
    if (!currentServer.value) {
      return { ok: false, message: "未选择服务器" };
    }
    return processApi.startServer(currentServer.value);
  }

  /**
   * 停止当前服务器
   * @returns 是否成功
   */
  async function stopServer(): Promise<boolean> {
    return processApi.stopServer();
  }

  /**
   * 重启服务器（先停止再启动）
   * @returns 启动结果（最终状态取决于启动步骤）
   */
  async function restartServer(): Promise<StartResult> {
    await stopServer();
    return startServer();
  }

  return { startServer, stopServer, restartServer };
}
