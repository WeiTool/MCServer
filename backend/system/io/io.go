// Package io 提供运行中 JVM 进程的磁盘读写速率统计
// 基于 gopsutil 进程 IO 计数器（累计字节数）的差值，按时间间隔计算每秒速率
package io

import (
	"fmt"
	"sync"
	"time"

	"MCServer/backend/model"

	gopsutilproc "github.com/shirou/gopsutil/v3/process"
)

// IOService 指定 JVM 进程的磁盘读写速率统计服务
type IOService struct {
	mu sync.Mutex
	// 上次采集时间（首次采集或换进程后重建基准）
	lastTime time.Time
	// 上次采集的累计读取字节数
	lastRead uint64
	// 上次采集的累计写入字节数
	lastWrite uint64
}

// NewIOService 创建磁盘读写速率统计服务实例
func NewIOService() *IOService {
	return &IOService{}
}

// GetStats 获取指定 PID 进程的磁盘读取/写入速率（字节/秒）
// 首次采集（无基准值）时只记录基准并返回 0；
// 进程重启（计数器被重置）时自动重建基准，避免速率突变
func (s *IOService) GetStats(pid int) (*model.IoStats, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("进程未运行")
	}

	proc, err := gopsutilproc.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("进程不存在: %v", err)
	}
	counters, err := proc.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("获取进程 IO 计数失败: %v", err)
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	stats := &model.IoStats{}
	// 有基准且当前计数未回退（进程未重启）时，按时间间隔计算速率
	if !s.lastTime.IsZero() &&
		counters.ReadBytes >= s.lastRead && counters.WriteBytes >= s.lastWrite {
		elapsed := now.Sub(s.lastTime).Seconds()
		if elapsed > 0 {
			stats.ReadBytesPerSec = float64(counters.ReadBytes-s.lastRead) / elapsed
			stats.WriteBytesPerSec = float64(counters.WriteBytes-s.lastWrite) / elapsed
		}
	}

	// 记录本次采集值作为下次基准
	s.lastTime = now
	s.lastRead = counters.ReadBytes
	s.lastWrite = counters.WriteBytes

	return stats, nil
}
