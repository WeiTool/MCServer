//go:build linux

// Package linux 提供 Linux 平台的 CPU 核心绑定实现
// 负责识别性能核心、查找负载最低的核心、绑定进程
package linux

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// GetPerformanceCores 获取性能核心编号列表
// Linux 通过读取 sysfs 的 core_cpus 拓扑判断物理核心（等价于 P-Core）
// 若无法识别，回退为全部逻辑核心
func GetPerformanceCores() ([]int, error) {
	cores, err := getPhysicalCores()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 回退: 全部逻辑核心
	total := GetTotalCores()
	return allCores(total), nil
}

// getPhysicalCores 从 /sys/devices/system/cpu 读取物理核心
// 通过 core_cpus 判断：拥有相同物理核心的 CPU 线程会共享 topology
func getPhysicalCores() ([]int, error) {
	total := GetTotalCores()
	if total == 0 {
		return nil, fmt.Errorf("无法获取逻辑核心数")
	}

	// 读取每个 CPU 的 core_cpus 列表，取物理核心
	var cores []int
	seen := make(map[string]bool)

	for cpu := 0; cpu < total; cpu++ {
		path := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/topology/core_cpus_list", cpu)
		content, err := readFile(path)
		if err != nil {
			// 某些系统可能没有此文件，直接跳过该核心
			continue
		}
		key := strings.TrimSpace(content)
		if !seen[key] {
			seen[key] = true
			cores = append(cores, cpu)
		}
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("未识别到物理核心")
	}
	return cores, nil
}

// FindLowestLoadCore 在指定核心范围内找出负载最低的核心
// 通过 /proc/stat 计算各逻辑核心的 CPU 使用率
func FindLowestLoadCore(coreList []int) (int, float64, error) {
	if len(coreList) == 0 {
		return 0, 0, fmt.Errorf("核心列表为空")
	}

	// 两次采样计算使用率
	prev, err := readCpuStats()
	if err != nil {
		return 0, 0, fmt.Errorf("读取 CPU 统计失败: %v", err)
	}

	bestCore := coreList[0]
	var bestLoad float64 = 999999

	for _, core := range coreList {
		load, err := sampleCoreLoad(core, prev)
		if err == nil && load < bestLoad {
			bestLoad = load
			bestCore = core
		}
	}

	if bestLoad == 999999 {
		return 0, 0, fmt.Errorf("无法获取候选核心的负载数据")
	}
	return bestCore, bestLoad, nil
}

// sampleCoreLoad 计算单个核心的使用率（基于 /proc/stat 两次采样）
func sampleCoreLoad(core int, prev map[string][4]uint64) (float64, error) {
	// 读取第二次采样
	cur, err := readCpuStats()
	if err != nil {
		return 0, err
	}

	key := fmt.Sprintf("cpu%d", core)
	prevCpu, ok1 := prev[key]
	curCpu, ok2 := cur[key]
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("缺少核心 %d 的统计", core)
	}

	prevIdle := prevCpu[3]
	curIdle := curCpu[3]
	prevTotal := prevCpu[0] + prevCpu[1] + prevCpu[2] + prevCpu[3]
	curTotal := curCpu[0] + curCpu[1] + curCpu[2] + curCpu[3]

	idleDelta := curIdle - prevIdle
	totalDelta := curTotal - prevTotal
	if totalDelta == 0 {
		return 0, nil
	}
	usage := 100.0 * (1.0 - float64(idleDelta)/float64(totalDelta))
	if usage < 0 {
		usage = 0
	}
	return usage, nil
}

// readCpuStats 读取 /proc/stat 的 CPU 统计
// 返回每个核心的 user/nice/system/idle 字段
func readCpuStats() (map[string][4]uint64, error) {
	content, err := readFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	result := make(map[string][4]uint64)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		name := fields[0]
		// 忽略汇总行 cpu（无编号）
		if name == "cpu" {
			continue
		}
		var vals [4]uint64
		for i := 0; i < 4 && i+1 < len(fields); i++ {
			vals[i], _ = strconv.ParseUint(fields[i+1], 10, 64)
		}
		result[name] = vals
	}
	return result, nil
}

// GetTotalCores 获取逻辑核心数
func GetTotalCores() int {
	return runtime.NumCPU()
}

// allCores 返回从 0 到 count-1 的核心编号列表
func allCores(count int) []int {
	if count <= 0 {
		count = 1
	}
	cores := make([]int, count)
	for i := 0; i < count; i++ {
		cores[i] = i
	}
	return cores
}
