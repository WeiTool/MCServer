package api

import (
	"context"

	"MCServer/backend/model"
	"MCServer/backend/service/config"
	"MCServer/backend/service/java"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// JavaApi Java API
// 供前端调用，管理 Java 环境
type JavaApi struct {
	ctx     context.Context
	service *java.JavaService
	config  *config.ConfigService
}

// NewJavaApi 创建 Java API 实例
func NewJavaApi() *JavaApi {
	return &JavaApi{
		service: java.NewJavaService(),
		config:  config.NewConfigService(),
	}
}

// Startup 保存应用上下文
func (api *JavaApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

// ScanJavaList 扫描系统所有 Java 环境
func (api *JavaApi) ScanJavaList() []model.JavaInfo {
	return api.service.ScanSystemJava()
}

// AddJavaByDialog 打开系统文件选择框让用户选择 java.exe
func (api *JavaApi) AddJavaByDialog() (model.JavaInfo, error) {
	selection, err := runtime.OpenFileDialog(api.ctx, runtime.OpenDialogOptions{
		Title: "选择 Java 可执行文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Java 可执行文件",
				Pattern:     "java.exe",
			},
		},
	})
	if err != nil {
		return model.JavaInfo{}, err
	}
	if selection == "" {
		return model.JavaInfo{}, nil
	}

	return api.service.AddJavaPath(selection)
}

// GetJavaList 读取已保存的 Java 列表
func (api *JavaApi) GetJavaList() []model.JavaInfo {
	return api.service.GetJavaListFromConfig()
}

// SaveJavaList 保存 Java 列表到 config
func (api *JavaApi) SaveJavaList(list []model.JavaInfo) error {
	return api.service.SaveJavaListToConfig(list)
}

// SetServerJava 为指定服务器设置 java.exe 路径
// 每个服务器可配置不同的 Java，存到 ServerList.json
func (api *JavaApi) SetServerJava(serverName, executable string) error {
	return api.config.SetServerJava(serverName, executable)
}

// GetServerJava 获取指定服务器配置的 Java 信息
// 返回完整 JavaInfo（含版本），即使该路径不在扫描列表中也能识别版本
func (api *JavaApi) GetServerJava(serverName string) (model.JavaInfo, error) {
	executable := api.config.GetServerJava(serverName)
	if executable == "" {
		return model.JavaInfo{}, nil
	}
	return api.service.AddJavaPath(executable)
}

// SetServerMemory 设置指定服务器的最大/最小内存（MB）
// 前端以 GB 输入，需转换为 MB 后传入
func (api *JavaApi) SetServerMemory(serverName string, xmxMB, xmsMB int) error {
	return api.config.SetServerMemory(serverName, xmxMB, xmsMB)
}

// GetServerMemory 获取指定服务器的最大/最小内存（MB）
// 返回 xmxMB, xmsMB
func (api *JavaApi) GetServerMemory(serverName string) (int, int) {
	return api.config.GetServerMemory(serverName)
}
