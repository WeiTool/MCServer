<script setup lang="ts">
// 模板所需图标（仅用于展示，在组件内导入）
import { LayoutDashboard, TerminalSquare } from '@lucide/vue';
// ViewModel：导航栏逻辑
import { useMainNav } from './MainNav'

const { activeKey, handleNav, handleMinimize, handleClose } = useMainNav()
</script>
<style scoped src="./MainNav.css"></style>

<template>
  <!-- 
    导航栏容器
    - 添加 .window-draggable 类启用窗口拖拽
    - 整个导航栏区域都可拖拽（除了按钮等交互元素）
  -->
  <nav class="main-nav window-draggable">
    <!-- Logo 区域 -->
    <div class="logo">
      <p>MCServer</p>
    </div>

    <!-- 导航菜单 -->
    <ul class="nav-menu">
      <li 
        class="nav-item" 
        :class="{ active: activeKey === 'home' }"
        @click="handleNav('home')"
      >
        <LayoutDashboard />
        <span>首页</span>
      </li>
      <li 
        class="nav-item" 
        :class="{ active: activeKey === 'console' }"
        @click="handleNav('console')"
      >
        <TerminalSquare />
        <span>控制台</span>
      </li>
    </ul>

    <!-- 右侧按钮组 -->
    <div class="buttons">
      <button class="min-btn" @click="handleMinimize" aria-label="最小化">−</button>
      <button class="close-btn" @click="handleClose" aria-label="关闭">✕</button>
    </div>
  </nav>
</template>
