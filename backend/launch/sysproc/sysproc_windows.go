//go:build windows

// Package sysproc 提供平台相关的进程属性配置
package sysproc

import "syscall"

const (
	// createNoWindow 表示进程不会新建控制台窗口（CREATE_NO_WINDOW）
	createNoWindow = 0x08000000
	// createNewProcessGroup 让子进程处于新的进程组，便于统一管理（CREATE_NEW_PROCESS_GROUP）
	createNewProcessGroup = 0x00000200
)

// NewSysProcAttr 返回 Windows 平台所需的 SysProcAttr
// Windows 下通过 Job Object（KILL_ON_JOB_CLOSE）管理子进程生命周期，
// 同时设置 HideWindow + CREATE_NO_WINDOW，避免启动 Java 服务器时弹出多余的 cmd/powershell 窗口
func NewSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow | createNewProcessGroup,
	}
}
