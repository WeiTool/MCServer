// Model 层：进程相关 API 封装
import { StartServer, SendCommand, StopServer, ConfirmSendVersion, GetServerUptime } from '../../wailsjs/go/main/App'

// 启动结果：是否成功 + 错误信息（失败时 message 含具体原因）
export interface StartResult {
  ok: boolean
  message: string
}

// 启动服务器（使用该服务器配置的 Java 和内存）
// 失败时把后端的错误信息透传给调用方，便于前端展示具体原因
export async function startServer(serverName: string): Promise<StartResult> {
  try {
    await StartServer(serverName)
    return { ok: true, message: '' }
  } catch (e) {
    // 提取 Wails 抛出的错误信息（可能为 Error 或字符串）
    const message = e instanceof Error ? e.message : String(e)
    console.error('启动服务器失败:', message)
    return { ok: false, message }
  }
}

// 向运行中的服务器发送命令
export async function sendCommand(command: string): Promise<boolean> {
  try {
    await SendCommand(command)
    return true
  } catch (e) {
    console.error('发送命令失败:', e)
    return false
  }
}

// 停止服务器
export async function stopServer(): Promise<boolean> {
  try {
    await StopServer()
    return true
  } catch (e) {
    console.error('停止服务器失败:', e)
    return false
  }
}

// 用户确认后，通知后端发送 /version 命令以提取版本号
export async function confirmSendVersion(serverName: string): Promise<boolean> {
  try {
    await ConfirmSendVersion(serverName)
    return true
  } catch (e) {
    console.error('发送 /version 失败:', e)
    return false
  }
}

// 获取服务器已运行秒数（未运行时返回 0）
export async function fetchServerUptime(): Promise<number> {
  try {
    return (await GetServerUptime()) || 0
  } catch (e) {
    console.error('获取运行时长失败:', e)
    return 0
  }
}
