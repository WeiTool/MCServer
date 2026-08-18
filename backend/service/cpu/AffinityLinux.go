//go:build linux

package cpu

import (
	linux "MCServer/backend/service/cpu/linux"
)

// getPerformanceCores Linux 平台：委托给 linux 子包实现
func getPerformanceCores() ([]int, error) {
	return linux.GetPerformanceCores()
}

// findLowestLoadCore Linux 平台：委托给 linux 子包实现
func findLowestLoadCore(coreList []int) (int, float64, error) {
	return linux.FindLowestLoadCore(coreList)
}

// getTotalCores Linux 平台：委托给 linux 子包实现
func getTotalCores() int {
	return linux.GetTotalCores()
}
