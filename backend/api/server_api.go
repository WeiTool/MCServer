package api

import (
	"context"

	"MCServer/backend/model"
	"MCServer/backend/server"
	"MCServer/backend/storage"
)

// ServerApi 服务器列表与信息 API
// 负责服务器目录扫描、统计信息（类型/版本/mod/插件数量）与活动服务器管理
type ServerApi struct {
	ctx     context.Context
	scanner *server.Scanner
	info    *server.ServerInfo
	config  *storage.Storage
}

// NewServerApi 创建服务器 API 实例
func NewServerApi(store *storage.Storage, scanner *server.Scanner) *ServerApi {
	return &ServerApi{
		scanner: scanner,
		info:    server.NewServerInfo(store, nil),
		config:  store,
	}
}

// Startup 启动初始化
func (api *ServerApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

// GetServerList 获取服务器列表
func (api *ServerApi) GetServerList() (*model.ServerListResult, error) {
	return api.scanner.ScanServers()
}

// GetServerModCount 获取指定服务器的 mod 数量
// 从后端持久化的 ServerList.json 的 info.modCount 读取
func (api *ServerApi) GetServerModCount(serverName string) (int, error) {
	return api.config.GetServerModCount(serverName), nil
}

// GetServerPluginCount 获取指定服务器的插件数量
// 从后端持久化的 ServerList.json 的 info.pluginCount 读取
func (api *ServerApi) GetServerPluginCount(serverName string) (int, error) {
	return api.config.GetServerPluginCount(serverName), nil
}

// GetServerType 获取指定服务器的类型（info.type）
// 从后端持久化的 ServerList.json 读取；未检测到则返回空字符串
func (api *ServerApi) GetServerType(serverName string) (string, error) {
	return api.info.GetType(serverName), nil
}

// GetServerVersion 获取指定服务器的版本（info.version）
// 从后端持久化的 ServerList.json 读取；未检测到则返回空字符串
func (api *ServerApi) GetServerVersion(serverName string) (string, error) {
	return api.info.GetVersion(serverName), nil
}

// SetActiveServer 设置当前活动服务器名称
// 持久化到 exe 同级 config/ServerList.json
func (api *ServerApi) SetActiveServer(name string) error {
	return api.config.SetActiveServer(name)
}

// GetActiveServer 获取当前活动服务器名称
func (api *ServerApi) GetActiveServer() (string, error) {
	return api.config.GetActiveServer()
}

// GetOnlinePlayers 获取当前活动服务器的在线玩家数量
func (api *ServerApi) GetOnlinePlayers() (int, error) {
	return api.info.GetOnlinePlayers()
}

// GetMaxPlayers 获取当前活动服务器的最大玩家数量
func (api *ServerApi) GetMaxPlayers() (int, error) {
	return api.info.GetMaxPlayers()
}

// GetPlayerList 获取当前活动服务器的完整玩家列表
func (api *ServerApi) GetPlayerList() ([]string, error) {
	return api.info.GetPlayerList()
}
