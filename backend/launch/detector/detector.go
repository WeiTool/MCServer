// 本文件实现服务端类型检测的状态机与关键字匹配
package detector

import (
	"strings"
	"sync"
)

// DetectionState 类型检测状态机
// 记录类型检测的进度
type DetectionState struct {
	// 互斥锁保护并发读写
	mu sync.Mutex
	// 目标服务器名
	serverName string
	// 是否需要检测类型
	needType bool
	// 是否需要检查版本
	needVersion bool
	// 类型是否已检测完成
	typeDone bool
	// 版本是否已检测完成
	versionDone bool
}

// NewDetectionState 创建检测状态
// 根据 json 中 type 是否为空，决定是否需要检测类型
func NewDetectionState(serverName string, needType bool, needVersion bool) *DetectionState {
	return &DetectionState{
		serverName:  serverName,
		needType:    needType,
		typeDone:    !needType, // 无需检测则视为完成
		needVersion: needVersion,
		versionDone: !needVersion,
	}
}

// ServerName 返回目标服务器名
func (d *DetectionState) ServerName() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.serverName
}

// NeedType 是否需要类型检测（且未完成）
func (d *DetectionState) NeedType() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.needType && !d.typeDone
}

// NeedVersion 是否需要版本检测（且未完成）
func (d *DetectionState) NeedVersion() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.needVersion && !d.versionDone
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
