// 本文件实现服务器 Java 路径配置的读写
package storage

import (
	"MCServer/backend/model"
)

// SetServerJava 设置指定服务器使用的 java.exe 路径
// 每个服务器可配置不同的 Java，持久化到 ServerList.json
func (s *Storage) SetServerJava(serverName, javaPath string) error {
	return s.UpdateServerList(func(list []model.ServerConfig) ([]model.ServerConfig, error) {
		// 按服务器名查找，若存在则更新其 Java 路径
		idx, found := findServerByName(list, serverName)
		if found {
			list[idx].Config.JavaPath = javaPath
		} else {
			// 若服务器不在列表中，追加一个新条目
			list = append(list, model.ServerConfig{
				ServerName: serverName,
				Config: model.ServerConfigExtra{
					IsActive: false,
					JavaPath: javaPath,
				},
			})
		}
		return list, nil
	})
}

// GetServerJava 获取指定服务器配置的 java.exe 路径
// 未配置时返回空字符串
func (s *Storage) GetServerJava(serverName string) string {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return ""
	}
	// 按服务器名查找，返回其 Java 路径
	idx, found := findServerByName(list, serverName)
	if found {
		return list[idx].Config.JavaPath
	}
	return ""
}
