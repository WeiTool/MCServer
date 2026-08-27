//go:build windows

package utils

import (
	"syscall"
)

// NewHiddenSysProcAttr 返回隐藏窗口的进程属性
// GUI 应用（Wails 无控制台）拉起的控制台子进程（powershell/java/jstat）若未设置
// 隐藏标志，会弹出独立控制台窗口；统一用本函数保证任何情况下都不弹窗
func NewHiddenSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000 | 0x00000200, // CREATE_NO_WINDOW | CREATE_NEW_PROCESS_GROUP
	}
}
