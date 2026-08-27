//go:build linux

package cpubind

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// bindToPerformanceCores Linux 平台：将进程绑定到所有性能核心（P-Core）
// 直接调用 sched_setaffinity 系统调用设置多核心亲和性，无需外部命令
func bindToPerformanceCores(pid int) error {
	// 1. 识别性能核心（P-Core）
	cores, err := GetPerformanceCores()
	if err != nil {
		return fmt.Errorf("识别性能核心失败: %v", err)
	}
	if len(cores) == 0 {
		return fmt.Errorf("未识别到任何性能核心")
	}

	// 2. 构造 CPU 亲和性掩码：每个 P 核对应一个比特位
	var set unix.CPUSet
	for _, c := range cores {
		set.Set(c)
	}

	// 3. 绑定亲和性（允许运行在所有 P 核上）
	if err := unix.SchedSetaffinity(pid, &set); err != nil {
		return fmt.Errorf("sched_setaffinity 绑定失败: %v", err)
	}
	return nil
}
