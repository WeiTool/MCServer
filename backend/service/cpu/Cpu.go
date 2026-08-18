// Package cpu 提供 CPU 核心识别的相关能力
// 负责识别性能核心(P-Core)、查找负载最低的核心
package cpu

// GetPerformanceCores 获取性能核心(P-Core)编号列表
// 采用三级策略，优先使用最准确的注册表法，逐步回退
// 具体实现由各平台文件提供
func GetPerformanceCores() ([]int, error) {
	return getPerformanceCores()
}

// FindLowestLoadCore 在指定核心范围内找出负载最低的核心
func FindLowestLoadCore(coreList []int) (int, float64, error) {
	return findLowestLoadCore(coreList)
}
