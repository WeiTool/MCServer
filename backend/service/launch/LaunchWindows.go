//go:build windows

package launch

import (
	"fmt"
	"syscall"

	"MCServer/backend/service/cpu"
)

// Windows API 相关句柄
var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess = kernel32.NewProc("OpenProcess")
	procSetAffinity = kernel32.NewProc("SetProcessAffinityMask")
	procCloseHandle = kernel32.NewProc("CloseHandle")
)

// 进程访问权限常量
const (
	processSetInformation   = 0x0200 // 允许设置进程信息(含亲和性)
	processQueryInformation = 0x0400 // 允许查询进程信息
	processQueryLimited     = 0x1000 // 允许受限查询进程信息
	processTerminate        = 0x0001 // 允许终止进程
)

// bindToLowestLoadCore Windows 平台：将进程绑定到负载最低的性能核心
func bindToLowestLoadCore(pid int) error {
	// 1. 识别性能核心
	cores, err := cpu.GetPerformanceCores()
	if err != nil {
		return fmt.Errorf("识别性能核心失败: %v", err)
	}

	// 2. 找到负载最低的性能核心
	bestCore, _, err := cpu.FindLowestLoadCore(cores)
	if err != nil {
		return fmt.Errorf("查找负载最低核心失败: %v", err)
	}

	// 3. 计算亲和性掩码 (1 << bestCore) 并绑定
	mask := uintptr(1) << uintptr(bestCore)
	return setProcessAffinity(pid, mask)
}

// openProcess 打开指定 PID 的进程句柄
// 使用完整权限标志，避免 SetProcessAffinityMask 因权限不足返回 Access denied
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

// setProcessAffinity 设置指定 PID 进程的 CPU 亲和性掩码
func setProcessAffinity(pid int, mask uintptr) error {
	handle, err := openProcess(pid)
	if err != nil {
		return err
	}
	// 关闭句柄，忽略返回值（CloseHandle 失败通常可忽略）
	defer procCloseHandle.Call(uintptr(handle))

	r, _, err := procSetAffinity.Call(uintptr(handle), mask, 0)
	if r == 0 {
		return fmt.Errorf("设置进程亲和性失败: %v", err)
	}
	return nil
}
