package api

import (
	"MCServer/backend/launch/process"
	"MCServer/backend/system/gc"
	"context"
	"path/filepath"

	"MCServer/backend/server"
	"MCServer/backend/storage"
)

// ProcessApi 进程 API
// 供前端调用，管理 Java 服务端进程
type ProcessApi struct {
	ctx     context.Context
	service *process.ProcessManager
	scanner *server.Scanner
	info    *server.ServerInfo
	config  *storage.Storage
	// 当前操作的服务器名（用于停止/重启后定位需要刷新统计的目录）
	currentServer string
}

// NewProcessApi 创建进程 API 实例
func NewProcessApi(store *storage.Storage, scanner *server.Scanner, gcService *gc.GCService) *ProcessApi {
	service := process.NewProcessManager(store, gcService)
	return &ProcessApi{
		service: service,
		scanner: scanner,
		info:    server.NewServerInfo(store, service),
		config:  store,
	}
}

// Startup 保存上下文并传递给进程服务
func (api *ProcessApi) Startup(ctx context.Context) {
	api.ctx = ctx
	api.service.Startup(ctx)
}

// StartServer 启动指定服务器的 Java 进程
// 使用 config 中持久化的 Java 路径启动
// 前端调用：window.go.main.App.StartServer(serverName)
func (api *ProcessApi) StartServer(serverName string) error {
	if err := api.service.StartServer(serverName); err != nil {
		return err
	}
	api.currentServer = serverName
	// 启动后扫描 mods/plugins 目录并写入 ServerList.json，供前端按钮回调时拉取
	api.refreshModCount(serverName)
	return nil
}

// SendCommand 向运行中的服务器发送控制台命令
// 前端调用：window.go.main.App.SendCommand(command)
func (api *ProcessApi) SendCommand(command string) error {
	return api.service.SendCommand(command)
}

// GetServerUptime 获取当前服务器已运行秒数
// 服务器未运行时返回 0
// 前端调用：window.go.main.App.GetServerUptime()
func (api *ProcessApi) GetServerUptime() int {
	return api.info.GetUptime(api.currentServer)
}

// StopServer 停止正在运行的服务器进程
// 前端调用：window.go.main.App.StopServer()
func (api *ProcessApi) StopServer() error {
	name := api.currentServer
	if err := api.service.StopServer(); err != nil {
		return err
	}
	// 停止后扫描 mods/plugins 目录并写入 ServerList.json，供前端按钮回调时拉取
	api.refreshModCount(name)
	api.currentServer = ""
	return nil
}

// GetJvmPid 获取当前运行中 Java 进程的 PID
// 未运行时返回 0；供状态服务采集 JVM GC 统计（不直接暴露给前端）
func (api *ProcessApi) GetJvmPid() int {
	return api.service.GetRunningPid()
}

// ShutdownAll 停止所有正在运行的服务器进程
// 在应用关闭时调用（后端清理，不暴露给前端）
func (api *ProcessApi) ShutdownAll() {
	api.service.ShutdownAll()
}

// refreshModCount 重新扫描指定服务器的 mods/plugins 目录
// 将最新数量写入 ServerList.json 的 info.modCount / info.pluginCount
// 前端在开服/重启/关服按钮动作完成后，通过 GetServerModCount / GetServerPluginCount 主动拉取最新值
func (api *ProcessApi) refreshModCount(serverName string) {
	if serverName == "" {
		return
	}
	serversRoot, err := api.scanner.GetServerPath()
	if err != nil {
		return
	}
	serverPath := filepath.Join(serversRoot, serverName)
	// 同时扫描 mod 与插件数量
	modCount := server.ScanMods(serverPath)
	pluginCount := server.ScanPlugins(serverPath)
	_ = api.config.SetServerExtensionsCount(serverName, modCount, pluginCount)
}
