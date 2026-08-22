// Package sysinfo 提供主机系统信息查询（内存使用率、P 核 CPU 使用率）
package sysinfo

import (
	"fmt"
	"time"

	"MCServer/backend/model"
	"MCServer/backend/system/cpubind"

	gopsutilcpu "github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// GetMemoryInfo 获取内存信息
// 返回包含内存使用率百分比的内存信息指针
func GetMemoryInfo() (*model.MemoryInfo, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &model.MemoryInfo{
		UsagePercent: memInfo.UsedPercent,
	}, nil
}

// GetJVMProcessMemoryUsage 获取指定PID的JVM进程内存使用率（占系统总内存的百分比）
func GetJVMProcessMemoryUsage(pid int32) (float64, error) {
	// 获取进程信息
	proc, err := process.NewProcess(pid)
	if err != nil {
		return 0, fmt.Errorf("进程不存在: %v", err)
	}

	// 获取进程内存（RSS）
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return 0, fmt.Errorf("获取进程内存失败: %v", err)
	}

	// 获取系统总内存
	sysMem, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("获取系统内存失败: %v", err)
	}

	// 计算使用率
	usagePercent := float64(memInfo.RSS) / float64(sysMem.Total) * 100
	return usagePercent, nil
}

// GetCPUInfo 获取CPU(P)信息
func GetCPUInfo() (*model.CpuInfo, error) {
	avgUsage, err := GetPCoresAverageUsage()
	if err != nil {
		return nil, err
	}
	return &model.CpuInfo{UsagePercent: avgUsage}, nil
}

// GetPCoresAverageUsage 返回性能核心（P-Core）的平均 CPU 使用率（百分比，0-100）
// 内部组合了 GetPerformanceCores 和 gopsutil，无需外部额外调用
func GetPCoresAverageUsage() (float64, error) {
	pCores, err := cpubind.GetPerformanceCores()
	if err != nil {
		return 0, fmt.Errorf("获取 P 核心列表失败: %v", err)
	}
	if len(pCores) == 0 {
		return 0, fmt.Errorf("未找到任何 P 核心")
	}

	allUsage, err := gopsutilcpu.Percent(200*time.Millisecond, true)
	if err != nil {
		return 0, fmt.Errorf("获取 CPU 使用率失败: %v", err)
	}

	var total float64
	var count int
	for _, core := range pCores {
		if core >= 0 && core < len(allUsage) {
			total += allUsage[core]
			count++
		}
	}

	if count == 0 {
		return 0, fmt.Errorf("没有有效的 P 核心数据（编号越界或采样失败）")
	}

	return total / float64(count), nil
}
