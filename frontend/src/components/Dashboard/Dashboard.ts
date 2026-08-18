// ViewModel：首页看板逻辑（组合式函数）
import type { Component } from "vue";
import { ref, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
import {
  Server,
  Gamepad,
  Power,
  RotateCw,
  Play,
  TerminalSquare,
} from "@lucide/vue";
import Froge from "../../assets/versionIcon/froge.png";
import Fabric from "../../assets/versionIcon/fabric.png";
// ViewModel：活动服务器共享状态
import { useActiveServer } from "../../composables/useActiveServer";
// Model：服务器与进程 API
import { serverApi, processApi } from "../../api";

type IconType = Component | string;

// 定义内存信息接口
interface MemoryInfo {
  usagePercent: number;
}

// 定义服务器实例接口
type ServerInstance = {
  name: string;
  path: string;
  hasJar: boolean;
  jarCount: number;
  jarFiles: string[];
  isRunning: boolean;
  pid: number;
};

/**
 * useDashboard
 * 首页看板业务逻辑。返回模板所需的状态与方法。
 */
export function useDashboard() {
  // 服务器类型（由后端检测后写入 json，响应式）
  const currentType = ref("");

  // 服务器版本（由后端检测后写入 json，响应式）
  const currentVersion = ref("");

  // 服务器运行时长（秒，每秒从后端刷新，响应式）
  const uptime = ref(0);

  // 模组数量
  const modCount = ref(0);

  // 插件数量
  const pluginCount = ref(0);

  // 加载当前服务器的 mod 与插件数量（调用 serverApi）
  async function loadExtensionsCount() {
    if (!currentServer.value) {
      modCount.value = 0;
      pluginCount.value = 0;
      return;
    }
    // 分别获取 mod 与插件数量
    modCount.value = await serverApi.fetchServerModCount(currentServer.value);
    pluginCount.value = await serverApi.fetchServerPluginCount(
      currentServer.value,
    );
  }

  // 加载当前服务器的类型与版本（从后端 json 读取）
  async function loadTypeAndVersion() {
    if (!currentServer.value) {
      currentType.value = "";
      currentVersion.value = "";
      return;
    }
    // 分别获取类型与版本
    currentType.value = await serverApi.fetchServerType(currentServer.value);
    currentVersion.value = await serverApi.fetchServerVersion(
      currentServer.value,
    );
  }

  // 服务器列表
  const serverList = ref<ServerInstance[]>([]);

  // 获取服务器列表（Model 层封装）
  async function loadServerList() {
    const res = await serverApi.fetchServerList();
    serverList.value = res?.servers || [];
  }

  // 内存使用率（百分比），由后端主动推送更新
  const memoryUsagePercent = ref(0);

  // 活动服务器状态（共享 ViewModel）
  const { currentServer, hasActiveServer, loadActiveServer, setActiveServer } =
    useActiveServer();

  // 路由实例（用于跳转到控制台页面）
  const router = useRouter();

  // 跳转到控制台页面
  function goToConsole() {
    router.push("/console");
  }

  // 版本映射
  const versionMap: Record<string, IconType> = {
    Froge: Froge,
    Fabric: Fabric,
  };

  function getVersionIcon(version: string): IconType {
    return versionMap[version] || Server;
  }

  // 处理后端推送的内存信息
  function handleMemoryUpdate(mem: MemoryInfo) {
    if (mem && typeof mem.usagePercent === "number") {
      memoryUsagePercent.value = Math.round(mem.usagePercent);
    }
  }

  // 运行时长定时器句柄（用于卸载时清理）
  let uptimeTimer: ReturnType<typeof setInterval> | null = null;

  // 刷新运行时长（每秒从后端读取精确秒数）
  async function refreshUptime() {
    uptime.value = await processApi.fetchServerUptime();
  }

  // 格式化运行时长：秒 -> "Xh Ym" / "Ym Zs" / "Zs"
  function formatUptime(seconds: number): string {
    if (seconds <= 0) return "0s";
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    if (h > 0) return `${h}h ${m}m`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
  }

  // 接收后端推送的类型更新（检测到类型后）
  function handleTypeUpdate(serverType: string) {
    currentType.value = serverType;
  }

  // 接收后端推送的版本更新（检测到版本后）
  function handleVersionUpdate(version: string) {
    currentVersion.value = version;
  }

  onMounted(async () => {
    // 订阅内存信息推送（后端每 2 秒推送一次）
    EventsOn("memory:update", handleMemoryUpdate);
    // 订阅类型/版本检测完成事件
    EventsOn("server:type", handleTypeUpdate);
    EventsOn("server:version", handleVersionUpdate);
    // 加载服务器列表和当前选中服务器
    await Promise.all([loadServerList(), loadActiveServer()]);
    // 活动服务器就绪后，加载其 mod/插件/类型/版本
    await Promise.all([
      loadExtensionsCount(),
      loadTypeAndVersion(),
      refreshUptime(),
    ]);
    // 每秒刷新运行时长
    uptimeTimer = setInterval(refreshUptime, 1000);
  });

  onBeforeUnmount(() => {
    // 组件卸载时取消订阅，避免泄漏
    EventsOff("memory:update");
    EventsOff("server:type");
    EventsOff("server:version");
    // 清理运行时长定时器
    if (uptimeTimer) {
      clearInterval(uptimeTimer);
      uptimeTimer = null;
    }
  });

  return {
    memoryUsagePercent,
    serverList,
    currentServer,
    hasActiveServer,
    currentType,
    currentVersion,
    uptime,
    formatUptime,
    getVersionIcon,
    goToConsole,
    loadServerList,
    setActiveServer,
    modCount,
    pluginCount,
  };
}
