// Package cpubind 提供 CPU 核心识别与进程亲和性绑定
// 负责识别性能核心(P-Core)、查找负载最低的核心、将进程绑定到 P 核
package cpubind

// GetPerformanceCores 获取性能核心(P-Core)编号列表
// 采用三级策略，优先使用最准确的注册表法，逐步回退
// 具体实现由各平台文件提供
func GetPerformanceCores() ([]int, error) {
	return getPerformanceCores()
}
