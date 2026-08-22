// Model 层：服务器相关 API 封装
// 统一管理所有 wails 后端调用，供 ViewModel/组件使用
import {
  GetServerList,
  SetActiveServer,
  GetActiveServer,
  GetServerModCount,
  GetServerPluginCount,
  GetServerType,
  GetServerVersion,
} from '../../wailsjs/go/main/App'
import type { model } from '../../wailsjs/go/models'

// 获取服务器列表
export async function fetchServerList(): Promise<model.ServerListResult | null> {
  try {
    return await GetServerList()
  } catch {
    return null
  }
}

// 设置当前活动服务器
export async function setActiveServer(name: string): Promise<boolean> {
  try {
    await SetActiveServer(name)
    return true
  } catch {
    return false
  }
}

// 获取当前活动服务器名称
export async function fetchActiveServer(): Promise<string> {
  try {
    return await GetActiveServer() || ''
  } catch {
    return ''
  }
}

// 获取指定服务器的 mod 数量（serverName 为服务器文件夹名称）
export async function fetchServerModCount(serverName: string): Promise<number> {
  try {
    return (await GetServerModCount(serverName)) || 0
  } catch {
    return 0
  }
}

// 获取指定服务器的插件数量（serverName 为服务器文件夹名称）
export async function fetchServerPluginCount(serverName: string): Promise<number> {
  try {
    return (await GetServerPluginCount(serverName)) || 0
  } catch {
    return 0
  }
}

// 获取指定服务器的类型（serverName 为服务器文件夹名称）
// 未检测到则返回空字符串
export async function fetchServerType(serverName: string): Promise<string> {
  try {
    return (await GetServerType(serverName)) || ''
  } catch {
    return ''
  }
}

// 获取指定服务器的版本（serverName 为服务器文件夹名称）
// 未检测到则返回空字符串
export async function fetchServerVersion(serverName: string): Promise<string> {
  try {
    return (await GetServerVersion(serverName)) || ''
  } catch {
    return ''
  }
}
