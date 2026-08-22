// Package detector 提供服务端类型的日志检测能力
// 本文件实现类型关键字配置（server-types.json）的加载
package detector

import (
	"encoding/json"
	"os"
	"path/filepath"

	"MCServer/backend/utils"
)

// typeKeyword 类型关键字匹配项（对应 server-types.json 中的每一项）
// 命中 keyword 即判定为对应 serverType
type typeKeyword struct {
	// 日志中要匹配的关键字
	Keyword string `json:"keyword"`
	// 命中的服务器类型名（如 Vanilla/Fabric/Forge）
	Type string `json:"type"`
}

// serverTypesFileName 类型关键字映射配置文件
// 存放在 exe 同级 config 目录，方便手动增删改类型
const serverTypesFileName = "server-types.json"

// serverTypeKeywords 类型关键字匹配表
// 首次从 server-types.json 加载，文件不存在时自动生成默认内容
var serverTypeKeywords = loadServerTypes()

// defaultServerTypeKeywords 默认的类型关键字映射（用于首次生成 JSON 文件）
// 修改类型请在 config/server-types.json 中调整，无需改代码
var defaultServerTypeKeywords = []typeKeyword{
	{Keyword: "Starting net.minecraft.server.Main", Type: "Vanilla"},
	{Keyword: "Starting net.fabricmc.loader.impl.game.minecraft.BundlerClassPathCapture", Type: "Fabric"},
	{Keyword: "MinecraftForge", Type: "Forge"},
	{Keyword: "NeoForge mod loading", Type: "NeoForge"},
	{Keyword: "Loading Paper", Type: "Paper"},
	{Keyword: "Loading Purpur", Type: "Purpur"},
	{Keyword: "Mohist", Type: "Mohist"},
	{Keyword: "Youer", Type: "Youer"},
}

// loadServerTypes 从 config/server-types.json 加载类型关键字映射表
// 若文件不存在，则用默认内容生成该文件，方便后续手动修改
func loadServerTypes() []typeKeyword {
	// 拼出 server-types.json 的完整路径
	filePath := filepath.Join(utils.GetConfigDir(), serverTypesFileName)

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 文件不存在或读取失败：生成默认 JSON 文件后返回默认映射
		if os.IsNotExist(err) {
			writeDefaultServerTypes(filePath)
		}
		return defaultServerTypeKeywords
	}

	// 解析 JSON 数组
	var keywords []typeKeyword
	if err := json.Unmarshal(data, &keywords); err != nil {
		// 解析失败时退回默认映射，避免影响检测
		return defaultServerTypeKeywords
	}

	// 解析成功，返回配置文件中的映射
	return keywords
}

// writeDefaultServerTypes 将默认类型关键字映射写入 server-types.json
// 仅在文件不存在时调用，供用户对照修改
func writeDefaultServerTypes(filePath string) {
	// 序列化为带缩进的 JSON
	data, err := json.MarshalIndent(defaultServerTypeKeywords, "", "    ")
	if err != nil {
		// 序列化失败则放弃写入
		return
	}
	// 写入文件（权限 0644：rw-r--r--）
	_ = os.WriteFile(filePath, data, 0644)
}
