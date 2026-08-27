package storage

import (
	"MCServer/backend/model"
	"MCServer/backend/utils"
	"encoding/json"
	"os"
	"path/filepath"
)

// 默认全局配置：预览版默认关闭
var defaultGlobalConfig = model.GlobalConfig{PreviewEnabled: false}

// globalConfigPath 返回全局配置文件路径（exe 同级 config/global_config.json）
func globalConfigPath() string {
	return filepath.Join(utils.GetConfigDir(), "global_config.json")
}

// GetGlobalConfig 读取全局配置
// 若文件不存在，返回默认配置（PreviewEnabled = false）
func GetGlobalConfig() (model.GlobalConfig, error) {
	data, err := os.ReadFile(globalConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return defaultGlobalConfig, nil
		}
		return defaultGlobalConfig, err
	}

	var cfg model.GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultGlobalConfig, err
	}
	return cfg, nil
}

// SaveGlobalConfig 保存全局配置（确保 config 目录存在后写入）
func SaveGlobalConfig(cfg model.GlobalConfig) error {
	dir := utils.GetConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(globalConfigPath(), data, 0644)
}
