//go:build darwin

package window

// 本文件仅 macOS 编译。
// macOS 使用逻辑点（point），Retina 屏自动适配，因此缩放因子固定为 1.0。
// 工作区通过 Cocoa NSScreen visibleFrame 获取，用于防止窗口超出屏幕。

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa
#include <AppKit/AppKit.h>

// 获取主屏幕可见区域（去除菜单栏/程序坞）
static void getVisibleFrame(double *x, double *y, double *width, double *height) {
    NSScreen *screen = [NSScreen mainScreen];
    NSRect frame = [screen visibleFrame];
    *x = frame.origin.x;
    *y = frame.origin.y;
    *width = frame.size.width;
    *height = frame.size.height;
}
*/
import "C"

// SystemScaleFactor 获取系统 DPI 缩放因子
// macOS 使用逻辑点渲染，Retina 屏自动适配，固定返回 1.0
func SystemScaleFactor() float64 {
	return 1.0
}

// SystemWorkArea 获取主显示器工作区尺寸（去除菜单栏/程序坞）
func SystemWorkArea() (int, int) {
	var x, y, width, height C.double
	// 调用 Cocoa 获取主屏可见区域
	C.getVisibleFrame(&x, &y, &width, &height)
	// 返回工作区宽高（逻辑点）
	return int(width), int(height)
}
