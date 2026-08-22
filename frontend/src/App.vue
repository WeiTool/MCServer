<script lang="ts" setup>
// 全局导航栏
import MainNav from './components/MainNav/MainNav.vue'
</script>

<template>
  <!-- n-message-provider -->
    <n-message-provider>
      <!-- 顶部导航栏 -->
      <MainNav />
      <!-- 路由出口：根据路由渲染对应页面 -->
      <div class="content">
        <router-view />
      </div>
    </n-message-provider>
</template>

<style>

</style>