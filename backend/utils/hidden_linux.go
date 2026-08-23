//go:build linux

package utils

import "syscall"

// NewHiddenSysProcAttr Linux 平台无窗口概念，返回 nil
func NewHiddenSysProcAttr() *syscall.SysProcAttr {
	return nil
}
