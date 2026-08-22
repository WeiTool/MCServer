// 本文件实现服务器 mod 数量的扫描
package server

import (
	"os"
	"path/filepath"

	"MCServer/backend/utils"
)

// ScanMods 统计服务器 mods 目录下的模组数量
// 只统计 .jar 文件；目录不存在或读取失败时返回 0
func ScanMods(serverPath string) int {
	return countJarInDir(filepath.Join(serverPath, "mods"))
}

// countJarInDir 统计指定目录下 .jar 文件数量
// 供 mod/plugin 扫描共用；目录不存在或读取失败返回 0
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
