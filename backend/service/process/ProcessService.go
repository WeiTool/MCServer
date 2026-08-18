// Package process 提供 Java 服务端进程的管理能力
// 本文件为对外门面：保持 ProcessService 对外接口不变，实际实现在 core 子包
package process

import (
	"context"

	"MCServer/backend/service/process/core"
)

// ProcessService 进程服务（对外门面）
// 所有方法委托给 core.ProcessManager，对外接口保持不变
type ProcessService struct {
	*core.ProcessManager
}

// NewProcessService 创建进程服务实例
func NewProcessService() *ProcessService {
	return &ProcessService{ProcessManager: core.NewProcessManager()}
}

// Startup 保存应用上下文
func (s *ProcessService) Startup(ctx context.Context) {
	s.ProcessManager.Startup(ctx)
}
