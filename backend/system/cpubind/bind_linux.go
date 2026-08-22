//go:build linux

package cpubind

import (
	"fmt"
	"os/exec"
	"strings"
)

// bindToPerformanceCores Linux 平台：将进程绑定到所有性能核心（P-Core）
// 使用 taskset -pc <列表> <pid> 设置多核心亲和性，简单可靠
func bindToPerformanceCores(pid int) error {
	// 1. 识别性能核心（P-Core）
	cores, err := GetPerformanceCores()
	if err != nil {
		return fmt.Errorf("识别性能核心失败: %v", err)
	}
	if len(cores) == 0 {
		return fmt.Errorf("未识别到任何性能核心")
	}

	// 2. 把核心数组拼接成逗号分隔字符串：[0,1,2,3] → "0,1,2,3"
	parts := make([]string, 0, len(cores))
	for _, c := range cores {
		parts = append(parts, fmt.Sprintf("%d", c))
	}
	list := strings.Join(parts, ",")

	// 3. 绑定亲和性（允许运行在所有 P 核上）
	return setAffinityList(pid, list)
}

// setAffinityList 通过 taskset 命令将进程绑定到指定 CPU 列表（逗号分隔）
// 例：taskset -pc 0,1,2,3,4,5,6,7 12345
func setAffinityList(pid int, list string) error {
	cmd := exec.Command("taskset", "-pc", list, fmt.Sprintf("%d", pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskset 绑定失败: %v (%s)", err, string(output))
	}
	return nil
}
