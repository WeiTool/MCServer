package api

import (
	"context"

	"MCServer/backend/model"
	"MCServer/backend/server"
)

// ExtractApi 服务器压缩包解压 API
// 负责将用户选择的 Minecraft 服务器压缩包解压到 servers 目录
type ExtractApi struct {
	ctx context.Context
}

// NewExtractApi 创建解压 API 实例
func NewExtractApi() *ExtractApi {
	return &ExtractApi{}
}

// Startup 启动初始化
func (api *ExtractApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

// ExtractServerZip 解压服务器压缩包到指定服务器目录
func (api *ExtractApi) ExtractServerZip(zipPath, serverName string) (model.ExtractResponse, error) {
	return server.ExtractServerZip(zipPath, serverName)
}
