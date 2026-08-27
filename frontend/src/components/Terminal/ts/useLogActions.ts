// ViewModel：控制台日志（追加/滚动/导出/清除）
// 从 useTerminal 拆出的独立模块，负责终端日志行的全部操作
import { ref, nextTick } from "vue";
import { storeToRefs } from "pinia";
import { SaveLogToFile } from "../../../../wailsjs/go/main/App";
import { useTerminalStore } from "../../../stores/terminal";
import { detectLogLevel } from "../../../utils/log";
import { useActiveServer } from "../../../composables/useActiveServer";

/**
 * useLogActions
 * 终端日志：追加（自动识别级别 + 贴底滚动）、导出全部/错误日志、清除。
 * 日志数据存于 Pinia store（跨路由保留），DOM 滚动容器引用由本模块持有。
 */
export function useLogActions() {
  const store = useTerminalStore();
  const { terminalLines } = storeToRefs(store);
  const { currentServer } = useActiveServer();

  // DOM 引用：终端滚动容器（模板 ref="terminalScreenRef" 绑定）
  const terminalScreenRef = ref<HTMLElement | null>(null);

  /**
   * 新增一行日志并自动滚动到底部
   * 自动识别日志级别（错误/警告/命令/普通信息）
   */
  async function appendLog(line: string) {
    // 追加前记录是否贴在底部（决定追加后是否自动滚动）
    const el = terminalScreenRef.value;
    const stickToBottom = el
      ? el.scrollHeight - el.scrollTop - el.clientHeight < 40
      : true;

    // 走 store action，跨路由保留
    store.appendLog({
      text: line,
      level: detectLogLevel(line),
    });

    // 仅在原本就贴在底部时自动滚动到底部
    // 用户上翻查看历史时不强制跳底，避免高日志频率下反复触发布局回流
    await nextTick();
    if (el && stickToBottom) {
      el.scrollTop = el.scrollHeight;
    }
  }

  /** 接收后端推送的服务器日志 */
  function handleServerLog(line: string) {
    appendLog(line);
  }

  /** 生成导出文件名：服务器名-类型-时间戳.log */
  function exportFileName(tag: string): string {
    const now = new Date();
    const pad = (n: number) => String(n).padStart(2, "0");
    const stamp =
      `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}` +
      `-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
    return `${currentServer.value || "server"}-${tag}-${stamp}.log`;
  }

  /** 调用后端保存对话框写文件；成功返回路径，取消/失败返回空串 */
  async function saveLogFile(fileName: string, content: string): Promise<string> {
    try {
      return await SaveLogToFile(fileName, content);
    } catch {
      appendLog("[错误] 导出日志失败");
      return "";
    }
  }

  /** 导出全部日志到文件 */
  async function handleExportLog() {
    if (terminalLines.value.length === 0) {
      appendLog("[提示] 当前没有日志可导出");
      return;
    }
    const content = terminalLines.value.map((l) => l.text).join("\n");
    const path = await saveLogFile(exportFileName("日志"), content);
    if (path) appendLog(`> 日志已导出: ${path}`);
  }

  /** 导出警告与错误日志（过滤 error/warning 级别）到文件 */
  async function handleExportErrorLog() {
    const errorLines = terminalLines.value.filter(
      (l) => l.level === "error" || l.level === "warning",
    );
    if (errorLines.length === 0) {
      appendLog("[提示] 当前没有警告或错误日志");
      return;
    }
    const content = errorLines.map((l) => l.text).join("\n");
    const path = await saveLogFile(exportFileName("错误日志"), content);
    if (path) appendLog(`> 已导出 ${errorLines.length} 条警告/错误日志: ${path}`);
  }

  /** 清除控制台日志并回到顶部 */
  function handleClearLogs() {
    store.clearLogs();
    if (terminalScreenRef.value) {
      terminalScreenRef.value.scrollTop = 0;
    }
  }

  return {
    terminalLines,
    terminalScreenRef,
    appendLog,
    handleServerLog,
    handleExportLog,
    handleExportErrorLog,
    handleClearLogs,
  };
}
