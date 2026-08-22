<script setup lang="ts">
// 模板所需图标（图标仅用于展示，在组件内维护字符串 → 图标组件映射）
import { Puzzle, Blocks, List, AlarmClock, Gamepad, Power, RotateCw, Play, TerminalSquare, Folder, RefreshCw, Server, User } from '@lucide/vue'
import GaugeChart from '../base/GaugeChart/GaugeChart.vue'
import GcChart from '../base/GcChart/GcChart.vue'
import IoChart from '../base/IoChart/IoChart.vue'
// ViewModel：首页看板逻辑
import { useDashboard } from './Dashboard'

const {
    memoryUsagePercent,
    CPUUsagePercent,
    jvmMemoryUsagePercent,
    ioReadMBps,
    ioWriteMBps,
    serverList,
    currentServer,
    hasActiveServer,
    goToConsole,
    loadServerList,
    setActiveServer,
    handleStart,
    handleStop,
    handleRestart,
    infoItems,
    gcPoints,
    playerList,
    isLoadingPlayers,
    refreshPlayerList,
} = useDashboard()

// 左侧信息面板图标映射（字符串 → 组件）
const infoIconMap: Record<string, any> = {
    server: Server,
    gamepad: Gamepad,
    clock: AlarmClock,
    puzzle: Puzzle,
    blocks: Blocks,
    user: User,
}

// 解析左侧面板图标（保持 infoItems 的响应式，通过函数映射）
function resolveInfoIcon(name: string) {
    return infoIconMap[name] || Server
}
</script>
<style scoped src="./Dashboard.css"></style>

<template>
    <div class="dashboard-container">
        <!-- 左侧区域 -->
        <div class="left-wrapper">
            <!-- 左上信息卡片：通过 infoItems 数据驱动渲染图标与文字 -->
            <div class="left-info">
                <div v-for="item in infoItems" :key="item.label" class="info-item stat-item">
                    <div class="info-icon" :style="{ color: item.color }">
                        <component :is="resolveInfoIcon(item.icon)" :size="22" />
                    </div>
                    <div class="stat-value-wrap">
                        <span class="stat-label">{{ item.label }}</span>
                        <span class="stat-value" :style="{ color: item.color }">{{ item.value }}</span>
                    </div>
                </div>
            </div>

            <!-- 左下信息卡片：服务器列表 -->
            <div class="left-bottom-info">
                <!-- 服务器列表：无滚动条的无极滚动 -->
                <div class="server-list-block">
                    <div class="info-item server-list">
                        <div class="info-icon">
                            <List :size="20" />
                        </div>
                        <span class="info-text">服务器列表: {{ serverList.length }}</span>
                        <RefreshCw :size="18" class="refresh-icon" @click="loadServerList" />
                    </div>
                    <!-- 手动滚动容器（隐藏滚动条） -->
                    <div class="server-scroll">
                        <div v-if="serverList.length" class="server-scroll-inner">
                            <div v-for="item in serverList" :key="item.name" class="server-item"
                                :class="{ current: item.name === currentServer }" @click="setActiveServer(item.name)">
                                <Folder :size="16" />
                                <span class="server-name">{{ item.name }}</span>
                                <span v-if="item.name === currentServer" class="server-current-tag">当前</span>
                            </div>
                        </div>
                        <div v-else class="server-empty">暂无服务器</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 无活动服务器时的提示 -->
        <div v-if="!hasActiveServer" class="right-empty">
            <div class="right-empty-icon">⚡</div>
            <span class="right-empty-text">请单击任意服务器为当前服务器</span>
        </div>

        <!-- 右侧区域：仅在存在活动服务器时显示 -->
        <div v-else class="right-wrapper">
            <!-- 状态卡片 -->
            <div class="status-card">
                <div class="performance-box">
                    <!-- CPU 仪表盘 -->
                    <div class="gauge-wrap">
                        <GaugeChart :value="CPUUsagePercent" :max="100" color="#4a9eff" height="160px" />
                        <span class="gauge-title">CPU-P核-使用率%</span>
                    </div>
                    <!-- 内存仪表盘 - 使用动态计算的值 -->
                    <div class="gauge-wrap">
                        <GaugeChart :value="memoryUsagePercent" :max="100" color="#36cfc9" height="160px" />
                        <span class="gauge-title">内存使用率%</span>
                    </div>
                    <!-- JVM 内存使用率仪表盘（后端 GetJVMProcessMemoryUsage 采集） -->
                    <div class="gauge-wrap">
                        <GaugeChart :value="jvmMemoryUsagePercent" :max="100" color="#ffa940" height="160px" />
                        <span class="gauge-title">JVM内存使用率%</span>
                    </div>
                    <!-- 磁盘读写：2 条横向柱状图（读取/写入），替换原 GC 仪表盘 -->
                    <div class="gauge-wrap">
                        <IoChart :read="ioReadMBps" :write="ioWriteMBps" height="160px" />
                        <span class="gauge-title disk-title">磁盘读写</span>
                    </div>
                </div>

                <div class="bottom-box">
                    <div class="tps-box">
                        <span class="tps-title">GC数据</span>
                        <!-- YGC/FGC/GCT 三折线图（图例+悬浮提示标明含义） -->
                        <GcChart :data="gcPoints" height="200px" />
                    </div>
                    <div class="player-box">
                        <div class="player-header">
                            <span class="player-title">玩家</span>
                            <span class="player-count">({{ playerList.length }})</span>
                            <RefreshCw :size="16" class="player-refresh" :class="{ spinning: isLoadingPlayers }"
                                @click="refreshPlayerList" />
                        </div>

                        <!-- 无玩家 -->
                        <div v-if="playerList.length === 0" class="player-empty">
                            <span>暂无玩家在线</span>
                        </div>

                        <!-- 3列网格展示玩家 -->
                        <div v-else class="player-grid">
                            <div v-for="name in playerList" :key="name" class="player-item">
                                <User :size="14" class="player-icon" />
                                <span class="player-name">{{ name }}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- 按钮卡片 -->
            <div class="button-card">
                <button class="action-btn" @click="handleStart">
                    <Play :size="18" />
                    开启
                </button>
                <button class="action-btn" @click="handleRestart">
                    <RotateCw :size="18" />
                    重启
                </button>
                <button class="action-btn" @click="handleStop">
                    <Power :size="18" />
                    停止
                </button>
                <button class="action-btn" @click="goToConsole">
                    <TerminalSquare :size="18" />
                    控制台
                </button>
            </div>
        </div>
    </div>
</template>
