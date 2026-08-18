<script setup lang="ts">
// 模板所需图标（图标仅用于展示，不属于业务逻辑，故在组件内导入）
import { Puzzle, Blocks, List, AlarmClock, Gamepad, Power, RotateCw, Play, TerminalSquare, Folder, RefreshCw } from '@lucide/vue'
import GaugeChart from '../base/GaugeChart/GaugeChart.vue'
import TpsChart from '../base/TpsChart/TpsChart.vue'
// ViewModel：首页看板逻辑
import { useDashboard } from './Dashboard'

const {
    memoryUsagePercent,
    serverList,
    currentServer,
    hasActiveServer,
    currentType,
    currentVersion,
    uptime,
    formatUptime,
    getVersionIcon,
    goToConsole,
    loadServerList,
    setActiveServer,
    modCount,
    pluginCount,
} = useDashboard()
</script>
<style scoped src="./Dashboard.css"></style>

<template>
    <div class="dashboard-container">
        <!-- 左侧区域 -->
        <div class="left-wrapper">
            <!-- 左上信息卡片 -->
            <div class="left-info">
                <!-- 服务器类型 -->
                <div class="info-item server-version">
                    <div class="info-icon server-icon">
                        <img v-if="typeof getVersionIcon(currentType) === 'string'"
                            :src="getVersionIcon(currentType) as string" />
                        <component v-else :is="getVersionIcon(currentType)" :size="20" />
                    </div>
                    <span class="info-text">类型:{{ currentType || '未知' }}</span>
                </div>
                <!-- Minecraft版本 -->
                <div class="info-item mc-version">
                    <div class="info-icon mc-icon">
                        <Gamepad :size="20" />
                    </div>
                    <span class="info-text">版本:{{ currentVersion || '未知' }}</span>
                </div>
                <!-- 运行时长 -->
                <div class="info-item server-time">
                    <div class="info-icon">
                        <AlarmClock :size="20" />
                    </div>
                    <span class="info-text">运行时长:{{ formatUptime(uptime) }}</span>
                </div>
            </div>

            <!-- 左下信息卡片 -->
            <div class="left-bottom-info">
                <!-- 模组数量 -->
                <div class="info-item stat-item">
                    <div class="info-icon">
                        <Puzzle :size="22" color="#ff6b6b" />
                    </div>
                    <div class="stat-value-wrap">
                        <span class="stat-label">模组数量</span>
                        <span class="stat-value" style="color:#ff6b6b">{{ modCount }}</span>
                    </div>
                </div>
                <!-- 插件数量 -->
                <div class="info-item stat-item">
                    <div class="info-icon">
                        <Blocks :size="22" color="#b37feb" />
                    </div>
                    <div class="stat-value-wrap">
                        <span class="stat-label">插件数量</span>
                        <span class="stat-value" style="color:#b37feb">{{ pluginCount }}</span>
                    </div>
                </div>
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
                        <GaugeChart :value="55" :max="100" color="#4a9eff" height="160px" />
                        <span class="gauge-title">CPU使用率%</span>
                    </div>
                    <!-- 内存仪表盘 - 使用动态计算的值 -->
                    <div class="gauge-wrap">
                        <GaugeChart :value="memoryUsagePercent" :max="100" color="#36cfc9" height="160px" />
                        <span class="gauge-title">内存使用率%</span>
                    </div>
                    <!-- 延迟仪表盘 -->
                    <div class="gauge-wrap">
                        <GaugeChart :value="30" :max="200" color="#ffa940" height="160px" />
                        <span class="gauge-title">延迟</span>
                    </div>
                    <!-- GC 仪表盘 -->
                    <div class="gauge-wrap">
                        <GaugeChart :value="85" :max="100" color="#ff6b6b" height="160px" />
                        <span class="gauge-title">GC频率</span>
                    </div>
                </div>

                <div class="bottom-box">
                    <div class="tps-box">
                        <span class="tps-title">TPS</span>
                        <!-- TPS 折线图 -->
                        <TpsChart :data="[20, 18, 19, 20, 17, 20, 19, 18, 20, 19, 18, 20, 17, 20, 18]" title="TPS"
                            height="200px" />
                    </div>
                    <div class="player-box">
                        <span class="player-title">玩家</span>
                        <div class="player-placeholder">暂无数据</div>
                    </div>
                </div>
            </div>

            <!-- 按钮卡片 -->
            <div class="button-card">
                <button class="action-btn">
                    <Play :size="18" />
                    开启
                </button>
                <button class="action-btn">
                    <RotateCw :size="18" />
                    重启
                </button>
                <button class="action-btn">
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
