// Package cpubind 提供 CPU 核心识别与进程亲和性绑定
// 负责识别性能核心(P-Core)、查找负载最低的核心、将进程绑定到 P 核
package cpubind

import "sync"

// pcoreCache 缓存已识别的 P-Core 列表（核心布局运行时不会变化）
// 避免每 2 秒的系统状态推送都重复执行高开销的 PowerShell/注册表查询
var (
	pcoreMu    sync.Mutex
	pcoreCache []int
)

// GetPerformanceCores 获取性能核心(P-Core)编号列表
// 采用三级策略，优先使用最准确的注册表法，逐步回退
// 具体实现由各平台文件提供；成功识别后缓存复用
func GetPerformanceCores() ([]int, error) {
	pcoreMu.Lock()
	defer pcoreMu.Unlock()

	if pcoreCache != nil {
		return pcoreCache, nil
	}

	cores, err := getPerformanceCores()
	if err != nil {
		return nil, err
	}
	pcoreCache = cores
	return cores, nil
}
