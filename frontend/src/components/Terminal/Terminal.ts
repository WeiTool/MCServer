// ViewModel：控制台（终端）业务逻辑（组合式函数）
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useMessage, useDialog, type DropdownOption } from 'naive-ui'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
// ViewModel：活动服务器共享状态
import { useActiveServer } from '../../composables/useActiveServer'
// Model：Java、进程与服务器 API
import { javaApi, processApi, serverApi } from '../../api'
// 工具：日志级别判断
import { detectLogLevel, type LogLine } from '../../utils/log'

/**
 * useTerminal
 * 控制台（终端）业务逻辑。返回模板所需的所有状态、数据与交互方法。
 */
export function useTerminal() {
    // naive-ui 消息提示
    const message = useMessage()
    // naive-ui 对话框（用于"是否输入 /version"确认）
    const dialog = useDialog()

    // Java 版本映射（Java 8 显示 1.8 等，17+ 直接显示 17）
    function javaDisplayName(java: { path: string; version: number; versionName: string }) {
        return `Java ${java.version} (${java.versionName})`
    }

    // 活动服务器状态（共享 ViewModel）
    const { currentServer, hasActiveServer, loadActiveServer } = useActiveServer()

    // mod 数量（由后端 ServerInfo 统计，动态刷新）
    const modCount = ref(0)
    // 插件数量（由后端 ServerInfo 统计，动态刷新）
    const pluginCount = ref(0)

    // 左侧信息面板数据
    const infoItems = computed(() => [
        { icon: 'cpu', label: 'CPU占用', value: '12%', color: '#4a9eff' },
        { icon: 'memory', label: '内存使用', value: '45%', color: '#36cfc9' },
        { icon: 'users', label: '在线玩家', value: '3/20', color: '#ffa940' },
        { icon: 'puzzle', label: '模组数量', value: String(modCount.value), color: '#ff6b6b' },
        { icon: 'blocks', label: '插件数量', value: String(pluginCount.value), color: '#b37feb' },
        { icon: 'clock', label: '运行时间', value: '2h 34m', color: '#73d13d' },
    ])

    onMounted(async () => {
        // 订阅后端推送的服务器日志与 mod 数量更新、版本检测询问
        EventsOn('server:log', handleServerLog)
        EventsOn('server:modcount', handleModCount)
        EventsOn('server:plugincount', handlePluginCount)
        EventsOn('server:askversion', handleAskVersion)
        // 加载活动服务器、Java 列表，再恢复该服务器已配置的 Java 和内存
        await loadActiveServer()
        await loadJavaList()
        await loadServerJava()
        await loadServerMemory()
        // 从后端 json 加载当前服务器的 mod 数量
        await loadModCount()
    })

    onBeforeUnmount(() => {
        // 组件卸载时取消订阅，避免泄漏
        EventsOff('server:log')
        EventsOff('server:modcount')
        EventsOff('server:plugincount')
        EventsOff('server:askversion')
    })

    // 终端输出内容（由后端推送的日志实时填充）
    const terminalLines = ref<LogLine[]>([])

    // 终端滚动容器引用（用于自动滚动到底部）
    const terminalScreenRef = ref<HTMLElement | null>(null)

    // 命令输入
    const commandInput = ref('')

    // 新增一行日志并自动滚动到底部
    async function appendLog(line: string) {
        terminalLines.value.push({
            text: line,
            level: detectLogLevel(line),
        })
        // 限制最大行数，避免内存无限增长
        if (terminalLines.value.length > 2000) {
            terminalLines.value.splice(0, terminalLines.value.length - 2000)
        }
        // 等 DOM 更新后滚动到底部
        await nextTick()
        if (terminalScreenRef.value) {
            terminalScreenRef.value.scrollTop = terminalScreenRef.value.scrollHeight
        }
    }

    // 接收后端推送的服务器日志
    function handleServerLog(line: string) {
        appendLog(line)
    }

    // ===== 服务器统计信息 =====

    // 加载当前服务器的 mod 与插件数量（从后端 ServerList.json 的 info 读取）
    async function loadModCount() {
        if (!currentServer.value) {
            modCount.value = 0
            pluginCount.value = 0
            return
        }
        // 分别获取 mod 与插件数量
        modCount.value = await serverApi.fetchServerModCount(currentServer.value)
        pluginCount.value = await serverApi.fetchServerPluginCount(currentServer.value)
    }

    // 接收后端在启动/停止/重启后推送的最新 mod 数量
    // 事件名：server:modcount
    function handleModCount(count: number) {
        modCount.value = count
    }

    // 接收后端在启动/停止/重启后推送的最新插件数量
    // 事件名：server:plugincount
    function handlePluginCount(count: number) {
        pluginCount.value = count
    }

    // 接收后端"是否输入 /version 提取版本"的询问
    // 事件名：server:askversion，payload 为服务器名
    // 弹确认框，用户确认后调用后端发送 /version
    function handleAskVersion(serverName: string) {
        dialog.warning({
            title: '检测服务器版本',
            content: '服务器已超过 10 秒无新日志输出，是否输入 /version 命令以获取版本号？',
            positiveText: '确认输入',
            negativeText: '跳过',
            onPositiveClick: async () => {
                await processApi.confirmSendVersion(serverName)
                appendLog('> 已发送 /version 命令')
            },
        })
    }

    // ===== Java 配置区域 =====

    // 系统检测到的 Java 列表
    const javaList = ref<Array<{ path: string; executable: string; version: number; versionName: string }>>([])

    // 当前选中的 Java（保存其安装目录路径）
    const selectedJava = ref('')

    // 加载系统 Java 列表（Model 层）
    async function loadJavaList() {
        javaList.value = await javaApi.scanJavaList()
    }

    // 加载当前服务器已配置的 Java
    // 后端返回完整 JavaInfo（含版本），即使不在扫描列表中也显示
    async function loadServerJava() {
        if (!currentServer.value) {
            selectedJava.value = ''
            return
        }
        const saved = await javaApi.getServerJava(currentServer.value)
        if (!saved) {
            selectedJava.value = ''
            return
        }

        // 在扫描列表中找到已保存的 Java（按 executable 匹配）
        const found = javaList.value.find((j) => j.executable === saved.executable)
        if (found) {
            selectedJava.value = found.path
            return
        }

        // 已保存的 Java 不在扫描列表中（如手动添加的自定义路径）
        // 将其补充进列表，确保下拉框能选中并显示
        javaList.value.push({
            path: saved.path,
            executable: saved.executable,
            version: saved.version,
            versionName: saved.versionName,
        })
        selectedJava.value = saved.path
    }

    // 下拉菜单选项：所有 Java 路径 + 添加选项
    const javaDropdownOptions = computed<DropdownOption[]>(() => {
        const options: DropdownOption[] = javaList.value.map((j) => ({
            label: `${j.path}  (${javaDisplayName(j)})`,
            key: j.path,
        }))
        options.push({
            label: '添加自定义 Java',
            key: '__add__',
        })
        return options
    })

    // 下拉选择处理
    async function handleJavaSelect(key: string | number) {
        // 需有活动服务器才能保存 Java 配置
        if (!currentServer.value) {
            message.warning('请先在首页选择当前服务器')
            return
        }

        // 添加自定义 Java
        if (key === '__add__') {
            const added = await javaApi.addJavaByDialog()
            // 用户取消选择时返回 null
            if (!added) return
            // 持久化选中的 Java 到当前服务器并刷新列表
            await javaApi.setServerJava(currentServer.value, added.executable)
            await loadJavaList()
            selectedJava.value = added.path
            message.success(`已添加 ${javaDisplayName(added)}`)
            return
        }

        // 选择已有 Java，找到对应对象并持久化到当前服务器
        selectedJava.value = String(key)
        const found = javaList.value.find((j) => j.path === selectedJava.value)
        if (found) {
            const ok = await javaApi.setServerJava(currentServer.value, found.executable)
            message[ok ? 'success' : 'error'](ok ? `已选择 ${javaDisplayName(found)}` : '保存 Java 选择失败')
        }
    }

    // 下拉选中项显示（展示 Java 版本 + 路径）
    const selectedJavaDisplay = computed(() => {
        const found = javaList.value.find((j) => j.path === selectedJava.value)
        if (found) {
            return javaDisplayName(found)
        }
        return '选择 Java'
    })

    // ===== 内存配置（GB 输入，后端存 MB） =====

    // 最大/最小内存（GB，前端显示单位）
    const xmxGB = ref<number>(4)
    const xmsGB = ref<number>(4)

    // 加载当前服务器的内存配置（后端存 MB，转 GB 显示）
    async function loadServerMemory() {
        if (!currentServer.value) return
        const [xmxMB, xmsMB] = await javaApi.getServerMemory(currentServer.value)
        if (xmxMB > 0) xmxGB.value = xmxMB / 1024
        if (xmsMB > 0) xmsGB.value = xmsMB / 1024
    }

    // 启动服务器（经过 CPU 检查流程）
    // 开服前自动保存前端的内存配置，再启动
    async function handleStartServer() {
        // 自动保存内存配置（GB → MB）
        if (currentServer.value) {
            const xmx = Math.round(xmxGB.value * 1024)
            const xms = Math.round(xmsGB.value * 1024)
            await javaApi.setServerMemory(currentServer.value, xmx, xms)
        }
        // 启动并接收具体错误信息（如未配置 Java 等）
        const result = await processApi.startServer(currentServer.value)
        if (!result.ok) {
            // 展示后端返回的具体错误原因
            appendLog(`[错误] 启动服务器失败：${result.message}`)
        }
    }

    // 停止服务器
    async function handleStopServer() {
        const ok = await processApi.stopServer()
        if (!ok) appendLog('[错误] 停止服务器失败')
    }

    // 发送命令
    async function sendCommand() {
        const cmd = commandInput.value.trim()
        if (!cmd) return
        commandInput.value = ''
        const ok = await processApi.sendCommand(cmd)
        if (!ok) appendLog('[错误] 命令发送失败')
    }

    // 右侧按钮（label/type 供模板使用；icon 为图标标识，由模板映射）
    const actionButtons = [
        { label: '开服', type: 'primary', icon: 'play', action: handleStartServer },
        { label: '关服', type: 'danger', icon: 'square', action: handleStopServer },
        { label: '重启', type: 'warning', icon: 'rotate', action: handleRestartServer },
        { label: '导出日志', type: 'default', icon: 'file', action: handleExportLog },
        { label: '导出错误日志', type: 'default', icon: 'alert', action: handleExportErrorLog },
    ]

    // 按钮点击分发
    function handleAction(action?: () => void) {
        action?.()
    }

    // 重启（先停止再启动）
    async function handleRestartServer() {
        await handleStopServer()
        await handleStartServer()
    }

    // 导出日志（暂为占位）
    function handleExportLog() {
        appendLog('[提示] 导出日志功能待实现')
    }

    // 导出错误日志（暂为占位）
    function handleExportErrorLog() {
        appendLog('[提示] 导出错误日志功能待实现')
    }

    return {
        // 状态
        infoItems,
        modCount,
        terminalLines,
        terminalScreenRef,
        commandInput,
        javaList,
        selectedJava,
        selectedJavaDisplay,
        javaDropdownOptions,
        xmxGB,
        xmsGB,
        currentServer,
        hasActiveServer,
        // 方法
        sendCommand,
        handleAction,
        handleJavaSelect,
        actionButtons,
    }
}
