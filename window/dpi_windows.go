//go:build windows

package window

// 本文件仅 Windows 编译。
// 提供 Windows 平台的系统缩放因子与工作区获取实现。

import (
	"syscall"
	"unsafe"
)

// Windows 系统 API 常量
const (
	// 获取主显示器工作区（去掉任务栏）
	spiGetWorkArea = 0x0030
)

var (
	// 用户32系统库（Windows API 入口）
	user32 = syscall.NewLazyDLL("user32.dll")
	// GetDpiForSystem 获取系统 DPI（仅 Win10+）
	procGetDpiForSystem = user32.NewProc("GetDpiForSystem")
	// SystemParametersInfo 获取系统参数（如工作区）
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

// RECT 屏幕矩形区域（Windows API 结构）
type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// SystemScaleFactor 获取系统 DPI 缩放因子（仅 Win10+ 生效）
func SystemScaleFactor() float64 {
	// 调用 GetDpiForSystem 获取系统 DPI
	dpi, _, _ := procGetDpiForSystem.Call()
	// DPI 为 0 时视为缩放 1.0（不缩放）
	if dpi == 0 {
		return 1.0
	}
	// 默认 DPI 为 96，缩放因子 = 实际 DPI / 96
	return float64(dpi) / 96.0
}

// SystemWorkArea 获取主显示器工作区尺寸（去除任务栏）
func SystemWorkArea() (int, int) {
	var r rect
	// 调用 SystemParametersInfo 获取主屏工作区
	procSystemParametersInfo.Call(
		uintptr(spiGetWorkArea),
		0,
		uintptr(unsafe.Pointer(&r)),
		0,
	)
	// 工作区宽 = 右 - 左
	width := int(r.Right - r.Left)
	// 工作区高 = 下 - 上
	height := int(r.Bottom - r.Top)
	return width, height
}
