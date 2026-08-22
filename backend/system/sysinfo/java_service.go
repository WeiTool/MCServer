// Package java 提供 Java 运行环境的检测能力
// 负责扫描系统 Java、识别版本（每台服务器的 Java 由配置单独指定）
package sysinfo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"MCServer/backend/model"
)

// JavaService Java 服务
// 负责扫描和识别系统中的 Java 环境
type JavaService struct {
}

// NewJavaService 创建 Java 服务实例
func NewJavaService() *JavaService {
	return &JavaService{}
}

// ScanSystemJava 扫描系统安装的所有 Java
// 返回 Java 列表（去重后的唯一路径）
func (s *JavaService) ScanSystemJava() []model.JavaInfo {
	// 收集所有候选 java.exe 路径
	exePaths := s.findJavaExecutables()

	// 去重
	seen := make(map[string]bool)
	var list []model.JavaInfo
	for _, exe := range exePaths {
		if seen[exe] {
			continue
		}
		seen[exe] = true

		if info, ok := s.detectJava(exe); ok {
			list = append(list, info)
		}
	}

	// 按版本号排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Version < list[j].Version
	})

	return list
}

// AddJavaPath 手动添加一个 Java 路径并识别版本
// path 可以是 java.exe 完整路径或 Java 安装目录
func (s *JavaService) AddJavaPath(path string) (model.JavaInfo, error) {
	exePath := s.resolveExecutable(path)
	if exePath == "" {
		return model.JavaInfo{}, fmt.Errorf("无法找到 java.exe: %s", path)
	}

	info, ok := s.detectJava(exePath)
	if !ok {
		return model.JavaInfo{}, fmt.Errorf("无法识别该 Java 版本: %s", exePath)
	}
	return info, nil
}

// resolveExecutable 解析传入路径为 java.exe 完整路径
// 支持传入 java.exe 路径或 Java 安装目录
func (s *JavaService) resolveExecutable(path string) string {
	// 若直接指向 java.exe
	if strings.EqualFold(filepath.Base(path), "java.exe") {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 若是安装目录，尝试 bin\java.exe
	candidates := []string{
		filepath.Join(path, "bin", "java.exe"),
		filepath.Join(path, "java.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// detectJava 检测指定 java.exe 的版本信息
func (s *JavaService) detectJava(exePath string) (model.JavaInfo, bool) {
	versionName, ok := s.getJavaVersion(exePath)
	if !ok {
		return model.JavaInfo{}, false
	}

	mainVersion := parseMajorVersion(versionName)

	return model.JavaInfo{
		Path:        filepath.Dir(filepath.Dir(exePath)), // bin 的上一级 = 安装目录
		Executable:  exePath,
		Version:     mainVersion,
		VersionName: versionName,
	}, true
}

// getJavaVersion 执行 java.exe -version 获取版本号
func (s *JavaService) getJavaVersion(exePath string) (string, bool) {
	cmd := exec.Command(exePath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", false
	}

	// 从输出中提取版本号
	// 格式1: openjdk version "17.0.1" 2021-10-19
	// 格式2: java version "1.8.0_291"
	text := string(output)
	if idx := strings.Index(text, `"`); idx >= 0 {
		rest := text[idx+1:]
		if end := strings.Index(rest, `"`); end >= 0 {
			return rest[:end], true
		}
	}
	return "", false
}

// parseMajorVersion 从完整版本号解析主版本号
// "1.8.0_291" -> 8, "17.0.1" -> 17
func parseMajorVersion(version string) int {
	// 去掉可能的 "1." 前缀（Java 8 及以下格式），无该前缀时原样返回
	v := strings.TrimPrefix(version, "1.")
	// 取第一个点前的数字
	first := strings.SplitN(v, ".", 2)[0]
	// 去掉可能的后缀（如 "8.0_291" 里的下划线部分）
	first = strings.SplitN(first, "_", 2)[0]
	n, err := strconv.Atoi(first)
	if err != nil {
		return 0
	}
	return n
}

// findJavaExecutables 查找系统中所有可能的 java.exe
// 搜索常见安装目录
func (s *JavaService) findJavaExecutables() []string {
	var results []string
	added := make(map[string]bool)

	add := func(path string) {
		if path == "" || added[path] {
			return
		}
		if _, err := os.Stat(path); err == nil {
			results = append(results, path)
			added[path] = true
		}
	}

	// 检查 PATH 中的 java
	if path, err := exec.LookPath("java"); err == nil {
		add(path)
	}

	// 搜索常见 Java 安装目录
	programFiles := []string{
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
	}
	for _, pf := range programFiles {
		if pf == "" {
			continue
		}
		javaDir := filepath.Join(pf, "Java")
		s.searchDirForJava(javaDir, add)

		// Eclipse Adoptium (Eclipse Temurin)
		adoptiumDir := filepath.Join(pf, "Eclipse Adoptium")
		s.searchDirForJava(adoptiumDir, add)

		// Microsoft OpenJDK
		microsoftDir := filepath.Join(pf, "Microsoft")
		s.searchDirForJava(microsoftDir, add)

		// Amazon Corretto
		correttoDir := filepath.Join(pf, "Amazon Corretto")
		s.searchDirForJava(correttoDir, add)

		// Zulu
		zuluDir := filepath.Join(pf, "Zulu")
		s.searchDirForJava(zuluDir, add)
	}

	return results
}

// searchDirForJava 递归搜索目录下最多一层子目录中的 java.exe
func (s *JavaService) searchDirForJava(dir string, add func(string)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// 检查该子目录的 bin\java.exe
		exe := filepath.Join(dir, entry.Name(), "bin", "java.exe")
		add(exe)
	}
}
