//go:build linux

// Package sysproc 提供平台相关的进程属性配置
package sysproc

import "syscall"

// NewSysProcAttr 返回 Linux 平台所需的 SysProcAttr
// 设置 Pdeathsig 信号：当 MCServer（父进程）退出时，
// Java 子进程自动收到 SIGTERM 被终止，确保关闭 APP 时所有服务端进程结束
func NewSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
}
