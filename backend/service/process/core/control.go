// 进程控制：命令发送、停止、状态查询
package core

import (
	"fmt"
	"io"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SendCommand 向运行中的服务器发送控制台命令
func (s *ProcessManager) SendCommand(command string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("服务器未在运行")
	}

	// 回显命令
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> %s", command))

	// 写入命令并换行
	if _, err := io.WriteString(s.stdin, command+"\n"); err != nil {
		return fmt.Errorf("发送命令失败: %v", err)
	}
	return nil
}

// StopServer 停止正在运行的服务器进程
func (s *ProcessManager) StopServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("服务器未在运行")
	}

	runtime.EventsEmit(s.ctx, "server:log", "> 正在停止服务器...")

	// 先尝试命令停止（发送 stop 命令给 Minecraft）
	if s.stdin != nil {
		_, _ = io.WriteString(s.stdin, "stop\n")
	}

	// 终止进程
	err := s.cmd.Process.Kill()
	s.cmd = nil
	s.stdin = nil
	// 清空检测状态机
	s.detector = nil
	// 清空启动时间（运行时长归零）
	s.startTime = time.Time{}

	if err != nil {
		return fmt.Errorf("停止服务器失败: %v", err)
	}
	runtime.EventsEmit(s.ctx, "server:log", "> 服务器已停止")
	return nil
}

// GetCurrentServer 返回当前正在运行的服务器名
// 若当前无运行进程，返回空字符串
func (s *ProcessManager) GetCurrentServer() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd == nil || s.cmd.Process == nil {
		return ""
	}
	return s.currentServer
}

// GetServerUptime 返回服务器已运行秒数
// 服务器未运行时返回 0
func (s *ProcessManager) GetServerUptime() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 进程未运行或未记录启动时间，返回 0
	if s.cmd == nil || s.cmd.Process == nil || s.startTime.IsZero() {
		return 0
	}
	// 计算距启动时间的秒数
	return int(time.Since(s.startTime).Seconds())
}

// ShutdownAll 停止所有正在运行的服务器进程
// 在应用退出时调用，确保关闭 APP 时所有服务端都被终止
func (s *ProcessManager) ShutdownAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return
	}

	// 尝试优雅停止（发送 stop 命令）
	if s.stdin != nil {
		_, _ = io.WriteString(s.stdin, "stop\n")
	}

	// 终止进程
	_ = s.cmd.Process.Kill()
	s.cmd = nil
	s.stdin = nil
}
