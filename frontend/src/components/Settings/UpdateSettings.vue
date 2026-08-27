<script setup lang="ts">
// 更新设置：是否获取预览版（beta）更新，持久化到 config/global_config.json
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { GetGlobalConfig, SaveGlobalConfig } from '../../../wailsjs/go/main/App'

const message = useMessage()

const previewEnabled = ref(false)
const saving = ref(false)

// 进入页面时读取当前配置
onMounted(async () => {
    try {
        const cfg = await GetGlobalConfig()
        previewEnabled.value = cfg.previewEnabled
    } catch {
        message.error('读取更新配置失败')
    }
})

// 切换开关时保存配置（失败则回滚）
async function onToggle(val: boolean) {
    saving.value = true
    try {
        await SaveGlobalConfig({ previewEnabled: val })
        message.success(val ? '已开启预览版更新' : '已关闭预览版更新')
    } catch {
        previewEnabled.value = !val
        message.error('保存更新配置失败')
    } finally {
        saving.value = false
    }
}
</script>

<template>
    <div class="update-settings">
        <h3 class="section-title">更新设置</h3>

        <div class="setting-item">
            <div class="setting-info">
                <span class="setting-label">获取预览版</span>
                <span class="setting-desc">开启后将检查测试版（beta）更新，包含未正式发布的版本</span>
            </div>
            <n-switch v-model:value="previewEnabled" :loading="saving" @update:value="onToggle" />
        </div>
    </div>
</template>

<style scoped>
.update-settings {
    max-width: 640px;
}

.section-title {
    margin: 0 0 16px;
    font-size: 16px;
}

.setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    border: 1px solid #e5e5e5;
    border-radius: 8px;
}

.setting-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.setting-label {
    font-size: 14px;
    font-weight: 500;
}

.setting-desc {
    font-size: 12px;
    color: #999;
}
</style>
