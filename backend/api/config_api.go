package api

import (
	"MCServer/backend/model"
	"MCServer/backend/storage"
	"MCServer/backend/system/sysinfo"
)

// ConfigApi 服务器配置 API
// 负责每台服务器的运行时配置：Java 路径、内存、server.properties
type ConfigApi struct {
	config     *storage.Storage
	java       *sysinfo.JavaService
	properties *storage.ServerProperties
}

// NewConfigApi 创建配置 API 实例
// 注入共享的 Storage（ServerList.json 持久化）与 JavaService（识别 Java 版本）
func NewConfigApi(store *storage.Storage, javaService *sysinfo.JavaService) *ConfigApi {
	return &ConfigApi{
		config:     store,
		java:       javaService,
		properties: storage.NewServerProperties(),
	}
}

// SetServerJava 为指定服务器设置 java.exe 路径
// 每个服务器可配置不同的 Java，存到 ServerList.json
func (api *ConfigApi) SetServerJava(serverName, executable string) error {
	return api.config.SetServerJava(serverName, executable)
}

// GetServerJava 获取指定服务器配置的 Java 信息
// 返回完整 JavaInfo（含版本），即使该路径不在扫描列表中也能识别版本
func (api *ConfigApi) GetServerJava(serverName string) (model.JavaInfo, error) {
	executable := api.config.GetServerJava(serverName)
	if executable == "" {
		return model.JavaInfo{}, nil
	}
	return api.java.AddJavaPath(executable)
}

// SetServerMemory 设置指定服务器的最大/最小内存（MB）
// 前端以 GB 输入，需转换为 MB 后传入
func (api *ConfigApi) SetServerMemory(serverName string, xmxMB, xmsMB int) error {
	return api.config.SetServerMemory(serverName, xmxMB, xmsMB)
}

// GetServerMemory 获取指定服务器的最大/最小内存（MB）
// 返回 xmxMB, xmsMB
func (api *ConfigApi) GetServerMemory(serverName string) (int, int) {
	return api.config.GetServerMemory(serverName)
}

// SetServerProperties 更新指定服务器的 server.properties 配置
// 前端传入修改后的 key=value 字典，该方法会将内容合并并保存到文件
func (api *ConfigApi) SetServerProperties(serverName string, properties map[string]string) error {
	return api.properties.Update(serverName, properties)
}

// GetServerProperties 读取指定服务器的 server.properties 全部字段
// 返回 key=value 字典；文件不存在（服务器未首次启动）时返回空 map
func (api *ConfigApi) GetServerProperties(serverName string) (map[string]string, error) {
	return api.properties.Load(serverName)
}
