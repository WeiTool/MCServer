<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import type { Component } from 'vue'
import { BookOutline as BookIcon, Save as SaveIcon } from '@vicons/ionicons5'
import { NIcon, NConfigProvider, useMessage } from 'naive-ui'
import { ref, onMounted, watch, h, computed } from 'vue'
import { GetServerProperties, SetServerProperties } from "../../../wailsjs/go/main/App";
import { useActiveServer } from "../../composables/useActiveServer";
import { formatProperties, gamemodeMap, difficultyMap, levelTypeMap, commonPropertyKeys, getPropertyLabel, translatePropertyValue } from '../../utils/serverProperties'

// ---------- 共享状态 ----------
const { currentServer } = useActiveServer();

// 初始化 message 实例
const message = useMessage()

// ---------- 生命周期 ----------
onMounted(() => {
  loadServerProperties()
})

watch(currentServer, () => {
  loadServerProperties()
})

// --------- 文件配置 ----------
const serverProperties = ref<Record<string, string> | null>(null)
const loading = ref(false)

function loadServerProperties() {
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

function updateServerProperty(key: string, value: string) {
  // 只更新内存中的值，不调后端
  if (serverProperties.value) {
    serverProperties.value[key] = value
  }
}

async function saveAllProperties() {
  if (!currentServer.value || !serverProperties.value) return

  try {
    await SetServerProperties(currentServer.value, serverProperties.value)
    message.success('保存成功')
  } catch (error) {
    message.error(`保存失败: ${error}`)
  }
}

// --------- 文件配置映射 ----------
const allProps = computed(() => formatProperties(serverProperties.value))

// 常用配置
const commonProps = computed(() => {
  const props = serverProperties.value  // 先赋值给局部变量
  if (!props) return []                 // 判断局部变量
  return commonPropertyKeys.map(key => ({
    key,
    label: getPropertyLabel(key),
    value: translatePropertyValue(key, props[key] || '')  // 使用局部变量
  }))
})

// 判断值类型
function getValueType(key: string, value: string): 'text' | 'select' | 'switch' {
  // 下拉选择类型
  if (key === 'gamemode') return 'select'
  if (key === 'difficulty') return 'select'
  if (key === 'level-type') return 'select'

  // 开关类型（布尔值）
  if (value === 'true' || value === 'false') return 'switch'

  // 普通文本
  return 'text'
}

// 获取下拉选项
function getSelectOptions(key: string): Array<{ label: string; value: string }> {
  if (key === 'gamemode') {
    return Object.entries(gamemodeMap).map(([value, label]) => ({ label, value }))
  }
  if (key === 'difficulty') {
    return Object.entries(difficultyMap).map(([value, label]) => ({ label, value }))
  }
  if (key === 'level-type') {
    return Object.entries(levelTypeMap).map(([value, label]) => ({ label, value }))
  }
  return []
}

// 获取开关状态
function getSwitchValue(value: string): boolean {
  return value === 'true'
}

// 判断是否为长文本（需要更宽的输入框）
function isLongText(key: string): boolean {
  const longTextKeys = ['motd', 'resource-pack', 'generator-settings', 'level-seed']
  return longTextKeys.includes(key)
}

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// --------- 侧边菜单配置 ----------
const menuOptions: MenuOption[] = [
  {
    label: `当前服务器:${currentServer.value || '无'}`,
    disabled: true,
  },
  {
    label: '基础配置',
    key: 'basic-config',
    icon: renderIcon(BookIcon)
  },
]

// 默认选中左侧第一个选项
const activeKey = ref<string | null>('basic-config')
const collapsed = ref(false)

</script>

<template>
  <n-config-provider>
    <n-message-provider>
      <n-layout has-sider style="height: 91.5vh">
        <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="240" :collapsed="collapsed"
          show-trigger @collapse="collapsed = true" @expand="collapsed = false">
          <n-menu v-model:value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="22"
            :options="menuOptions" />
        </n-layout-sider>

        <!-- 右侧内容区 -->
        <n-layout style="position: relative; overflow-y: auto;">
          <div style="padding: 20px;">
            <div v-if="loading" style="text-align: center; padding: 40px;">
              <n-spin size="large" />
              <p>加载配置中...</p>
            </div>

            <div v-else-if="!serverProperties" style="text-align: center; padding: 40px; color: #999;">
              <n-empty description="暂无配置数据，请先启动服务器" />
            </div>

            <div v-else>
              <!-- 常用配置卡片 - 2列4行 -->
              <div class="common-config-wrapper">
                <n-h3 prefix="bar" type="success">常用配置</n-h3>
                <div class="common-config-card">
                  <div v-for="(item, index) in commonProps" :key="item.key"
                    class="common-config-item"
                    :class="{
                      'border-right': (index + 1) % 2 !== 0 && index < commonProps.length - 1,
                      'border-bottom': index < commonProps.length - 2
                    }">
                    <div class="common-config-label">{{ item.label }}</div>
                    <div class="common-config-control">
                      <n-select v-if="getValueType(item.key, item.value) === 'select'"
                        :value="item.value"
                        :options="getSelectOptions(item.key)"
                        size="small"
                        style="width: 140px"
                        @update:value="(val: string) => updateServerProperty(item.key, val)" />

                      <n-switch v-else-if="getValueType(item.key, item.value) === 'switch'"
                        :value="getSwitchValue(item.value)"
                        size="small"
                        @update:value="(val: boolean) => updateServerProperty(item.key, String(val))" />

                      <n-input v-else
                        :value="item.value"
                        size="small"
                        style="width: 140px"
                        @update:value="(val: string) => updateServerProperty(item.key, val)" />
                    </div>
                  </div>
                </div>
              </div>

              <n-space vertical size="large">
                <template v-if="activeKey === 'basic-config'">
                  <n-h3 prefix="bar" type="info">全部配置</n-h3>

                  <n-descriptions :column="2" bordered label-placement="left" label-align="right" :label-style="{
                    width: '300px',
                    fontWeight: 'bold',
                    paddingRight: '16px',
                    verticalAlign: 'middle'
                  }" :content-style="{
                    textAlign: 'left'
                  }">
                    <n-descriptions-item v-for="item in allProps" :key="item.key" :label="item.label" :content-style="{
                      padding: '6px 16px',
                    }">
                      <div style="display: flex; align-items: center; min-height: 32px;">

                        <n-select v-if="getValueType(item.key, item.value) === 'select'" :value="item.value"
                          :options="getSelectOptions(item.key)" size="small" style="width: 200px"
                          @update:value="(val: string) => updateServerProperty(item.key, val)" />

                        <div v-else-if="getValueType(item.key, item.value) === 'switch'"
                          style="display: flex; justify-content: center; width: 100%;">
                          <n-switch :value="getSwitchValue(item.value)" size="small"
                            @update:value="(val: boolean) => updateServerProperty(item.key, String(val))" />
                        </div>

                        <n-input v-else-if="isLongText(item.key)" :value="item.value" size="small"
                          style="width: 100%; max-width: 500px"
                          @update:value="(val: string) => updateServerProperty(item.key, val)" />

                        <n-input v-else :value="item.value" size="small" style="width: 200px"
                          @update:value="(val: string) => updateServerProperty(item.key, val)" />

                      </div>
                    </n-descriptions-item>
                  </n-descriptions>
                </template>
              </n-space>
            </div>
          </div>

          <!-- 原生 Vue 悬浮按钮 -->
          <div v-if="activeKey === 'basic-config' && !loading && serverProperties" class="save-float-btn"
            @click="saveAllProperties" title="保存基础配置">
            <n-icon size="28">
              <SaveIcon />
            </n-icon>
          </div>

        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
