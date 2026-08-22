//go:build linux

package cpubind

import (
	linux "MCServer/backend/system/cpubind/linux"
)

// getPerformanceCores Linux 平台：委托给 linux 子包实现
func getPerformanceCores() ([]int, error) {
	return linux.GetPerformanceCores()
}

// findLowestLoadCore Linux 平台：委托给 linux 子包实现
func findLowestLoadCore(coreList []int) (int, float64, error) {
	return linux.FindLowestLoadCore(coreList)
}
