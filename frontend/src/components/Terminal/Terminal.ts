// ViewModel：控制台（终端）业务逻辑（组合式函数）
// 跨路由保留的状态走 Pinia store（stores/terminal.ts），组件卸载不清空 store
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from "vue";
import { storeToRefs } from "pinia";
import { useMessage, type DropdownOption } from "naive-ui";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
import { GetOnlinePlayers, GetMaxPlayers, SaveLogToFile } from "../../../wailsjs/go/main/App";
import { useActiveServer } from "../../composables/useActiveServer";
import { useServerControl } from "../../composables/useServerControl";
import {
  useTerminalStore,
  type JavaInfo,
  type SystemInfo,
} from "../../stores/terminal";
import { javaApi, processApi, serverApi } from "../../api";
import { detectLogLevel, type LogLine } from "../../utils/log";
import { formatUptime } from "../../utils/formatUptime.js";

// ============================================================
//  核心业务逻辑
// ============================================================

/**
 * useTerminal
 * 控制台（终端）业务逻辑。返回模板所需的所有状态、数据与交互方法。
 * 跨路由状态由 useTerminalStore 持有，路由切换不清空。
 */
export function useTerminal() {
  // ---------- UI 工具 ----------
  const message = useMessage();

  // ---------- 共享状态 ----------
  const { currentServer, hasActiveServer, loadActiveServer } =
    useActiveServer();

  // ---------- Pinia store（跨路由保留的状态） ----------
  const store = useTerminalStore();
  // 用 storeToRefs 解构以保持响应式
  const {
    terminalLines,
    modCount,
    pluginCount,
    uptime,
    javaList,
    selectedJava,
    xmxGB,
    xmsGB,
    cpuUsagePercent,
    memoryUsagePercent,
  } = storeToRefs(store);

  // ---------- 组件局部状态（不跨路由） ----------
  // DOM 引用、定时器实例、单条命令输入框 —— 仍由组件生命周期管理
  const terminalScreenRef = ref<HTMLElement | null>(null);
  const commandInput = ref("");
  let uptimeTimer: ReturnType<typeof setInterval> | null = null;

  // ---------- 生命周期钩子 ----------
  onMounted(async () => {
    // 1. 注册后端事件监听（每次挂载都注册，卸载时取消）
    EventsOn("server:log", handleServerLog);
    EventsOn("memory:update", handleMemoryUpdate);
    EventsOn("cpu:update", handleCPUUpdate);

    // 2. 仅首次进入时跑完整初始化链路；路由切回时跳过，复用 store 中已有数据
    if (!store.isInitialized()) {
      await loadActiveServer();
      await Promise.all([
        loadJavaList(),
        loadServerJava(),
        loadServerMemory(),
        loadModCount(),
        refreshUptime(),
      ]);
      store.markInitialized();
    }

    // 3. 启动运行时长轮询（路由切回也会重启）
    uptimeTimer = setInterval(refreshUptime, 1000);

    // 4. 玩家信息轮询：仅在本页面可见且窗口聚焦时运行
    //    最小化/切走/失焦即停止，回到页面再恢复
    refreshPlayerInfo();
    startPlayerPolling();
    document.addEventListener("visibilitychange", handleVisibilityChange);
    window.addEventListener("focus", handleWindowFocus);
    window.addEventListener("blur", handleWindowBlur);
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
    stopPlayerPolling();

    document.removeEventListener("visibilitychange", handleVisibilityChange);
    window.removeEventListener("focus", handleWindowFocus);
    window.removeEventListener("blur", handleWindowBlur);
  });

  // ---------- Java 版本显示辅助 ----------
  /**
   * 生成 Java 的可读显示名称
   * @example Java 17 (17.0.9) / Java 1.8 (1.8.0_391)
   */
  function javaDisplayName(java: JavaInfo): string {
    return `Java ${java.version} (${java.versionName})`;
  }

  // ---------- 服务器统计信息 ----------
  /** 从后端刷新运行时长（秒） */
  async function refreshUptime() {
    store.setUptime(await processApi.fetchServerUptime());
  }

  /** 从后端加载当前服务器的模组与插件数量（按钮动作完成后主动调用刷新缓存值） */
  async function loadModCount() {
    if (!currentServer.value) {
      store.setModCount(0);
      store.setPluginCount(0);
      return;
    }
    const [mods, plugins] = await Promise.all([
      serverApi.fetchServerModCount(currentServer.value),
      serverApi.fetchServerPluginCount(currentServer.value),
    ]);
    store.setModCount(mods);
    store.setPluginCount(plugins);
  }

  // ---------- 玩家信息 ----------
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

  // ---------- 左侧信息面板 ----------
  const infoItems = computed(() => [
    { icon: "cpu", label: "P核-使用率", value: `${cpuUsagePercent.value}%`, color: "#4a9eff" },
    { icon: "memory", label: "内存使用", value: `${memoryUsagePercent.value}%`, color: "#36cfc9" },
    {
      icon: "users",
      label: "在线玩家",
      value: `${onlinePlayers.value}/${maxPlayers.value}`,
      color: "#ffa940",
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
    {
      icon: "clock",
      label: "运行时间",
      value: formatUptime(uptime.value),
      color: "#73d13d",
    },
  ]);

  // ---------- 终端日志 ----------
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

  // ---------- Java 配置 ----------
  /** 从后端扫描系统 Java 列表 */
  async function loadJavaList() {
    store.setJavaList(await javaApi.scanJavaList());
  }

  /** 加载当前服务器已配置的 Java（含不在扫描列表中的自定义 Java） */
  async function loadServerJava() {
    if (!currentServer.value) {
      store.setSelectedJava("");
      return;
    }

    const saved = await javaApi.getServerJava(currentServer.value);
    if (!saved) {
      store.setSelectedJava("");
      return;
    }

    // 优先从扫描列表中匹配（按 executable 路径）
    const found = javaList.value.find((j) => j.executable === saved.executable);
    if (found) {
      store.setSelectedJava(found.path);
      return;
    }

    // 自定义 Java：补充到列表末尾并选中
    store.pushJava({
      path: saved.path,
      executable: saved.executable,
      version: saved.version,
      versionName: saved.versionName,
    });
    store.setSelectedJava(saved.path);
  }

  /** Java 下拉菜单选项（含"添加自定义 Java"入口） */
  const javaDropdownOptions = computed<DropdownOption[]>(() => {
    const options: DropdownOption[] = javaList.value.map((j) => ({
      label: `${j.path}  (${javaDisplayName(j)})`,
      key: j.path,
    }));
    options.push({
      label: "添加自定义 Java",
      key: "__add__",
    });
    return options;
  });

  /** 当前选中 Java 的显示名称 */
  const selectedJavaDisplay = computed(() => {
    const found = javaList.value.find((j) => j.path === selectedJava.value);
    return found ? javaDisplayName(found) : "选择 Java";
  });

  /**
   * Java 下拉菜单选择处理
   * - 选择已有 Java：直接保存
   * - 选择 "__add__"：打开文件对话框添加自定义 Java
   */
  async function handleJavaSelect(key: string | number) {
    if (!currentServer.value) {
      message.warning("请先在首页选择当前服务器");
      return;
    }

    // 添加自定义 Java
    if (key === "__add__") {
      const added = await javaApi.addJavaByDialog();
      if (!added) return; // 用户取消

      await javaApi.setServerJava(currentServer.value, added.executable);
      await loadJavaList();
      store.setSelectedJava(added.path);
      message.success(`已添加 ${javaDisplayName(added)}`);
      return;
    }

    // 选择已有 Java
    store.setSelectedJava(String(key));
    const found = javaList.value.find((j) => j.path === selectedJava.value);
    if (found) {
      const ok = await javaApi.setServerJava(
        currentServer.value,
        found.executable,
      );
      message[ok ? "success" : "error"](
        ok ? `已选择 ${javaDisplayName(found)}` : "保存 Java 选择失败",
      );
    }
  }

  // ---------- 内存配置 ----------
  /** 加载当前服务器的内存配置（后端存储 MB，前端显示 GB） */
  async function loadServerMemory() {
    if (!currentServer.value) return;
    const [xmxMB, xmsMB] = await javaApi.getServerMemory(currentServer.value);
    if (xmxMB > 0) store.setXmx(xmxMB / 1024);
    if (xmsMB > 0) store.setXms(xmsMB / 1024);
  }

  // ---------- 服务器操作 ----------
  // 进程控制（启动/停止/重启）复用公用函数 useServerControl，
  // 此处仅保留终端特有的"保存内存配置"与"错误日志回显"逻辑
  const { startServer, stopServer } = useServerControl();

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
      action: handleExportLog,
    },
    {
      label: "导出错误日志",
      type: "default",
      icon: "alert",
      action: handleExportErrorLog,
    },
    {
      label: "清除控制台",
      type: "default",
      icon: "trash",
      action: handleClearLogs,
    },
  ];

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
      appendLog(`[错误] 启动服务器失败：${result.message}`);
    }
    // 后端启动后已重新扫描 mods/plugins 并写入 JSON，这里拉取最新值刷新面板
    await loadModCount();
  }

  /** 停止服务器，完成后主动重新拉取 mod/插件数量 */
  async function handleStopServer() {
    const ok = await stopServer();
    if (!ok) appendLog("[错误] 停止服务器失败");
    // 后端停止后已重新扫描 mods/plugins 并写入 JSON，这里拉取最新值刷新面板
    await loadModCount();
  }

  /** 重启服务器（先停止再启动），完成后数量已在启动步骤末尾刷新 */
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
    if (!ok) appendLog("[错误] 命令发送失败");
  }

  // ---------- 工具功能 ----------
  // CPU/内存使用率由后端周期推送，已搬到 store 以保留跨路由状态

  /** 处理后端推送的 CPU 使用率更新 */
  function handleCPUUpdate(cpu: SystemInfo) {
    if (cpu?.usagePercent !== undefined) {
      store.setCpuUsagePercent(Math.round(cpu.usagePercent));
    }
  }

  /** 处理后端推送的内存使用率更新 */
  function handleMemoryUpdate(mem: SystemInfo) {
    if (mem?.usagePercent !== undefined) {
      store.setMemoryUsagePercent(Math.round(mem.usagePercent));
    }
  }

  // ---------- 日志导出与清除 ----------

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

  /** 按钮点击分发 */
  function handleAction(action?: () => void) {
    action?.();
  }

  // ============================================================
  //  对外暴露
  // ============================================================

  return {
    // 信息面板
    infoItems,

    // 终端
    terminalLines,
    terminalScreenRef,
    commandInput,

    // Java 配置
    selectedJavaDisplay,
    javaDropdownOptions,

    // 内存配置
    xmxGB,
    xmsGB,

    // 状态
    currentServer,
    hasActiveServer,

    // 方法
    sendCommand,
    handleAction,
    handleJavaSelect,
    actionButtons,
  };
}
