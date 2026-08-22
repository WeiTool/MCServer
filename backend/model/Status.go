package model

// MemoryInfo 内存信息
type MemoryInfo struct {
	// 内存使用率百分比
	UsagePercent float64 `json:"usagePercent"`
}

// CPU P核心CPU信息
type CpuInfo struct {
	UsagePercent float64 `json:"usagePercent"`
}
