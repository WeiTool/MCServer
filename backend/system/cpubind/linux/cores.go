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

// GetPerformanceCores 获取性能核心(P-Core)编号列表
// Linux 三级策略：
//  1. 读取 /sys/devices/system/cpu/cpu*/cpu_capacity（由 DT/ACPI 暴露容量等级，
//     Intel ADL/RPL 为 P 核提供较大的 capacity 值，E 核较小；AMD 所有核相等）
//     取 capacity 最大的一组逻辑 CPU 作为 P-Core。
//  2. 读取 /proc/cpuinfo 里的 "cpu MHz"（最大/当前频率）近似判断，取频率高的一组。
//  3. 若 sysfs / cpuinfo 都不可靠，回退为全部逻辑核心（对 AMD / Xeon 等无 E 核
//     的平台是完全正确的，因为所有核都同 capacity 或同频率）。
func GetPerformanceCores() ([]int, error) {
	// 方法1: cpu_capacity
	cores, err := getPCoresByCapacity()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法2: cpuinfo 频率近似
	cores, err = getPCoresByCpuinfoFreq()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法3: 全部逻辑核心
	total := GetTotalCores()
	return allCores(total), nil
}

// getPCoresByCapacity 通过 /sys/devices/system/cpu/cpuN/cpu_capacity
// 按"最大值视为 P 核"语义展开。所有同最大值的逻辑 CPU 都保留。
func getPCoresByCapacity() ([]int, error) {
	total := GetTotalCores()
	if total == 0 {
		return nil, fmt.Errorf("无法获取逻辑核心数")
	}
	capacity := make(map[int]int)
	maxCap := 0
	for cpu := 0; cpu < total; cpu++ {
		path := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpu_capacity", cpu)
		s, err := readFile(path)
		if err != nil {
			continue
		}
		v, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			continue
		}
		capacity[cpu] = v
		if v > maxCap {
			maxCap = v
		}
	}
	if len(capacity) == 0 || maxCap == 0 {
		return nil, fmt.Errorf("未读取到 cpu_capacity")
	}

	// 取 cpu_capacity 等于最大值的所有逻辑 CPU 作为 P 核
	// （AMD/服务器/无 E 核平台所有核同 capacity，即全 P-Core）
	var pCores []int
	for cpu, v := range capacity {
		if v == maxCap {
			pCores = append(pCores, cpu)
		}
	}
	if len(pCores) == 0 {
		return nil, fmt.Errorf("无法识别出 P 核")
	}
	return sortedInts(pCores), nil
}

// getPCoresByCpuinfoFreq 从 /proc/cpuinfo 读取每颗 CPU 的 "cpu MHz" / "cpu MHz dynamic" 等字段，
// 用最大频率那组作为 P 核（AMD/Intel Xeon 等频率相同即全核）。
func getPCoresByCpuinfoFreq() ([]int, error) {
	total := GetTotalCores()
	if total == 0 {
		return nil, fmt.Errorf("无法获取逻辑核心数")
	}
	content, err := readFile("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	// /proc/cpuinfo 按处理器段分隔（段间空行），逐段解析 processor 与 cpu MHz
	freqMap := make(map[int]int)
	cur := -1
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			cur = -1
			continue
		}
		if strings.HasPrefix(line, "processor") {
			// "processor : 0"
			if fields := strings.SplitN(line, ":", 2); len(fields) == 2 {
				if n, err := strconv.Atoi(strings.TrimSpace(fields[1])); err == nil {
					cur = n
				}
			}
			continue
		}
		if cur < 0 {
			continue
		}
		if strings.HasPrefix(line, "cpu MHz") {
			if fields := strings.SplitN(line, ":", 2); len(fields) == 2 {
				mhz, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
				if err == nil {
					mhzInt := int(mhz * 10) // 保留 1 位小数精度
					if _, exists := freqMap[cur]; !exists || mhzInt > freqMap[cur] {
						freqMap[cur] = mhzInt
					}
				}
			}
		}
	}
	if len(freqMap) == 0 {
		return nil, fmt.Errorf("未从 cpuinfo 解析出 CPU 频率")
	}
	maxF := 0
	for _, f := range freqMap {
		if f > maxF {
			maxF = f
		}
	}
	// 频率差值在 5% 以内认为同等级（避免睿频抖动误判）
	minF := int(float64(maxF) * 0.95)
	var pCores []int
	for cpu, f := range freqMap {
		if f >= minF {
			pCores = append(pCores, cpu)
		}
	}
	if len(pCores) == 0 {
		return nil, fmt.Errorf("频率筛选后无 P 核")
	}
	return sortedInts(pCores), nil
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

// sortedInts 将 []int 升序排序并返回（就地排序；直接返回便于链式调用）
func sortedInts(a []int) []int {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
	return a
}
