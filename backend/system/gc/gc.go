// Package gc 提供 JVM 垃圾回收统计的采集能力
// 基于 JDK 自带的 jstat 命令（jstat -gc <pid>）解析 YGC/FGC/GCT，
// 供前端绘制 GC 折线图
package gc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"MCServer/backend/model"
)

// GCService JVM GC 统计服务
// 通过 jstat 查询运行中 JVM 的 GC 状态（jstat 位于 Java 安装目录的 bin 下）
type GCService struct {
	// Java 可执行文件所在目录（bin），用于定位同目录下的 jstat
	javaBin string
}

// NewGCService 创建 GC 统计服务实例
func NewGCService() *GCService {
	return &GCService{}
}

// SetJavaBin 设置 Java 可执行文件所在目录（bin）
// 进程启动时由进程管理器调用，使 jstat 定位到与所用 JDK 同目录的版本
func (s *GCService) SetJavaBin(dir string) {
	s.javaBin = dir
}

// GetStats 获取指定 PID 的 JVM GC 统计
// 返回年轻代 GC 次数（YGC）、Full GC 次数（FGC）、GC 总耗时秒数（GCT）
func (s *GCService) GetStats(pid int) (*model.GcStats, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("进程未运行")
	}

	jstat, err := s.locateJstat()
	if err != nil {
		return nil, err
	}

	out, err := exec.Command(jstat, "-gc", strconv.Itoa(pid)).Output()
	if err != nil {
		return nil, fmt.Errorf("jstat 查询失败(PID=%d): %v", pid, err)
	}

	return parseJstatGC(string(out))
}

// locateJstat 定位 jstat 可执行文件
// 优先使用配置 Java 同目录下的 jstat（版本与运行中的 JVM 一致）；
// 找不到时回退到 PATH 中的 jstat
func (s *GCService) locateJstat() (string, error) {
	if s.javaBin != "" {
		candidate := filepath.Join(s.javaBin, jstatExecutable())
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if path, err := exec.LookPath("jstat"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("未找到 jstat，请确认 JDK 完整安装（bin 下含 jstat）")
}

// parseJstatGC 解析 jstat -gc 输出
// 输出为两行：表头 + 数据，列按空格分隔；
// 按表头动态定位 YGC/FGC/GCT 列，避免依赖固定列号
func parseJstatGC(output string) (*model.GcStats, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("jstat 输出格式异常: %s", output)
	}

	header := strings.Fields(lines[0])
	data := strings.Fields(lines[1])
	if len(data) != len(header) {
		return nil, fmt.Errorf("jstat 输出列数不匹配")
	}

	// 按表头名定位列号
	colIndex := func(name string) (int, error) {
		for i, h := range header {
			if h == name {
				return i, nil
			}
		}
		return -1, fmt.Errorf("jstat 输出缺少 %s 列", name)
	}

	idxYGC, err := colIndex("YGC")
	if err != nil {
		return nil, err
	}
	idxFGC, err := colIndex("FGC")
	if err != nil {
		return nil, err
	}
	idxGCT, err := colIndex("GCT")
	if err != nil {
		return nil, err
	}

	ygc, err := strconv.ParseUint(data[idxYGC], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析 YGC 失败: %v", err)
	}
	fgc, err := strconv.ParseUint(data[idxFGC], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析 FGC 失败: %v", err)
	}
	gct, err := strconv.ParseFloat(data[idxGCT], 64)
	if err != nil {
		return nil, fmt.Errorf("解析 GCT 失败: %v", err)
	}

	return &model.GcStats{YGC: ygc, FGC: fgc, GCT: gct}, nil
}
