// Package utils 提供跨包复用的工具函数
package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetExeDir 获取当前可执行文件所在目录的绝对路径
// 若获取失败，返回空字符串（调用方需自行处理）
func GetExeDir() string {
	// 获取可执行文件绝对路径
	execPath, err := os.Executable()
	if err != nil {
		// 获取失败时返回空字符串，避免返回无效路径
		return ""
	}
	// 返回可执行文件所在目录
	return filepath.Dir(execPath)
}

// GetConfigDir 获取 config 配置文件夹的绝对路径（exe 同级）
// 所有 JSON 配置文件（ServerList.json、JavaList.json 等）统一存放在该目录
func GetConfigDir() string {
	// 先取 exe 所在目录
	exeDir := GetExeDir()
	if exeDir == "" {
		// exe 目录获取失败，退回相对路径 config
		return "config"
	}
	// exe 目录下拼接 config
	return filepath.Join(exeDir, "config")
}

// GetServersRoot 获取 servers 服务器根目录的绝对路径（exe 同级）
func GetServersRoot() string {
	// 先取 exe 所在目录
	exeDir := GetExeDir()
	if exeDir == "" {
		// exe 目录获取失败，退回相对路径 servers
		return "servers"
	}
	// exe 目录下拼接 servers
	return filepath.Join(exeDir, "servers")
}

// GetServerFolderPath 获取指定服务器的文件夹绝对路径（servers 根目录 + 服务器名）
func GetServerFolderPath(serverName string) string {
	// 服务器根目录拼接服务器名
	return filepath.Join(GetServersRoot(), serverName)
}

// IsJarFile 判断文件名是否为 .jar 文件（不区分大小写）
// 用于统一识别 Minecraft 服务端 jar、mod、插件文件
func IsJarFile(fileName string) bool {
	// 转小写后判断后缀是否为 .jar
	return strings.HasSuffix(strings.ToLower(fileName), ".jar")
}
