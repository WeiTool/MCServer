package api

import (
	"context"

	"MCServer/backend/model"
	"MCServer/backend/service/config"
	"MCServer/backend/service/serverList"
)

// ServerListApi 服务器列表 API
type ServerListApi struct {
	ctx     context.Context
	scanner *serverList.ServerScanner
	config  *config.ConfigService
}

// NewServerListApi 创建 API 实例
func NewServerListApi() *ServerListApi {
	return &ServerListApi{
		scanner: serverList.NewServerScanner(),
		config:  config.NewConfigService(),
	}
}

// Startup 启动初始化
func (api *ServerListApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

// GetServerList 获取服务器列表
func (api *ServerListApi) GetServerList() (*model.ServerListResult, error) {
	return api.scanner.ScanServers()
}

// GetServerModCount 获取指定服务器的 mod 数量
// 从后端持久化的 ServerList.json 的 info.modCount 读取
func (api *ServerListApi) GetServerModCount(serverName string) (int, error) {
	return api.config.GetServerModCount(serverName), nil
}

// GetServerPluginCount 获取指定服务器的插件数量
// 从后端持久化的 ServerList.json 的 info.pluginCount 读取
func (api *ServerListApi) GetServerPluginCount(serverName string) (int, error) {
	return api.config.GetServerPluginCount(serverName), nil
}

// GetServerType 获取指定服务器的类型（info.type）
// 从后端持久化的 ServerList.json 读取；未检测到则返回空字符串
func (api *ServerListApi) GetServerType(serverName string) (string, error) {
	return api.config.GetServerType(serverName), nil
}

// GetServerVersion 获取指定服务器的版本（info.version）
// 从后端持久化的 ServerList.json 读取；未检测到则返回空字符串
func (api *ServerListApi) GetServerVersion(serverName string) (string, error) {
	return api.config.GetServerVersion(serverName), nil
}
