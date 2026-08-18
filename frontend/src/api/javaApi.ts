// Model 层：Java 相关 API 封装
import {
  ScanJavaList,
  AddJavaByDialog,
  SetServerJava,
  GetServerJava,
  SetServerMemory,
  GetServerMemory,
} from '../../wailsjs/go/main/App'
import type { model } from '../../wailsjs/go/models'

// Java 信息类型（复用后端 model.JavaInfo）
export type JavaInfo = model.JavaInfo

// 扫描系统所有 Java 环境
export async function scanJavaList(): Promise<JavaInfo[]> {
  try {
    return await ScanJavaList() || []
  } catch (e) {
    console.error('扫描 Java 失败:', e)
    return []
  }
}

// 打开系统文件选择框添加 Java
export async function addJavaByDialog(): Promise<JavaInfo | null> {
  try {
    const info = await AddJavaByDialog()
    if (!info || !info.path) return null
    return info
  } catch (e) {
    console.error('添加 Java 失败:', e)
    return null
  }
}

// 为指定服务器设置 java.exe 路径
export async function setServerJava(serverName: string, executable: string): Promise<boolean> {
  try {
    await SetServerJava(serverName, executable)
    return true
  } catch (e) {
    console.error('保存 Java 失败:', e)
    return false
  }
}

// 获取指定服务器配置的 Java 信息（含版本）
// 返回 null 表示未配置
export async function getServerJava(serverName: string): Promise<JavaInfo | null> {
  try {
    const info = await GetServerJava(serverName)
    if (!info || !info.executable) return null
    return info
  } catch (e) {
    return null
  }
}

// 为指定服务器设置内存（MB）
export async function setServerMemory(serverName: string, xmxMB: number, xmsMB: number): Promise<boolean> {
  try {
    await SetServerMemory(serverName, xmxMB, xmsMB)
    return true
  } catch (e) {
    console.error('保存内存失败:', e)
    return false
  }
}

// 获取指定服务器内存（MB），返回 [xmxMB, xmsMB]
export async function getServerMemory(serverName: string): Promise<[number, number]> {
  try {
    // Wails 多返回值实际为数组
    const res = await GetServerMemory(serverName) as unknown as number[]
    return [res[0] || 0, res[1] || 0]
  } catch (e) {
    return [0, 0]
  }
}
