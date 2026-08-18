//go:build linux

package launch

import (
	"fmt"
	"os/exec"

	"MCServer/backend/service/cpu"
)

// bindToLowestLoadCore Linux 平台：将进程绑定到负载最低的性能核心
// 使用 taskset 命令设置 CPU 亲和性，简单可靠
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

	// 3. 绑定亲和性
	return setAffinity(pid, bestCore)
}

// setAffinity 通过 taskset 命令将进程绑定到指定核心
func setAffinity(pid int, core int) error {
	cmd := exec.Command("taskset", "-pc", fmt.Sprintf("%d", core), fmt.Sprintf("%d", pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskset 绑定失败: %v (%s)", err, string(output))
	}
	return nil
}

// InitJobObject Linux 平台占位实现
// Linux 下无 Windows Job Object 概念，子进程随父进程退出由系统管理
func InitJobObject() error {
	return nil
}

// AssignProcessToJob Linux 平台占位实现
// 子进程天然受父进程生命周期管理，无需额外处理
func AssignProcessToJob(pid int) error {
	return nil
}
