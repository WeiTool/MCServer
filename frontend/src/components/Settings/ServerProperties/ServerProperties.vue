<script setup lang="ts">
// 基础配置（server.properties）页
// 数据逻辑拆分到 useServerProperties / propertyControls 两个 ts 模块
import { onMounted, watch } from 'vue'
import { Save as SaveIcon } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { useServerProperties } from './useServerProperties'
import { getValueType, getSelectOptions, getSwitchValue, isLongText } from './propertyControls'
import { useActiveServer } from '../../../composables/useActiveServer'

const { currentServer } = useActiveServer()
const { serverProperties, loading, allProps, commonProps, loadServerProperties, updateServerProperty, saveAllProperties } = useServerProperties()

onMounted(() => {
  loadServerProperties()
})

watch(currentServer, () => {
  loadServerProperties()
})
</script>

<template>
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
    </n-space>

    <!-- 原生 Vue 悬浮按钮 -->
    <div class="save-float-btn" @click="saveAllProperties" title="保存基础配置">
      <n-icon size="28">
        <SaveIcon />
      </n-icon>
    </div>
  </div>
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
