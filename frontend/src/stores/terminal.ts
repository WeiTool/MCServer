// ViewModel 层：终端跨路由共享状态（Pinia store）
// 把跨路由需要保留的状态集中到 store，组件挂载/卸载不会清空，
// 路由切换再回来时 UI 仍能看到上次的日志、mod 数量、Java 配置等。
import { defineStore } from "pinia";
import { ref } from "vue";
import type { LogLine } from "../utils/log";

// ============================================================
//  类型定义
// ============================================================

/** Java 运行时信息（由后端扫描系统 Java 返回） */
export interface JavaInfo {
  /** Java 安装目录路径 */
  path: string;
  /** Java 可执行文件路径（如 /bin/java.exe） */
  executable: string;
  /** Java 版本号（如 17、21、1.8） */
  version: number;
  /** Java 版本名称（如 "17.0.9" 或 "1.8.0_391"） */
  versionName: string;
}

/** 系统资源信息（由后端周期推送） */
export interface SystemInfo {
  /** 当前资源使用率（百分比，0-100） */
  usagePercent: number;
}

// ============================================================
//  Store 定义
// ============================================================

/**
 * useTerminalStore
 * 终端跨路由共享状态。所有状态在路由切换时保留，组件卸载不清空。
 * DOM 引用、定时器、UI 局部状态（commandInput）仍由组件自行管理。
 */
export const useTerminalStore = defineStore("terminal", () => {
  // ---------- 终端日志 ----------
  // 日志上限 10000 行，超过后从头部裁剪，防止内存无限增长
  const terminalLines = ref<LogLine[]>([]);
  const maxLogLines = 10000;

  /** 末尾追加一行日志（自动限制最大行数） */
  function appendLog(line: LogLine) {
    terminalLines.value.push(line);
    if (terminalLines.value.length > maxLogLines) {
      terminalLines.value.splice(0, terminalLines.value.length - maxLogLines);
    }
  }

  /** 清空所有日志 */
  function clearLogs() {
    terminalLines.value = [];
  }

  // ---------- 服务器统计 ----------
  const modCount = ref(0);
  const pluginCount = ref(0);
  const uptime = ref(0);

  function setModCount(n: number) {
    modCount.value = n;
  }
  function setPluginCount(n: number) {
    pluginCount.value = n;
  }
  function setUptime(n: number) {
    uptime.value = n;
  }

  // ---------- Java 配置 ----------
  const javaList = ref<JavaInfo[]>([]);
  const selectedJava = ref("");

  function setJavaList(list: JavaInfo[]) {
    javaList.value = list;
  }
  function setSelectedJava(path: string) {
    selectedJava.value = path;
  }
  /** 把自定义 Java 追加到列表末尾（不重复） */
  function pushJava(info: JavaInfo) {
    const exists = javaList.value.some((j) => j.executable === info.executable);
    if (!exists) javaList.value.push(info);
  }

  // ---------- 内存配置 ----------
  const xmxGB = ref(4);
  const xmsGB = ref(4);

  function setXmx(v: number) {
    xmxGB.value = v;
  }
  function setXms(v: number) {
    xmsGB.value = v;
  }

  // ---------- 实时资源使用率 ----------
  const cpuUsagePercent = ref(0);
  const memoryUsagePercent = ref(0);

  function setCpuUsagePercent(n: number) {
    cpuUsagePercent.value = n;
  }
  function setMemoryUsagePercent(n: number) {
    memoryUsagePercent.value = n;
  }

  // ---------- 初始化标志 ----------
  // 防止组件二次挂载（路由切回）时重复跑初始化链路
  const initialized = ref(false);
  function markInitialized() {
    initialized.value = true;
  }
  function isInitialized() {
    return initialized.value;
  }

  return {
    // 终端日志
    terminalLines,
    appendLog,
    clearLogs,
    // 服务器统计
    modCount,
    pluginCount,
    uptime,
    setModCount,
    setPluginCount,
    setUptime,
    // Java 配置
    javaList,
    selectedJava,
    setJavaList,
    setSelectedJava,
    pushJava,
    // 内存配置
    xmxGB,
    xmsGB,
    setXmx,
    setXms,
    // 资源使用率
    cpuUsagePercent,
    memoryUsagePercent,
    setCpuUsagePercent,
    setMemoryUsagePercent,
    // 初始化标志
    initialized,
    markInitialized,
    isInitialized,
  };
});
