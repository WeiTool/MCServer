// 日志导出 API
// 弹出系统保存对话框，将内容写入用户选择的文件（供控制台导出日志/错误日志）
package api

import (
	"context"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportApi 日志导出 API
type ExportApi struct {
	ctx context.Context
}

// NewExportApi 创建日志导出 API 实例
func NewExportApi() *ExportApi {
	return &ExportApi{}
}

// Startup 保存应用上下文
func (api *ExportApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

// SaveLogToFile 弹出保存对话框，将内容写入用户选择的文件
// 返回实际保存的文件路径；用户取消返回空字符串
func (api *ExportApi) SaveLogToFile(defaultName, content string) (string, error) {
	selection, err := runtime.SaveFileDialog(api.ctx, runtime.SaveDialogOptions{
		Title:           "导出日志",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "日志文件", Pattern: "*.log"},
			{DisplayName: "文本文件", Pattern: "*.txt"},
		},
	})
	if err != nil {
		return "", err
	}
	if selection == "" {
		// 用户取消，不算错误
		return "", nil
	}
	if err := os.WriteFile(selection, []byte(content), 0644); err != nil {
		return "", err
	}
	return selection, nil
}
