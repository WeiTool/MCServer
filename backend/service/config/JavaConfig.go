// Package config 提供应用配置的持久化能力
// 本文件负责服务器内存配置（最大/最小内存）的读写
package config

import (
	"MCServer/backend/model"
)

// SetServerMemory 设置指定服务器的最大/最小内存（MB）
// 前端以 GB 输入，需先转换为 MB 再传入
func (s *ConfigService) SetServerMemory(serverName string, xmxMB, xmsMB int) error {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return err
	}

	// 按服务器名查找，若存在则更新其内存配置
	idx, found := findServerByName(list, serverName)
	if found {
		list[idx].Config.Xmx = xmxMB
		list[idx].Config.Xms = xmsMB
	} else {
		// 若服务器不在列表中，追加一个新条目
		list = append(list, model.ServerConfig{
			ServerName: serverName,
			Config: model.ServerConfigExtra{
				Xmx: xmxMB,
				Xms: xmsMB,
			},
		})
	}

	return s.SaveServerList(list)
}

// GetServerMemory 获取指定服务器的最大/最小内存（MB）
// 返回 xmxMB, xmsMB；未配置时返回 0
func (s *ConfigService) GetServerMemory(serverName string) (int, int) {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return 0, 0
	}
	// 按服务器名查找，返回其内存配置
	idx, found := findServerByName(list, serverName)
	if found {
		return list[idx].Config.Xmx, list[idx].Config.Xms
	}
	return 0, 0
}
