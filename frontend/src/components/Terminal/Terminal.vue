<script setup lang="ts">
// 模板所需图标（图标仅用于展示，在组件内维护字符串 → 图标组件映射）
import { Cpu, MemoryStick, Users, Puzzle, Blocks, Clock, Play, Power, RotateCw, FileText, AlertTriangle, Send, FolderOpen, Coffee, Trash2 } from '@lucide/vue'
// ViewModel：控制台逻辑
import { useTerminal } from './Terminal'

const {
  infoItems,
  terminalLines,
  terminalScreenRef,
  commandInput,
  selectedJavaDisplay,
  javaDropdownOptions,
  xmxGB,
  xmsGB,
  currentServer,
  hasActiveServer,
  sendCommand,
  handleAction,
  handleJavaSelect,
  actionButtons,
} = useTerminal()

// terminalScreenRef 通过模板 ref="terminalScreenRef" 绑定到终端滚动容器，
// 滚动逻辑在 Terminal.ts 的 appendLog 中复用；此处显式标记以消除 lint 的"未读取"警告
void terminalScreenRef

// 左侧信息面板图标映射（字符串 → 组件）
const infoIconMap: Record<string, any> = {
  cpu: Cpu,
  memory: MemoryStick,
  users: Users,
  puzzle: Puzzle,
  blocks: Blocks,
  clock: Clock,
}

// 右侧功能按钮图标映射（按 icon 标识）
const actionIconMap: Record<string, any> = {
  play: Play,
  power: Power,
  rotate: RotateCw,
  file: FileText,
  alert: AlertTriangle,
  trash: Trash2,
}

// 解析左侧面板图标（保持 infoItems 的响应式，通过函数映射）
function resolveInfoIcon(name: string) {
  return infoIconMap[name] || Cpu
}

// 解析右侧功能按钮图标（按 icon 标识映射）
function resolveActionIcon(icon: string) {
  return actionIconMap[icon] || Play
}
</script>
<style scoped src="./Terminal.css"></style>

<template>
  <!-- 无活动服务器时的提示 -->
  <div v-if="!hasActiveServer" class="console-empty">
    <span class="console-empty-text">请单击任意服务器为当前服务器</span>
  </div>

  <!-- 有活动服务器时的控制台主体 -->
  <div v-else class="console-page">
    <!-- 左侧信息面板 -->
    <div class="left-panel">
      <div class="info-card">
        <div v-for="item in infoItems" :key="item.label" class="info-row">
          <div class="info-icon" :style="{ color: item.color }">
            <component :is="resolveInfoIcon(item.icon)" :size="20" />
          </div>
          <div class="info-text">
            <span class="info-label">{{ item.label }}</span>
            <span class="info-value" :style="{ color: item.color }">{{ item.value }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 中间终端区域 -->
    <div class="center-panel">
      <div class="terminal-wrapper">
        <!-- 当前服务器信息栏 -->
        <div class="terminal-header">
          <FolderOpen :size="14" />
          <span class="terminal-current-server">
            {{ currentServer ? currentServer : '未选择服务器' }}
          </span>
        </div>
        <div ref="terminalScreenRef" class="terminal-screen">
          <!-- v-memo：内容未变的日志行跳过重渲染，10000 行上限下追加日志只渲染新增行 -->
          <div v-for="(line, i) in terminalLines" :key="i" v-memo="[line.text, line.level]"
            class="terminal-line" :class="`log-${line.level}`">
            {{ line.text }}
          </div>
        </div>
      </div>
      <!-- 命令输入栏：独立白色卡片，与黑色终端区域隔开 -->
      <div class="terminal-input-bar">
        <input v-model="commandInput" class="command-input" placeholder="输入命令..." @keydown.enter="sendCommand" />
        <button class="send-btn" @click="sendCommand">
          <Send :size="16" />
          发送
        </button>
      </div>
    </div>

    <!-- 右侧统一卡片：用分割线区分功能区和配置区 -->
    <div class="side-card">
      <!-- 功能区域 -->
      <div class="side-section">
        <div class="section-title">
          <Play :size="14" />
          <span>功能</span>
        </div>
        <div class="action-btns">
          <button v-for="btn in actionButtons" :key="btn.label" class="action-btn" :class="btn.type"
            @click="handleAction(btn.action)">
            <component :is="resolveActionIcon(btn.icon)" :size="16" />
            <span>{{ btn.label }}</span>
          </button>
        </div>
      </div>

      <!-- 分割线 -->
      <div class="section-divider"></div>

      <!-- 配置区域 -->
      <div class="side-section">
        <div class="section-title">
          <Coffee :size="14" />
          <span>配置</span>
        </div>

        <!-- Java 下拉选择 -->
        <n-dropdown trigger="click" :options="javaDropdownOptions" @select="handleJavaSelect">
          <n-button block :loading="false">
            {{ selectedJavaDisplay }}
          </n-button>
        </n-dropdown>

        <!-- 内存配置（GB） -->
        <div class="memory-config">
          <div class="memory-item">
            <span class="memory-label">最大内存</span>
            <div class="memory-input-wrap">
              <n-input-number v-model:value="xmxGB" :min="1" :max="64" :step="1" size="small" />
              <span class="memory-unit">GB</span>
            </div>
          </div>
          <div class="memory-item">
            <span class="memory-label">最小内存</span>
            <div class="memory-input-wrap">
              <n-input-number v-model:value="xmsGB" :min="1" :max="64" :step="1" size="small" />
              <span class="memory-unit">GB</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>