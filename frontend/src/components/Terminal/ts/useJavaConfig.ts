// ViewModel：控制台 Java 配置（扫描列表 / 当前服务器选择 / 自定义添加）
// 从 useTerminal 拆出的独立模块，负责 Java 下拉菜单相关逻辑
import { computed } from "vue";
import { storeToRefs } from "pinia";
import { useMessage, type DropdownOption } from "naive-ui";
import { useActiveServer } from "../../../composables/useActiveServer";
import {
  useTerminalStore,
  type JavaInfo,
} from "../../../stores/terminal";
import { javaApi } from "../../../api";

/**
 * useJavaConfig
 * Java 下拉菜单：扫描系统 Java 列表、加载当前服务器已配置的 Java、
 * 处理选择与"添加自定义 Java"动作。
 */
export function useJavaConfig() {
  const message = useMessage();
  const { currentServer } = useActiveServer();
  const store = useTerminalStore();
  const { javaList, selectedJava } = storeToRefs(store);

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

  /** 生成 Java 的可读显示名称 */
  function javaDisplayName(java: JavaInfo): string {
    return `Java ${java.version} (${java.versionName})`;
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

  /** 加载当前服务器的内存配置（后端存储 MB，前端显示 GB） */
  async function loadServerMemory() {
    if (!currentServer.value) return;
    const [xmxMB, xmsMB] = await javaApi.getServerMemory(currentServer.value);
    if (xmxMB > 0) store.setXmx(xmxMB / 1024);
    if (xmsMB > 0) store.setXms(xmsMB / 1024);
  }

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

  return {
    loadJavaList,
    loadServerJava,
    loadServerMemory,
    javaDropdownOptions,
    selectedJavaDisplay,
    handleJavaSelect,
  };
}
