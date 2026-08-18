// eula 自动处理：等待 eula.txt 出现，走"关服→改true→自动重启"流程
package core

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// watchEulaAndRestart 等待 eula.txt 出现，处理"关服→改true→自动重启"流程
// 仅在启动时 eula.txt 缺失的情况下由 StartServer 启动
func (s *ProcessManager) watchEulaAndRestart(serverName string) {
	runtime.EventsEmit(s.ctx, "server:log", "> 未检测到 eula.txt，等待服务端生成...")

	// 持续轮询 eula.txt，最多等待 5 分钟
	appeared := WaitEulaAppear(serverName, 2*time.Second, 5*time.Minute)
	if !appeared {
		runtime.EventsEmit(s.ctx, "server:log", "[警告] 等待 eula.txt 超时，请手动处理")
		return
	}

	runtime.EventsEmit(s.ctx, "server:log", "> 检测到 eula.txt 出现，2 秒后自动重启以同意 EULA...")

	// 等待 2 秒
	time.Sleep(2 * time.Second)

	// 关服（忽略错误，因为服务端可能已因 eula 未同意而自行退出）
	s.mu.Lock()
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		s.cmd = nil
		s.stdin = nil
	}
	s.mu.Unlock()

	// 修改 eula.txt 为 eula=true
	if err := SetEulaAgreed(serverName); err != nil {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[错误] 修改 eula.txt 失败: %v", err))
		return
	}
	runtime.EventsEmit(s.ctx, "server:log", "> 已同意 EULA，自动重新启动服务器...")

	// 自动重启（继续类型/版本检测）
	if err := s.StartServer(serverName); err != nil {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[错误] 自动重启失败: %v", err))
	}
}
