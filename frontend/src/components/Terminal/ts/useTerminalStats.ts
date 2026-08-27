// ViewModel：控制台服务器统计（运行时长/模组插件/信息面板/资源推送）
// 从 useTerminal 拆出的独立模块，负责服务器统计信息的刷新与信息面板组装
import { computed, type Ref } from "vue";
import { storeToRefs } from "pinia";
import { useActiveServer } from "../../../composables/useActiveServer";
import {
  useTerminalStore,
  type SystemInfo,
} from "../../../stores/terminal";
import { processApi, serverApi } from "../../../api";
import { formatUptime } from "../../../utils/formatUptime.js";

/**
 * useTerminalStats
 * 服务器统计：运行时长（每秒刷新）、模组/插件数量，以及 CPU/内存推送事件到 store。
 * 信息面板依赖在线/最大玩家数（usePlayerPolling 提供），经 useTerminal 注入。
 */
export function useTerminalStats(players: {
  onlinePlayers: Ref<number>;
  maxPlayers: Ref<number>;
}) {
  const { currentServer } = useActiveServer();
  const store = useTerminalStore();
  const { modCount, pluginCount, uptime, cpuUsagePercent, memoryUsagePercent } =
    storeToRefs(store);

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

  // ---------- 左侧信息面板 ----------
  const infoItems = computed(() => [
    { icon: "cpu", label: "P核-使用率", value: `${cpuUsagePercent.value}%`, color: "#4a9eff" },
    { icon: "memory", label: "内存使用", value: `${memoryUsagePercent.value}%`, color: "#36cfc9" },
    {
      icon: "users",
      label: "在线玩家",
      value: `${players.onlinePlayers.value}/${players.maxPlayers.value}`,
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

  return {
    refreshUptime,
    loadModCount,
    handleCPUUpdate,
    handleMemoryUpdate,
    infoItems,
  };
}
