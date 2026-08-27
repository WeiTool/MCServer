<script setup lang="ts">
// Settings 壳组件：侧边导航 + 内容区
// 各导航项按文件夹拆分：ServerProperties/（基础配置）等。
// 新增导航项时：在 ServerProperties 同级建文件夹 + 在 menuOptions 增加菜单项 + 在内容区挂组件
import type { MenuOption } from 'naive-ui'
import type { Component } from 'vue'
import { BookOutline as BookIcon, SettingsOutline as SettingsIcon } from '@vicons/ionicons5'
import { NIcon, NConfigProvider } from 'naive-ui'
import { ref, h } from 'vue'
import { useActiveServer } from "../../composables/useActiveServer";
import ServerProperties from './ServerProperties/ServerProperties.vue'
import UpdateSettings from './UpdateSettings.vue'

// ---------- 共享状态 ----------
const { currentServer } = useActiveServer();

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// --------- 侧边菜单配置 ----------
// 每个导航项对应一个独立组件/文件夹
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
  {
    label: '更新设置',
    key: 'update-settings',
    icon: renderIcon(SettingsIcon)
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

        <!-- 右侧内容区：按导航项渲染对应组件 -->
        <n-layout style="position: relative; overflow-y: auto;">
          <div style="padding: 20px;">
            <!-- 基础配置（server.properties） -->
            <ServerProperties v-if="activeKey === 'basic-config'" />
            <!-- 更新设置（预览版开关） -->
            <UpdateSettings v-if="activeKey === 'update-settings'" />
          </div>
        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>
