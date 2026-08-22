package model

// IoStats 指定 JVM 进程的磁盘读写速率
type IoStats struct {
	// 磁盘读取速率（字节/秒）
	ReadBytesPerSec float64 `json:"readBytesPerSec"`
	// 磁盘写入速率（字节/秒）
	WriteBytesPerSec float64 `json:"writeBytesPerSec"`
}
