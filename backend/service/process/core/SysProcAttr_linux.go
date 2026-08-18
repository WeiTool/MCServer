//go:build linux

package core

import "syscall"

// newSysProcAttr 返回 Linux 平台所需的 SysProcAttr
// 设置 Pdeathsig 信号：当 MCServer（父进程）退出时，
// Java 子进程自动收到 SIGTERM 被终止，确保关闭 APP 时所有服务端进程结束
func newSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
}
