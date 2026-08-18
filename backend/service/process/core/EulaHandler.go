// 本文件实现 eula.txt 的检查与自动处理逻辑
package core

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"MCServer/backend/utils"
)

// EulaStatus eula.txt 的检查结果
type EulaStatus int

const (
	// EulaReady eula.txt 已存在且已同意（eula=true），可直接启动
	EulaReady EulaStatus = iota
	// EulaNotAgreed eula.txt 存在但未同意（eula=false），需改为 true 后启动
	EulaNotAgreed
	// EulaMissing eula.txt 不存在，需等待其出现后处理
	EulaMissing
)

// CheckEula 检查指定服务器文件夹的 eula.txt 状态
// 返回 EulaReady / EulaNotAgreed / EulaMissing
func CheckEula(serverName string) EulaStatus {
	// 拼出 eula.txt 的绝对路径
	eulaPath := filepath.Join(utils.GetServerFolderPath(serverName), "eula.txt")

	// 读取文件内容
	content, err := os.ReadFile(eulaPath)
	if err != nil {
		// 文件不存在，返回缺失状态
		if os.IsNotExist(err) {
			return EulaMissing
		}
		// 其他读取错误，按缺失处理
		return EulaMissing
	}

	// 文件存在，检查是否包含 eula=true
	if strings.Contains(string(content), "eula=true") {
		return EulaReady
	}
	// 存在但未同意
	return EulaNotAgreed
}

// SetEulaAgreed 将指定服务器的 eula.txt 内容改为 eula=true
// 若文件不存在则创建
func SetEulaAgreed(serverName string) error {
	// 拼出 eula.txt 的绝对路径
	eulaPath := filepath.Join(utils.GetServerFolderPath(serverName), "eula.txt")

	// 写入 eula=true 内容（权限 0644：rw-r--r--）
	return os.WriteFile(eulaPath, []byte("eula=true\n"), 0644)
}

// WaitEulaAppear 持续轮询等待 eula.txt 出现
// 参数 interval 为轮询间隔，timeout 为最长等待时间（0 表示不超时）
// 返回 true 表示文件已出现，false 表示超时
func WaitEulaAppear(serverName string, interval, timeout time.Duration) bool {
	// 计算超时截止时间
	deadline := time.Now().Add(timeout)

	for {
		// 检查 eula.txt 是否已出现（不为缺失状态即视为已出现）
		if CheckEula(serverName) != EulaMissing {
			return true
		}

		// 若设置了超时且已超时，返回 false
		if timeout > 0 && time.Now().After(deadline) {
			return false
		}

		// 等待一个轮询间隔再检查
		time.Sleep(interval)
	}
}
