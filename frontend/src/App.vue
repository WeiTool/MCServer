<script lang="ts" setup>
// 全局导航栏
import MainNav from './components/MainNav/MainNav.vue'
</script>

<template>
  <!-- 需要同时包裹 n-dialog-provider 和 n-message-provider -->
  <n-dialog-provider>
    <n-message-provider>
      <!-- 顶部导航栏 -->
      <MainNav />
      <!-- 路由出口：根据路由渲染对应页面 -->
      <div class="content">
        <router-view />
      </div>
    </n-message-provider>
  </n-dialog-provider>
</template>

<style>

</style>