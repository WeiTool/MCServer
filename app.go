package main

import (
	"context"
	"os"

	"MCServer/backend/api"
	"MCServer/backend/server"
	"MCServer/backend/storage"
	"MCServer/backend/system/gc"
	"MCServer/backend/system/io"
	"MCServer/backend/system/sysinfo"
	"MCServer/backend/utils"
)

// App 应用结构体
// 承载通过 Wails 暴露给前端调用的后端 API
type App struct {
	// 应用上下文（用于 runtime 事件推送等）
	ctx context.Context
	// 嵌入各领域 API，继承其全部方法，供前端调用
	*api.StatusApi
	*api.ServerApi
	*api.ConfigApi
	*api.ProcessApi
	*api.JavaApi
	*api.ExportApi
}

// NewApp 创建一个新的 App 应用结构体
func NewApp() *App {
	// 共享实例统一注入各 API，避免重复创建
	store := storage.NewStorage()
	scanner := server.NewScanner()
	javaService := sysinfo.NewJavaService()
	// GC 统计服务：进程管理器（记录 Java bin）与状态服务（采集推送）共用同一实例
	gcService := gc.NewGCService()
	// 磁盘读写速率统计服务（按 JVM 进程采集）
	ioService := io.NewIOService()
	processApi := api.NewProcessApi(store, scanner, gcService)

	return &App{
		StatusApi:  api.NewStatusApi(processApi, gcService, ioService),
		ServerApi:  api.NewServerApi(store, scanner),
		ConfigApi:  api.NewConfigApi(store, javaService),
		ProcessApi: processApi,
		JavaApi:    api.NewJavaApi(javaService),
		ExportApi:  api.NewExportApi(),
	}
}

// Startup 应用启动回调
// 保存上下文，并传递给状态服务
// 同时检测并创建 servers 目录
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.StatusApi.Startup(ctx)
	a.ProcessApi.Startup(ctx)
	a.JavaApi.Startup(ctx)
	a.ExportApi.Startup(ctx)

	// 检测并创建 servers 目录
	a.ensureServersDir()
}

// ensureServersDir 确保 servers 目录存在
// 在 exe 同级目录下创建
func (a *App) ensureServersDir() {
	// 获取 servers 根目录路径（统一由 utils 提供）
	serversPath := utils.GetServersRoot()
	if serversPath == "" {
		// 获取失败，静默处理
		return
	}
	// 创建目录（权限 0755：rwxr-xr-x）
	_ = os.MkdirAll(serversPath, 0755)
}
