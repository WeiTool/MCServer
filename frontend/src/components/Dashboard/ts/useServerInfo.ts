// ViewModel：首页服务器信息（类型/版本/时长/模组/插件/服务器列表）
// 从 useDashboard 拆出的独立模块，只负责服务器静态信息的加载与展示数据
import { ref, computed } from "vue";
import { useActiveServer } from "../../../composables/useActiveServer";
import { serverApi, processApi } from "../../../api";
import { formatUptime } from "../../../utils/formatUptime.js";
import type { ServerInstance } from "./types";

/**
 * useServerInfo
 * 首页左侧信息面板与服务器列表相关状态。
 * 只读服务器信息，不包含进程控制（见 useProcessControl）与实时指标（见 useRealtimeStats）。
 */
export function useServerInfo() {
  // ---------- 服务器类型与版本 ----------
  const currentType = ref("");
  const currentVersion = ref("");

  async function loadTypeAndVersion() {
    const { currentServer } = useActiveServer();
    if (!currentServer.value) {
      currentType.value = "";
      currentVersion.value = "";
      return;
    }
    const [type, version] = await Promise.all([
      serverApi.fetchServerType(currentServer.value),
      serverApi.fetchServerVersion(currentServer.value),
    ]);
    currentType.value = type;
    currentVersion.value = version;
  }

  function handleTypeUpdate(serverType: string) {
    currentType.value = serverType;
  }

  // ---------- 服务器运行时长 ----------
  const uptime = ref(0);

  async function refreshUptime() {
    uptime.value = await processApi.fetchServerUptime();
  }

  // ---------- 模组与插件统计 ----------
  const modCount = ref(0);
  const pluginCount = ref(0);

  async function loadExtensionsCount() {
    const { currentServer } = useActiveServer();
    if (!currentServer.value) {
      modCount.value = 0;
      pluginCount.value = 0;
      return;
    }
    const [mods, plugins] = await Promise.all([
      serverApi.fetchServerModCount(currentServer.value),
      serverApi.fetchServerPluginCount(currentServer.value),
    ]);
    modCount.value = mods;
    pluginCount.value = plugins;
  }

  // ---------- 服务器列表 ----------
  const serverList = ref<ServerInstance[]>([]);

  async function loadServerList() {
    const res = await serverApi.fetchServerList();
    serverList.value = res?.servers || [];
  }

  // ---------- 左侧信息面板 ----------
  const infoItems = computed(() => [
    {
      icon: "server",
      label: "类型",
      value: currentType.value || "未知",
      color: "#4a9eff",
    },
    {
      icon: "gamepad",
      label: "版本",
      value: currentVersion.value || "未知",
      color: "#36cfc9",
    },
    {
      icon: "clock",
      label: "运行时长",
      value: formatUptime(uptime.value),
      color: "#73d13d",
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
  ]);

  return {
    currentType,
    currentVersion,
    uptime,
    modCount,
    pluginCount,
    serverList,
    loadTypeAndVersion,
    handleTypeUpdate,
    refreshUptime,
    loadExtensionsCount,
    loadServerList,
    infoItems,
  };
}