/* 常用配置卡片样式 - 2列4行 */
.common-config-wrapper {
  margin-bottom: 24px;
}

.common-config-card {
  background-color: var(--n-color, #ffffff);
  border-radius: 8px;
  border: 1px solid var(--n-border-color, #e5e5e5);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  display: grid;
  grid-template-columns: 1fr 1fr;
  overflow: hidden;
}

.common-config-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 20px;
  min-height: 52px;
}

/* 右侧边框线（奇数列，且不是最后一个） */
.common-config-item.border-right {
  border-right: 1px solid var(--n-border-color, #f0f0f0);
}

/* 底部边框线（前6个，即第1、2行，因为总共4行） */
.common-config-item.border-bottom {
  border-bottom: 1px solid var(--n-border-color, #f0f0f0);
}

.common-config-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--n-text-color, #333333);
  flex-shrink: 0;
  margin-right: 12px;
  white-space: nowrap;
}

.common-config-control {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

/* 原生悬浮按钮样式 */
.save-float-btn {
  position: absolute;
  left: 24px;
  bottom: 15px;
  width: 60px;
  height: 60px;
  background-color: #18a058;
  color: white;
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(24, 160, 88, 0.3);
  transition: all 0.2s ease;
  z-index: 100;
}

.save-float-btn:hover {
  background-color: #36ad6a;
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(24, 160, 88, 0.4);
}

.save-float-btn:active {
  transform: scale(0.95);
}
</style>