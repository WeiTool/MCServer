// Package window 提供窗口尺寸自适应相关的配置逻辑
// 包含窗口逻辑尺寸常量、启动时尺寸调整函数
// 以及各平台（Windows / Linux / macOS）的 DPI 与工作区实现
package window

import (
	"context"
	"math"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 固定窗口的逻辑尺寸（供 main.go 窗口配置和 SetupWindow 共用）
const (
	// LogicalWidth 窗口逻辑宽度
	LogicalWidth = 1280
	// LogicalHeight 窗口逻辑高度
	LogicalHeight = 720
)

// SetupWindow 应用启动时调整窗口尺寸
// 需要持有 context，因此由 main.go 的 OnStartup 回调调用
func SetupWindow(ctx context.Context) {
	// 计算最终缩放因子：
	//  - 系统 DPI 缩放：保证内容清晰不模糊（高 DPI 屏渲染更多像素）；
	//  - 屏幕适配缩放：保证窗口不超过屏幕可用区域（低分辨率屏不会过大）。
	// 取两者较小值，从而同时满足“防模糊”和“防过大”。
	dpiScale := SystemScaleFactor()
	workW, workH := SystemWorkArea()

	fitScale := 1.0
	if LogicalWidth > 0 && LogicalHeight > 0 {
		// 按宽度适配的缩放
		scaleW := float64(workW) / float64(LogicalWidth)
		// 按高度适配的缩放
		scaleH := float64(workH) / float64(LogicalHeight)
		// 取两者较小值
		if scaleW < fitScale {
			fitScale = scaleW
		}
		// 高度适配进一步收紧
		if scaleH < fitScale {
			fitScale = scaleH
		}
	}

	// 最终缩放取 DPI 缩放与适配缩放的较小值
	scale := dpiScale
	if fitScale < scale {
		scale = fitScale
	}

	// 计算实际物理像素尺寸
	width := int(math.Round(float64(LogicalWidth) * scale))
	height := int(math.Round(float64(LogicalHeight) * scale))

	// 锁定窗口最小、最大与实际尺寸一致
	runtime.WindowSetMinSize(ctx, width, height)
	runtime.WindowSetMaxSize(ctx, width, height)
	// 设置实际窗口尺寸
	runtime.WindowSetSize(ctx, width, height)
	// 窗口居中显示
	runtime.WindowCenter(ctx)
}
