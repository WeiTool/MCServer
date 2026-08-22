// 本文件实现服务器统计信息配置的读写（server_info_config）
// 包括 mod/插件数量、类型、版本
package storage

import (
	"MCServer/backend/model"
)

// setServerInfoField 更新指定服务器的 Info 字段值并写回 ServerList.json
// 参数 apply 为具体的字段赋值回调（更新 Type 或 Version 等）
// 服务器不存在时追加一个新条目（仅含该字段）
func (s *Storage) setServerInfoField(serverName, value string, apply func(*model.ServerInfo, string)) error {
	return s.UpdateServerList(func(list []model.ServerConfig) ([]model.ServerConfig, error) {
		// 按服务器名查找
		idx, found := findServerByName(list, serverName)
		if found {
			// 更新该服务器的 Info 字段
			apply(&list[idx].Info, value)
		} else {
			// 服务器不存在，追加一个新条目
			newInfo := model.ServerInfo{}
			apply(&newInfo, value)
			list = append(list, model.ServerConfig{
				ServerName: serverName,
				Info:       newInfo,
			})
		}
		return list, nil
	})
}

// SetServerType 设置指定服务器的类型（info.type）并写回 ServerList.json
// 类型如：Vanilla / Fabric / Forge / Paper / Mohist / NeoForge / Youer / Purpur
func (s *Storage) SetServerType(serverName, serverType string) error {
	// 复用通用字段写入逻辑，更新 Info.Type
	return s.setServerInfoField(serverName, serverType, func(info *model.ServerInfo, v string) {
		info.Type = v
	})
}

// GetServerType 读取指定服务器的类型（info.type）
// 未配置时返回空字符串
func (s *Storage) GetServerType(serverName string) string {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return ""
	}
	// 按服务器名查找
	idx, found := findServerByName(list, serverName)
	if found {
		return list[idx].Info.Type
	}
	return ""
}

// SetServerVersion 记录版本信息
func (s *Storage) SetServerVersion(serverName, version string) error {
	// 复用通用字段写入逻辑，更新 Info.Version
	return s.setServerInfoField(serverName, version, func(info *model.ServerInfo, v string) {
		info.Version = v
	})
}

// GetServerVersion 读取指定服务器的版本（info.version）
// 未配置时返回空字符串
func (s *Storage) GetServerVersion(serverName string) string {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return ""
	}
	// 按服务器名查找
	idx, found := findServerByName(list, serverName)
	if found {
		return list[idx].Info.Version
	}
	return ""
}

// SetServerExtensionsCount 同时更新指定服务器的 mod 与插件数量并写回 ServerList.json
// 若服务器不在列表中，追加一个新条目（仅含 mod/pluginCount）
func (s *Storage) SetServerExtensionsCount(serverName string, modCount, pluginCount int) error {
	return s.UpdateServerList(func(list []model.ServerConfig) ([]model.ServerConfig, error) {
		// 按服务器名查找，若存在则更新其 mod 数量
		idx, found := findServerByName(list, serverName)
		if found {
			list[idx].Info.ModCount = modCount
			list[idx].Info.PluginCount = pluginCount
		} else {
			// 若服务器不在列表中，追加一个新条目
			list = append(list, model.ServerConfig{
				ServerName: serverName,
				Info: model.ServerInfo{
					ModCount:    modCount,
					PluginCount: pluginCount,
				},
			})
		}
		return list, nil
	})
}

// GetServerModCount 读取指定服务器配置的 mod 数量（info.modCount）
// 未配置时返回 0
func (s *Storage) GetServerModCount(serverName string) int {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return 0
	}
	// 按服务器名查找
	idx, found := findServerByName(list, serverName)
	if found {
		return list[idx].Info.ModCount
	}
	return 0
}

// GetServerPluginCount 读取指定服务器配置的插件数量（info.pluginCount）
// 未配置时返回 0
func (s *Storage) GetServerPluginCount(serverName string) int {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return 0
	}
	// 按服务器名查找
	idx, found := findServerByName(list, serverName)
	if found {
		return list[idx].Info.PluginCount
	}
	return 0
}
