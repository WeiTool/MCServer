// Package api 提供前端可调用的 API 层
// 负责接收前端请求，并调用 service 层处理业务逻辑
package api

import (
	"context"
	"time"

	"MCServer/backend/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StatusApi 系统状态 API
// 封装系统状态相关的接口，供前端调用
type StatusApi struct {
	// 状态服务实例
	StatusService *service.StatusService
}

// NewStatusApi 创建系统状态 API 实例
func NewStatusApi() *StatusApi {
	return &StatusApi{
		StatusService: service.NewStatusService(),
	}
}

// Startup 启动初始化
// 保存上下文，并启动内存信息实时推送
func (api *StatusApi) Startup(ctx context.Context) {
	api.StatusService.Startup(ctx)
	// 启动内存监控 goroutine，主动推送内存信息给前端
	go api.watchMemory(ctx)
}

// watchMemory 定时采集内存信息并主动推送给前端
// 事件名：memory:update，数据：*model.MemoryInfo
// 相比前端轮询，主动推送更省资源
func (api *StatusApi) watchMemory(ctx context.Context) {
	// 推送间隔 2 秒，避免频繁采集占用系统资源
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 启动后立即推送一次，让前端快速拿到首屏数据
	api.pushMemory(ctx)

	for {
		select {
		// 应用退出时停止监控，避免 goroutine 泄漏
		case <-ctx.Done():
			return
		// 定时推送内存信息
		case <-ticker.C:
			api.pushMemory(ctx)
		}
	}
}

// pushMemory 采集内存信息并推送事件给前端
func (api *StatusApi) pushMemory(ctx context.Context) {
	mem, err := api.StatusService.GetMemoryInfo()
	if err != nil {
		return
	}
	runtime.EventsEmit(ctx, "memory:update", mem)
}
