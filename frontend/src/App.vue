<script lang="ts" setup>
// 全局导航栏
import MainNav from './components/MainNav/MainNav.vue'
// 全局初始化：版本检测等（必须位于 dialog provider 内部）
import AppInit from './components/AppInit.vue'
</script>

<template>
  <n-notification-provider>
    <n-message-provider>
      <n-dialog-provider>
        <AppInit />
        <MainNav />
        <div class="content">
          <router-view />
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-notification-provider>
</template>

<style>

</style>
