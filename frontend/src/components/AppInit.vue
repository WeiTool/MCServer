<script setup lang="ts">
// 全局初始化：应用启动时检查并自动下载更新（纯逻辑组件，无 UI）
// 必须渲染在 <n-dialog-provider> 内部，useDialog 才能工作
import { onMounted, watch } from 'vue'
import { useDialog } from 'naive-ui'
import { GetUpdateState } from '../../wailsjs/go/main/App'
import { useVersionCheck } from '../composables/useVersionCheck'

const dialog = useDialog()

const { checkVersion, downloadUpdate, downloadComplete } = useVersionCheck()

onMounted(async () => {
    // 1. 检查上次更新结果（config/update.json：updated=已更新 / error=替换失败）
    let updateState = null
    try {
        updateState = await GetUpdateState()
    } catch {
        // 文件不存在或读取失败，静默忽略（无历史状态）
    }

    if (updateState?.status === 'updated') {
        dialog.success({
            title: '更新完成',
            content: `已成功更新至 v${updateState.version}，感谢使用！`,
            positiveText: '知道了',
            draggable: true,
        })
    } else if (updateState?.status === 'error') {
        dialog.error({
            title: '更新失败',
            content: updateState.error || '替换失败，请手动重新下载',
            positiveText: '确定',
            draggable: true,
        })
    }

    // 2. 检查新版本：有更新则自动下载（不打扰用户，结果下次打开提示）
    try {
        const result = await checkVersion()
        if (result?.hasUpdate && result.versionInfo?.latest) {
            await downloadUpdate() // 等待下载完成
            // 如果成功，downloadComplete 会被 watch 捕获，无需额外操作
        }
    } catch (err) {
        // 显示错误对话框
        dialog.error({
            title: '更新失败',
            content: err instanceof Error ? err.message : '版本检查或下载失败，请稍后重试',
        })
    }
})

// 监听下载完成：提示用户重启即可完成更新
watch(downloadComplete, (done) => {
    if (done) {
        dialog.success({
            title: '下载完成',
            content: '新版本已下载完成，重启即可更新。',
            positiveText: '知道了',
            draggable: true,
        })
        // 重置，防止重复触发
        downloadComplete.value = false
    }
})
</script>

<template>
    <!-- 纯逻辑组件，不渲染任何 UI -->
</template>
