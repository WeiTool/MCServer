//go:build windows

package cpubind

import (
	win "MCServer/backend/system/cpubind/windows"
)

// getPerformanceCores Windows 平台：委托给 windows 子包实现
func getPerformanceCores() ([]int, error) {
	return win.GetPerformanceCores()
}

// findLowestLoadCore Windows 平台：委托给 windows 子包实现
func findLowestLoadCore(coreList []int) (int, float64, error) {
	return win.FindLowestLoadCore(coreList)
}
