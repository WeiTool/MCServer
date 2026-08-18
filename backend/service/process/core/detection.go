// 类型/版本检测：分析服务器日志行，识别服务端类型与版本
package core

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// handleDetectionLine 分析一行服务器日志，执行类型/版本检测
// 由 streamLines 在推送日志后调用
func (s *ProcessManager) handleDetectionLine(line string) {
	// 检测状态机为空时直接返回（无检测需求）
	if s.detector == nil {
		return
	}

	// 更新最近收到日志的时间（用于 10 秒静默判断）
	s.detector.UpdateLogTime()

	// 1. 类型检测：若仍需类型检测，匹配类型关键字
	if s.detector.NeedType() {
		if serverType, ok := DetectType(line); ok {
			// 命中类型关键字，写入 json 并标记完成
			if err := s.cfg.SetServerType(s.detector.serverName, serverType); err == nil {
				s.detector.MarkTypeDone()
				runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 已检测到服务器类型: %s", serverType))
				// 推送类型更新事件给前端（payload 为类型名）
				runtime.EventsEmit(s.ctx, "server:type", serverType)
			}
		}
	}

	// 2. 版本检测：若已发送 /version 且仍需版本检测，尝试从当前行提取 name 值
	if s.detector.IsVersionSent() && s.detector.NeedVersion() {
		if ver := ExtractVersion(line); ver != "" {
			if err := s.cfg.SetServerVersion(s.detector.serverName, ver); err == nil {
				s.detector.MarkVersionDone()
				runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 已检测到服务器版本: %s", ver))
				// 推送版本更新事件给前端（payload 为版本号）
				runtime.EventsEmit(s.ctx, "server:version", ver)
			}
		}
	}

	// 3. 版本静默检测：若需要版本检测且静默超 10 秒，发事件询问前端
	if s.detector.ShouldAskVersion() {
		s.detector.MarkVersionAsked()
		runtime.EventsEmit(s.ctx, "server:askversion", s.detector.serverName)
	}
}

// ConfirmSendVersion 前端确认后，发送 /version 命令以提取版本
// 由 ProcessApi.ConfirmSendVersion 调用
func (s *ProcessManager) ConfirmSendVersion(serverName string) error {
	// 检测状态机为空或不是当前服务器，忽略
	if s.detector == nil || s.detector.serverName != serverName {
		return nil
	}

	// 标记已发送 /version（防止重复提取时误判）
	s.detector.MarkVersionSent()

	// 发送 /version 命令
	return s.SendCommand("/version")
}
