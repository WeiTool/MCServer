package api

import (
	"context"

	"MCServer/backend/model"
	"MCServer/backend/system/sysinfo"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// JavaApi Java 环境 API
// 负责系统 Java 环境的扫描与手动添加（每台服务器的 Java 由 ConfigApi 单独配置）
type JavaApi struct {
	ctx     context.Context
	service *sysinfo.JavaService
}

// NewJavaApi 创建 Java 环境 API 实例
func NewJavaApi(javaService *sysinfo.JavaService) *JavaApi {
	return &JavaApi{
		service: javaService,
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
