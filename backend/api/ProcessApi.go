package api

import (
	"context"
	"path/filepath"

	"MCServer/backend/service/config"
	"MCServer/backend/service/process"
	"MCServer/backend/service/serverList"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProcessApi 进程 API
// 供前端调用，管理 Java 服务端进程
type ProcessApi struct {
	ctx     context.Context
	service *process.ProcessService
	scanner *serverList.ServerScanner
	info    *serverList.ServerInfo
	config  *config.ConfigService
	// 当前操作的服务器名（用于停止/重启后定位需要刷新统计的目录）
	currentServer string
}

// NewProcessApi 创建进程 API 实例
func NewProcessApi() *ProcessApi {
	return &ProcessApi{
		service: process.NewProcessService(),
		scanner: serverList.NewServerScanner(),
		info:    serverList.NewServerInfo(),
		config:  config.NewConfigService(),
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
	// 启动后刷新该服务器的 mod 数量并推送前端
	api.refreshModCount(serverName)
	return nil
}

// SendCommand 向运行中的服务器发送控制台命令
// 前端调用：window.go.main.App.SendCommand(command)
func (api *ProcessApi) SendCommand(command string) error {
	return api.service.SendCommand(command)
}

// ConfirmSendVersion 前端确认后，向后端发送 /version 命令以提取版本号
// 由前端在收到 server:askversion 事件并弹窗确认后调用
// 前端调用：window.go.main.App.ConfirmSendVersion(serverName)
func (api *ProcessApi) ConfirmSendVersion(serverName string) error {
	return api.service.ConfirmSendVersion(serverName)
}

// GetServerUptime 获取服务器已运行秒数
// 服务器未运行时返回 0
// 前端调用：window.go.main.App.GetServerUptime()
func (api *ProcessApi) GetServerUptime() int {
	return api.service.GetServerUptime()
}

// StopServer 停止正在运行的服务器进程
// 前端调用：window.go.main.App.StopServer()
func (api *ProcessApi) StopServer() error {
	name := api.currentServer
	if err := api.service.StopServer(); err != nil {
		return err
	}
	// 停止后刷新该服务器的 mod 数量并推送前端
	api.refreshModCount(name)
	api.currentServer = ""
	return nil
}

// ShutdownAll 停止所有正在运行的服务器进程
// 在应用关闭时调用（后端清理，不暴露给前端）
func (api *ProcessApi) ShutdownAll() {
	api.service.ShutdownAll()
}

// refreshModCount 重新扫描指定服务器的 mods 目录
// 将最新数量写入 ServerList.json 的 info.modCount，并通过事件推送给前端
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
	modCount := api.info.ScanMods(serverPath)
	pluginCount := api.info.ScanPlugins(serverPath)
	_ = api.config.SetServerExtensionsCount(serverName, modCount, pluginCount)
	// 分别推送 mod 与插件数量给前端
	runtime.EventsEmit(api.ctx, "server:modcount", modCount)
	runtime.EventsEmit(api.ctx, "server:plugincount", pluginCount)
}
