package main

import (
	"context"
	"embed"

	"MCServer/window"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// 固定窗口的逻辑尺寸为 1280x720（常量定义见 window/window.go）
	// 启动后由 window.SetupWindow 根据 DPI 缩放自适应实际物理像素大小
	err := wails.Run(&options.App{
		Title:            "MCServer",
		Width:            window.LogicalWidth,
		Height:           window.LogicalHeight,
		DisableResize:    true, // 禁止用户调整窗口大小
		Frameless:        true, // 移除原生标题栏（顶层字段，可用）
		Fullscreen:       false,
		AlwaysOnTop:      false,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Min/Max 与固定尺寸保持一致，彻底锁定窗口大小（防止边框拖动导致变形）
		MinWidth:  window.LogicalWidth,
		MaxWidth:  window.LogicalWidth,
		MinHeight: window.LogicalHeight,
		MaxHeight: window.LogicalHeight,
		OnStartup: func(ctx context.Context) {
			// 调整窗口尺寸（DPI 自适应）
			window.SetupWindow(ctx)
			// 初始化 App（保存上下文，传递状态服务）
			app.Startup(ctx)
		},
		OnBeforeClose: func(ctx context.Context) bool {
			// 关闭 APP 前停止所有正在运行的服务器进程
			app.ProcessApi.ShutdownAll()
			return false
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			// 禁用 WebView 的手势缩放，避免界面被意外缩放导致模糊
			DisablePinchZoom:     true,
			IsZoomControlEnabled: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
