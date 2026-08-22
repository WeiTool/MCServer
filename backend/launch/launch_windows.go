//go:build windows

package launch

import (
	"fmt"
	"syscall"
)

// Windows API 相关句柄（作业对象打开/关闭进程）
var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess = kernel32.NewProc("OpenProcess")
	procCloseHandle = kernel32.NewProc("CloseHandle")
)

// 进程访问权限常量
const (
	processSetInformation   = 0x0200
	processQueryInformation = 0x0400
	processQueryLimited     = 0x1000
	processTerminate        = 0x0001
)

// openProcess 打开指定 PID 的进程句柄
func openProcess(pid int) (syscall.Handle, error) {
	handle, _, err := procOpenProcess.Call(
		uintptr(processSetInformation|processQueryInformation|processQueryLimited|processTerminate),
		uintptr(0),
		uintptr(pid),
	)
	if handle == 0 {
		return 0, fmt.Errorf("打开进程失败(PID=%d): %v", pid, err)
	}
	return syscall.Handle(handle), nil
}
