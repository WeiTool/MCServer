//go:build windows

package cpubind

import (
	"fmt"
	"syscall"
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

// bindToPerformanceCores Windows 平台：将进程绑定到所有性能核心（P-Core）
// 取 P 核列表后按位或组合为多比特亲和掩码（允许进程在任意 P 核上被调度执行）
func bindToPerformanceCores(pid int) error {
	// 1. 识别性能核心（P-Core）
	cores, err := GetPerformanceCores()
	if err != nil {
		return fmt.Errorf("识别性能核心失败: %v", err)
	}
	if len(cores) == 0 {
		return fmt.Errorf("未识别到任何性能核心")
	}

	// 2. 组合多比特亲和掩码：每个核心对应一个比特位，全部置 1
	//    例如 P 核 [0,1,2,3,4,5,6,7] → mask = 0xFF = 只允许 CPU0~CPU7
	var mask uintptr
	for _, c := range cores {
		mask |= uintptr(1) << uintptr(c)
	}

	// 3. 设置进程亲和性
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
