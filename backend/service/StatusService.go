package service

import (
	"context"

	"MCServer/backend/model"

	"github.com/shirou/gopsutil/v3/mem"
)

// StatusService 系统状态服务
// 负责采集系统内存信息
type StatusService struct {
	// 应用上下文
	Ctx context.Context
}

// NewStatusService 创建系统状态服务实例
func NewStatusService() *StatusService {
	return &StatusService{}
}

// Startup 启动初始化
// 保存应用上下文
func (s *StatusService) Startup(ctx context.Context) {
	s.Ctx = ctx
}

// GetMemoryInfo 获取内存信息
// 返回包含内存使用率百分比的内存信息指针
func (s *StatusService) GetMemoryInfo() (*model.MemoryInfo, error) {
	// 获取虚拟内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// 直接使用 gopsutil 计算好的使用率，前端无需再计算
	return &model.MemoryInfo{
		UsagePercent: memInfo.UsedPercent,
	}, nil
}