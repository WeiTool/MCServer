// Package api 提供前端可调用的 API 层
// 负责接收前端请求，并调用底层服务处理业务逻辑
package api

import (
	"context"
	"time"

	"MCServer/backend/model"
	"MCServer/backend/system/gc"
	"MCServer/backend/system/io"
	"MCServer/backend/system/sysinfo"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// pidProvider 提供当前运行中的 JVM 进程 PID（由进程服务实现）
type pidProvider interface {
	GetJvmPid() int
}

// StatusApi 系统状态 API
// 封装系统状态相关的接口，供前端调用
type StatusApi struct {
	// 当前运行中 JVM 的 PID 提供者（用于采集 GC/JVM 内存/磁盘 IO）
	pidProvider pidProvider
	// GC 统计服务（jstat 采集）
	gcService *gc.GCService
	// 磁盘读写速率统计服务
	ioService *io.IOService
}

// NewStatusApi 创建系统状态 API 实例
func NewStatusApi(provider pidProvider, gcService *gc.GCService, ioService *io.IOService) *StatusApi {
	return &StatusApi{
		pidProvider: provider,
		gcService:   gcService,
		ioService:   ioService,
	}
}

// Startup 启动初始化
// 启动内存/CPU 信息实时推送
func (api *StatusApi) Startup(ctx context.Context) {
	// 启动系统状态监控 goroutine，主动推送内存/CPU 信息给前端
	go api.watchSystem(ctx)
}

// watchSystem 定时采集系统状态并主动推送给前端
// 事件名：memory:update / cpu:update
// 相比前端轮询，主动推送更省资源
func (api *StatusApi) watchSystem(ctx context.Context) {
	// 推送间隔 2 秒，避免频繁采集占用系统资源
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 启动后立即推送一次，让前端快速拿到首屏数据
	api.pushSystem(ctx)

	for {
		select {
		// 应用退出时停止监控，避免 goroutine 泄漏
		case <-ctx.Done():
			return
		// 定时推送系统状态
		case <-ticker.C:
			api.pushSystem(ctx)
		}
	}
}

// pushSystem 采集内存/CPU/GC/JVM/磁盘 IO 信息并推送事件给前端
func (api *StatusApi) pushSystem(ctx context.Context) {
	// 推送内存
	if mem, err := sysinfo.GetMemoryInfo(); err == nil {
		runtime.EventsEmit(ctx, "memory:update", mem)
	}
	// 推送 CPU
	if cpu, err := sysinfo.GetCPUInfo(); err == nil {
		runtime.EventsEmit(ctx, "cpu:update", cpu)
	}

	// 服务器运行时，推送 JVM 相关统计（GC/内存/磁盘读写），失败静默跳过
	if api.pidProvider == nil {
		return
	}
	pid := api.pidProvider.GetJvmPid()
	if pid <= 0 {
		return
	}

	// 推送 JVM GC 统计
	if api.gcService != nil {
		if stats, err := api.gcService.GetStats(pid); err == nil {
			runtime.EventsEmit(ctx, "gc:update", stats)
		}
	}
	// 推送 JVM 进程内存使用率（占系统总内存的百分比）
	if usage, err := sysinfo.GetJVMProcessMemoryUsage(int32(pid)); err == nil {
		runtime.EventsEmit(ctx, "jvm:update", &model.MemoryInfo{UsagePercent: usage})
	}
	// 推送 JVM 进程磁盘读写速率
	if api.ioService != nil {
		if stats, err := api.ioService.GetStats(pid); err == nil {
			runtime.EventsEmit(ctx, "io:update", stats)
		}
	}
}
