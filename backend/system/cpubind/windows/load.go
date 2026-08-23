//go:build windows

package windows

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"MCServer/backend/utils"
)

// runJSONList 执行 PowerShell 命令并解析为整数列表
// 兼容单值和多值两种 JSON 输出格式
func runJSONList(psCmd string) ([]int, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCmd)
	// 隐藏窗口：GUI 程序拉起的 powershell 不弹控制台
	cmd.SysProcAttr = utils.NewHiddenSysProcAttr()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var cores []int
	if err := json.Unmarshal(output, &cores); err != nil {
		var single int
		if err2 := json.Unmarshal(output, &single); err2 == nil {
			return []int{single}, nil
		}
		return nil, fmt.Errorf("解析失败: %v", err)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("未找到性能核心")
	}

	return cores, nil
}

// findLowestLoadCore 在指定核心范围内找出负载最低的核心
// 通过 PowerShell 获取所有逻辑核心的 PercentProcessorTime 并比较
func findLowestLoadCore(coreList []int) (int, float64, error) {
	if len(coreList) == 0 {
		return 0, 0, fmt.Errorf("核心列表为空")
	}

	psCmd := `
		$cores = Get-CimInstance -ClassName Win32_PerfFormattedData_Counters_ProcessorInformation | Where-Object { $_.Name -match '^\d+,\d+$' }

		if ($cores.Count -eq 0) {
			ConvertTo-Json @()
			exit 0
		}

		$result = @()
		foreach ($c in $cores) {
			$idx = [int]($c.Name -split ',')[1]
			$result += [PSCustomObject]@{ Core = $idx; Load = $c.PercentProcessorTime }
		}
		$result | ConvertTo-Json
	`

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCmd)
	// 隐藏窗口：GUI 程序拉起的 powershell 不弹控制台
	cmd.SysProcAttr = utils.NewHiddenSysProcAttr()
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("获取负载失败: %v", err)
	}

	var loadData []struct {
		Core int     `json:"Core"`
		Load float64 `json:"Load"`
	}
	if err := json.Unmarshal(output, &loadData); err != nil {
		var single struct {
			Core int     `json:"Core"`
			Load float64 `json:"Load"`
		}
		if err2 := json.Unmarshal(output, &single); err2 == nil {
			loadData = []struct {
				Core int     `json:"Core"`
				Load float64 `json:"Load"`
			}{{Core: single.Core, Load: single.Load}}
		} else {
			return 0, 0, fmt.Errorf("解析负载数据失败: %v", err)
		}
	}

	loadMap := make(map[int]float64)
	for _, item := range loadData {
		loadMap[item.Core] = item.Load
	}

	var bestCore int
	var bestLoad float64 = 999999
	found := false

	for _, core := range coreList {
		if load, ok := loadMap[core]; ok {
			if !found || load < bestLoad {
				bestLoad = load
				bestCore = core
				found = true
			}
		}
	}

	if !found {
		return 0, 0, fmt.Errorf("未找到任何候选核心的负载数据")
	}

	return bestCore, bestLoad, nil
}

// FindLowestLoadCore 导出包装：在指定核心范围内找出负载最低的核心
func FindLowestLoadCore(coreList []int) (int, float64, error) {
	return findLowestLoadCore(coreList)
}
