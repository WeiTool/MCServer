//go:build linux

package window

// 本文件仅 Linux 编译。
// Linux（GTK）下逻辑像素自动处理 HiDPI，因此缩放因子固定为 1.0。
// 工作区通过 GDK 获取，用于防止窗口超出屏幕。

/*
#cgo linux pkg-config: gtk+-3.0
#include <gtk/gtk.h>
*/
import "C"

// SystemScaleFactor 获取系统 DPI 缩放因子
// GTK 使用逻辑像素渲染，WebView 已自动适配 HiDPI，固定返回 1.0
func SystemScaleFactor() float64 {
	return 1.0
}

// SystemWorkArea 获取主显示器工作区尺寸（去除任务栏/面板）
func SystemWorkArea() (int, int) {
	// 获取默认显示器
	display := C.gdk_display_get_default()
	// 无显示器时返回默认尺寸 0
	if display == nil {
		return 0, 0
	}
	// 获取主显示器
	monitor := C.gdk_display_get_primary_monitor(display)
	// 无主显示器时返回默认尺寸 0
	if monitor == nil {
		return 0, 0
	}
	// 获取工作区矩形（逻辑像素，GTK 已按 HiDPI 缩放）
	var workarea C.GdkRectangle
	C.gdk_monitor_get_workarea(monitor, &workarea)
	// 返回工作区宽高
	return int(workarea.width), int(workarea.height)
}
