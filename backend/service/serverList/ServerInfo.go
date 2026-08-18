// Package serverList 提供服务器列表的扫描与信息统计能力
package serverList

import (
	"os"
	"path/filepath"
	"strconv"

	"MCServer/backend/model"
	"MCServer/backend/utils"
)

// ServerInfo 服务器信息统计服务
// 负责查询单个服务器的 mod 信息、插件信息、类型与版本
type ServerInfo struct{}

// NewServerInfo 创建服务器信息统计服务实例
func NewServerInfo() *ServerInfo { return &ServerInfo{} }

// ScanMods 统计服务器 mods 目录下的模组数量
// 只统计 .jar 文件；目录不存在或读取失败时返回 0
func (s *ServerInfo) ScanMods(serverPath string) int {
	return countJarInDir(filepath.Join(serverPath, "mods"))
}

// ScanPlugins 统计服务器 plugins 目录下的插件数量
// 只统计 .jar 文件；目录不存在或读取失败时返回 0
func (s *ServerInfo) ScanPlugins(serverPath string) int {
	return countJarInDir(filepath.Join(serverPath, "plugins"))
}

// ScanServerInfo 综合查询服务器统计信息
// 返回 mod 数量与插件数量
func (s *ServerInfo) ScanServerInfo(serverPath string) model.ServerInfo {
	return model.ServerInfo{
		ModCount:    strconv.Itoa(s.ScanMods(serverPath)),
		PluginCount: strconv.Itoa(s.ScanPlugins(serverPath)),
	}
}

// countJarInDir 统计指定目录下 .jar 文件数量
// 目录不存在或读取失败返回 0
func countJarInDir(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// 统一用 utils.IsJarFile 判断是否为 jar 文件
		if utils.IsJarFile(entry.Name()) {
			count++
		}
	}
	return count
}
