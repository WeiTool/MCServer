//go:build windows

// Package windows 提供 Windows 平台的 CPU 核心绑定实现
// 负责识别性能核心(P-Core)、查找负载最低的核心、绑定进程
package windows

import (
	"strings"
)

// GetPerformanceCores 获取性能核心(P-Core)编号列表
// 采用三级策略，优先使用最准确的注册表法，逐步回退
func GetPerformanceCores() ([]int, error) {
	// 方法1: 尝试从注册表读取 EfficientClass (最准确)
	cores, err := getCoresFromRegistry()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法2: 直接使用 CPU 型号推断 (最可靠)
	cores, err = getCoresFallback()
	if err == nil && len(cores) > 0 {
		return cores, nil
	}

	// 方法3: 最终回退 - 全部核心
	total := GetTotalCores()
	cores = make([]int, total)
	for i := 0; i < total; i++ {
		cores[i] = i
	}
	return cores, nil
}

// getCoresFromRegistry 通过 PowerShell 读取注册表 EfficientClass 识别 P-Core
// 值为 0 表示性能核心(P-Core)
func getCoresFromRegistry() ([]int, error) {
	psCmd := `
		$pCores = @()
		$index = 0
		while ($true) {
			$path = "HKLM:\HARDWARE\DESCRIPTION\System\CentralProcessor\$index"
			try {
				$class = (Get-ItemProperty -Path $path -Name "EfficientClass" -ErrorAction Stop).EfficientClass
				if ($class -eq 0) {
					$pCores += $index
				}
			} catch {
				break
			}
			$index++
		}

		if ($pCores.Count -gt 0) {
			$pCores = $pCores | Sort-Object
			$pCores | ConvertTo-Json
		} else {
			ConvertTo-Json @()
		}
	`
	return runJSONList(psCmd)
}

// getCoresFallback 基于已知 CPU 型号推断性能核心
// 当注册表法失败时作为回退方案
func getCoresFallback() ([]int, error) {
	total := GetTotalCores()
	cpuName := getCPUName()

	// ============================================
	// 已知 CPU 型号的精确配置
	// ============================================

	// Intel 12代 移动端 (Alder Lake)
	if strings.Contains(cpuName, "12th Gen Intel") {
		// i5-12450H: 4 P-Core + 4 E-Core, 总共12逻辑核心
		if strings.Contains(cpuName, "i5-12450H") {
			return []int{0, 1, 2, 3, 4, 5, 6, 7}, nil
		}
		// 其他 12代 型号，按比例计算
		return coresByRatio(total), nil
	}

	// Intel 13代 移动端 (Raptor Lake)
	if strings.Contains(cpuName, "13th Gen Intel") {
		return coresByRatio(total), nil
	}

	// Intel 14代 移动端 (Meteor Lake)
	if strings.Contains(cpuName, "14th Gen Intel") {
		return coresByRatio(total), nil
	}

	// AMD 处理器 (全部是性能核心)
	if strings.Contains(cpuName, "AMD") {
		return allCores(total), nil
	}

	// 通用推断: 如果总核心数 > 8, 取前 2/3, 否则全部
	return coresByRatio(total), nil
}

// coresByRatio 按 2/3 比例返回核心列表
// 如果总核心数 <= 8，返回全部核心
func coresByRatio(total int) []int {
	var count int
	if total > 8 {
		count = total * 2 / 3
	} else {
		count = total
	}
	return allCores(count)
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
