package api

import (
	"MCServer/backend/service/config"
)

// ConfigApi 配置 API
// 供前端调用，读写应用配置
type ConfigApi struct {
	service *config.ConfigService
}

// NewConfigApi 创建配置 API 实例
func NewConfigApi() *ConfigApi {
	return &ConfigApi{
		service: config.NewConfigService(),
	}
}

// SetActiveServer 设置当前活动服务器名称
// 前端调用：window.go.main.App.SetActiveServer(name)
// 持久化到 exe 同级 config/ServerList.json
func (api *ConfigApi) SetActiveServer(name string) error {
	return api.service.SetActiveServer(name)
}

// GetActiveServer 获取当前活动服务器名称
// 前端调用：window.go.main.App.GetActiveServer()
func (api *ConfigApi) GetActiveServer() (string, error) {
	return api.service.GetActiveServer()
}
