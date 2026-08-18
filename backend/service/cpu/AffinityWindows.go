//go:build windows

package cpu

import (
	win "MCServer/backend/service/cpu/windows"
)

// getPerformanceCores Windows 平台：委托给 windows 子包实现
func getPerformanceCores() ([]int, error) {
	return win.GetPerformanceCores()
}

// findLowestLoadCore Windows 平台：委托给 windows 子包实现
func findLowestLoadCore(coreList []int) (int, float64, error) {
	return win.FindLowestLoadCore(coreList)
}

// getTotalCores Windows 平台：委托给 windows 子包实现
func getTotalCores() int {
	return win.GetTotalCores()
}
