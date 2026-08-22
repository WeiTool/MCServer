package detector

import (
	"MCServer/backend/storage"
	"MCServer/backend/utils"
	"fmt"
	"time"
)

// GetVersionWithState 带状态机的版本获取（配合 DetectionState 使用）
// 如果版本检测未完成，持续尝试直到完成
func GetVersionWithState(state *DetectionState, port string) (string, error) {
	// 检查是否需要获取版本
	if !state.NeedVersion() {
		// 已经完成或不需要版本检测
		return "", nil
	}

	serverName := state.ServerName()

	// 确保服务启用
	err := EnsureServicesEnabled(serverName)
	if err != nil {
		return "", fmt.Errorf("启用服务失败: %w", err)
	}

	// 持续获取版本信息，直到成功
	const retryInterval = 2 * time.Second

	for {
		version := utils.GetStatusVersion(port)
		if version != "" {
			// 获取成功，标记完成
			state.MarkVersionDone()
			return version, nil
		}

		// 未获取到版本，等待后继续尝试
		time.Sleep(retryInterval)
	}
}

// EnsureServicesEnabled 确保 enable-query 为 true
// 直接强制设置 enable-query=true 并保存到文件
func EnsureServicesEnabled(serverName string) error {
	props := storage.NewServerProperties()

	// 直接强制设置 enable-query 为 true
	updates := map[string]string{
		"enable-query": "true",
	}

	return props.Update(serverName, updates)
}
