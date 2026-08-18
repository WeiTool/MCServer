// 本文件实现服务端类型与版本的日志检测逻辑
package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

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

// versionNameRegex 匹配 /version 输出中的 "name = xxx" 行
// 支持两种日志格式：
//   [xx] [Server thread/INFO] [minecraft/MinecraftServer]: name = 26.2
//   [xx] [Server thread/INFO]: name = 26.2
// 提取 name 后的版本值（如 26.2）
var versionNameRegex = regexp.MustCompile(`name\s*=\s*([^\s]+)`)

// DetectionState 检测状态机
// 记录当前需要检测的项、是否已完成、以及版本检测的静默计时
type DetectionState struct {
	// 互斥锁保护并发读写
	mu sync.Mutex
	// 目标服务器名
	serverName string
	// 是否需要检测类型
	needType bool
	// 是否需要检测版本
	needVersion bool
	// 类型是否已检测完成
	typeDone bool
	// 版本是否已检测完成
	versionDone bool
	// 版本确认是否已发出（防止重复发事件询问）
	versionAsked bool
	// 上次收到日志的时间（用于 10 秒静默判断）
	lastLogTime time.Time
	// 是否已发送 /version 命令（等待输出中）
	versionSent bool
}

// NewDetectionState 创建检测状态
// 根据 json 中 type/version 是否为空，决定需要检测哪些项
func NewDetectionState(serverName string, needType, needVersion bool) *DetectionState {
	return &DetectionState{
		serverName:   serverName,
		needType:     needType,
		needVersion:  needVersion,
		lastLogTime:  time.Now(),
		typeDone:     !needType,    // 无需检测则视为完成
		versionDone:  !needVersion, // 无需检测则视为完成
	}
}

// AllDone 是否所有检测项都已完成
func (d *DetectionState) AllDone() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.typeDone && d.versionDone
}

// NeedVersion 是否需要版本检测（且未完成）
func (d *DetectionState) NeedVersion() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.needVersion && !d.versionDone
}

// NeedType 是否需要类型检测（且未完成）
func (d *DetectionState) NeedType() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.needType && !d.typeDone
}

// MarkTypeDone 标记类型检测完成
func (d *DetectionState) MarkTypeDone() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.typeDone = true
}

// MarkVersionDone 标记版本检测完成
func (d *DetectionState) MarkVersionDone() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.versionDone = true
}

// UpdateLogTime 更新最近一次收到日志的时间
func (d *DetectionState) UpdateLogTime() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastLogTime = time.Now()
}

// SilentSeconds 返回距上次日志的静默秒数
func (d *DetectionState) SilentSeconds() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return time.Since(d.lastLogTime).Seconds()
}

// ShouldAskVersion 判断是否应该弹出"是否输入 /version"询问
// 条件：需要版本检测、未完成、尚未询问过、静默超过 10 秒
func (d *DetectionState) ShouldAskVersion() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.needVersion || d.versionDone || d.versionAsked || d.versionSent {
		return false
	}
	return time.Since(d.lastLogTime).Seconds() > 10
}

// MarkVersionAsked 标记已发送询问事件
func (d *DetectionState) MarkVersionAsked() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.versionAsked = true
}

// MarkVersionSent 标记已发送 /version 命令
func (d *DetectionState) MarkVersionSent() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.versionSent = true
}

// IsVersionSent 是否已发送过 /version 命令
func (d *DetectionState) IsVersionSent() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.versionSent
}

// DetectType 匹配日志行中的类型关键字
// 命中返回 (类型名, true)，未命中返回 ("", false)
func DetectType(line string) (string, bool) {
	// 遍历类型关键字映射（来自 server-types.json）
	for _, kw := range serverTypeKeywords {
		// 命中关键字即返回对应类型
		if strings.Contains(line, kw.Keyword) {
			return kw.Type, true
		}
	}
	return "", false
}

// ExtractVersion 从日志行中提取 /version 输出的 name 字段值
// 匹配 "name = xxx" 格式，返回 xxx；未匹配返回空字符串
func ExtractVersion(line string) string {
	m := versionNameRegex.FindStringSubmatch(line)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
