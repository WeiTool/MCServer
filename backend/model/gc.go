package model

// GcStats JVM 垃圾回收统计（来自 jstat -gc）
type GcStats struct {
	// 年轻代 GC 次数（Young GC Count）
	YGC uint64 `json:"ygc"`
	// Full GC 次数（Full GC Count）
	FGC uint64 `json:"fgc"`
	// GC 总耗时（秒，GC Time）
	GCT float64 `json:"gct"`
}
