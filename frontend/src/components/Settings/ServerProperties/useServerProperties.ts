// ViewModel：基础配置（server.properties）数据加载与保存
// Settings 中"基础配置"导航项的核心逻辑：拉取/修改/保存服务器属性
import { ref, computed } from "vue";
import { useMessage } from "naive-ui";
import { GetServerProperties, SetServerProperties } from "../../../../wailsjs/go/main/App";
import { useActiveServer } from "../../../composables/useActiveServer";
import { formatProperties, commonPropertyKeys, getPropertyLabel, translatePropertyValue } from "../../../utils/serverProperties";

/**
 * useServerProperties
 * 基础配置页数据层：加载服务器 server.properties、
 * 内存中修改单个属性、整体保存回后端。
 */
export function useServerProperties() {
  const { currentServer } = useActiveServer();
  const message = useMessage();

  /** server.properties 原始键值（null = 未加载/无数据） */
  const serverProperties = ref<Record<string, string> | null>(null);
  const loading = ref(false);

  /** 从后端加载当前服务器的 server.properties */
  async function loadServerProperties() {
    if (!currentServer.value) {
      serverProperties.value = null
      return
    }

    if (loading.value) return
    loading.value = true

    GetServerProperties(currentServer.value)
      .then((properties) => {
        serverProperties.value = properties
      })
      .catch(() => {
        // 加载失败对用户可见，用 Message 通知
        serverProperties.value = null
        message.error('获取服务器配置失败')
      })
      .finally(() => {
        loading.value = false
      })
  }

  /** 更新单个属性（只改内存，不调后端） */
  function updateServerProperty(key: string, value: string) {
    if (serverProperties.value) {
      serverProperties.value[key] = value
    }
  }

  /** 全部保存回后端 */
  async function saveAllProperties() {
    if (!currentServer.value || !serverProperties.value) return

    try {
      await SetServerProperties(currentServer.value, serverProperties.value)
      message.success('保存成功')
    } catch (error) {
      message.error(`保存失败: ${error}`)
    }
  }

  // ---------- 配置展示映射 ----------
  /** 全部属性（格式化后的展示数组） */
  const allProps = computed(() => formatProperties(serverProperties.value))

  /** 常用配置（按常用键列表提取 + 翻译值） */
  const commonProps = computed(() => {
    const props = serverProperties.value
    if (!props) return []
    return commonPropertyKeys.map(key => ({
      key,
      label: getPropertyLabel(key),
      value: translatePropertyValue(key, props[key] || '')
    }))
  })

  return {
    serverProperties,
    loading,
    allProps,
    commonProps,
    loadServerProperties,
    updateServerProperty,
    saveAllProperties,
  };
}
