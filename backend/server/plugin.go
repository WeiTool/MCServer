// 本文件实现服务器插件数量的扫描
package server

import (
	"path/filepath"
)

// ScanPlugins 统计服务器 plugins 目录下的插件数量
// 只统计 .jar 文件；目录不存在或读取失败时返回 0
func ScanPlugins(serverPath string) int {
	return countJarInDir(filepath.Join(serverPath, "plugins"))
}
