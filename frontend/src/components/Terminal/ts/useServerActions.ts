// ViewModel：控制台服务器操作（启停重启/发命令/按钮配置）
// 从 useTerminal 拆出的独立模块，负责与服务器进程交互的动作
import { ref } from "vue";
import { storeToRefs } from "pinia";
import { useActiveServer } from "../../../composables/useActiveServer";
import { useTerminalStore } from "../../../stores/terminal";
import { useServerControl } from "../../../composables/useServerControl";
import { javaApi, processApi } from "../../../api";

/**
 * useServerActions
 * 服务器操作：启动（先保存内存配置）/停止/重启/发送命令。
 * 动作结果通过 appendLog 回显到终端，导出/清除日志按钮复用 useLogActions
 * 提供的能力（二者均经 useTerminal 注入，避免模块间直接耦合）。
 */
export function useServerActions(deps: {
  appendLog: (line: string) => void;
  loadModCount: () => Promise<void>;
  handleExportLog: () => Promise<void>;
  handleExportErrorLog: () => Promise<void>;
  handleClearLogs: () => void;
}) {
  const { currentServer } = useActiveServer();
  const store = useTerminalStore();
  const { xmxGB, xmsGB } = storeToRefs(store);

  // 进程控制（启动/停止/重启）复用公用函数 useServerControl
  const { startServer, stopServer } = useServerControl();

  // 单条命令输入框（组件局部状态，不跨路由）
  const commandInput = ref("");

  /**
   * 启动服务器
   * 自动保存内存配置后调用公用启动函数，失败时在终端显示错误原因；
   * 启动完成后主动重新拉取 mod/插件数量（后端已扫描并写入 ServerList.json）
   */
  async function handleStartServer() {
    // 保存内存配置（GB → MB）
    if (currentServer.value) {
      const xmx = Math.round(xmxGB.value * 1024);
      const xms = Math.round(xmsGB.value * 1024);
      await javaApi.setServerMemory(currentServer.value, xmx, xms);
    }

    const result = await startServer();
    if (!result.ok) {
      deps.appendLog(`[错误] 启动服务器失败：${result.message}`);
    }
    // 后端启动后已重新扫描 mods/plugins 并写入 JSON，这里拉取最新值刷新面板
    await deps.loadModCount();
  }

  /** 停止服务器，完成后主动重新拉取 mod/插件数量 */
  async function handleStopServer() {
    const ok = await stopServer();
    if (!ok) deps.appendLog("[错误] 停止服务器失败");
    // 后端停止后已重新扫描 mods/plugins 并写入 JSON，这里拉取最新值刷新面板
    await deps.loadModCount();
  }

  /** 重启服务器（先停止再启动），数量已在启动步骤末尾刷新 */
  async function handleRestartServer() {
    await handleStopServer();
    await handleStartServer();
  }

  /** 向运行中的服务器发送控制台命令 */
  async function sendCommand() {
    const cmd = commandInput.value.trim();
    if (!cmd) return;

    commandInput.value = "";
    const ok = await processApi.sendCommand(cmd);
    if (!ok) deps.appendLog("[错误] 命令发送失败");
  }

  // ---------- 按钮配置 ----------
  const actionButtons = [
    { label: "开服", type: "primary", icon: "play", action: handleStartServer },
    { label: "关服", type: "danger", icon: "power", action: handleStopServer },
    {
      label: "重启",
      type: "warning",
      icon: "rotate",
      action: handleRestartServer,
    },
    {
      label: "导出日志",
      type: "default",
      icon: "file",
      action: deps.handleExportLog,
    },
    {
      label: "导出错误日志",
      type: "default",
      icon: "alert",
      action: deps.handleExportErrorLog,
    },
    {
      label: "清除控制台",
      type: "default",
      icon: "trash",
      action: deps.handleClearLogs,
    },
  ];

  /** 按钮点击分发 */
  function handleAction(action?: () => void) {
    action?.();
  }

  return {
    commandInput,
    sendCommand,
    actionButtons,
    handleAction,
  };
}
