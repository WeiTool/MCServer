package api

import (
	"MCServer/backend/service"
	"context"

	"MCServer/backend/model"
)

// FileOperationApi 服务器压缩包解压 API
// 负责将用户选择的 Minecraft 服务器压缩包解压到 servers 目录
type FileOperationApi struct {
	ctx context.Context
}

// NewExtractApi 创建解压 API 实例
func NewFileOperationApi() *FileOperationApi {
	return &FileOperationApi{}
}

// Startup 启动初始化
func (api *FileOperationApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

// ExtractServerZip 解压服务器压缩包到指定服务器目录
func (api *FileOperationApi) ExtractServerZip(zipPath, serverName string) (model.FileOperationResponse, error) {
	return service.ExtractServerZip(zipPath, serverName)
}

// CopyJarFile 复制服务器jar到指定目录
func (api *FileOperationApi) CopyJarFile(jarPath, serverName string) (model.FileOperationResponse, error) {
	return service.CopyJarFile(jarPath, serverName)
}
