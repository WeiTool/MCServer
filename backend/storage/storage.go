// Package storage 提供应用配置的持久化能力
// ServerList.json 存于 exe 同级 config 文件夹，不存在会自动创建
// 本文件实现存储基础设施与活动服务器配置
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"MCServer/backend/model"
	"MCServer/backend/utils"
)

// configFileName 服务器列表配置文件（JSON 数组）
const configFileName = "ServerList.json"

// mu 串行化所有对 ServerList.json 的读-改-写
// 多个服务（进程检测、列表扫描、配置读写）会并发读写该文件，
// 不加锁时并发读-改-写会导致丢失更新、字段错位（如类型写入错误条目）
var mu sync.Mutex

// Storage 应用配置存储服务
// 负责读写 exe 同级 config 文件夹中的 JSON 配置文件
type Storage struct {
	// config 文件夹的绝对路径
	dir string
}

// NewStorage 创建存储服务实例
// 初始化时会确保 config 文件夹存在
func NewStorage() *Storage {
	s := &Storage{
		dir: utils.GetConfigDir(),
	}
	// 确保 config 文件夹存在
	_ = os.MkdirAll(s.dir, 0755)
	return s
}

// getConfigPath 获取 ServerList.json 的完整路径
func (s *Storage) getConfigPath() string {
	return filepath.Join(s.dir, configFileName)
}

// loadServerList 读取服务器列表配置（不加锁，调用方需持有 mu）
// 文件不存在返回空数组；解析失败返回错误，避免调用方用空列表覆盖原数据
func (s *Storage) loadServerList() ([]model.ServerConfig, error) {
	data, err := os.ReadFile(s.getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []model.ServerConfig{}, nil
		}
		return nil, err
	}

	var list []model.ServerConfig
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("解析 ServerList.json 失败: %v", err)
	}
	return list, nil
}

// saveServerList 保存服务器列表配置（不加锁，调用方需持有 mu）
func (s *Storage) saveServerList(list []model.ServerConfig) error {
	// 确保 config 文件夹存在（防止被手动删除）
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.getConfigPath(), data, 0644)
}

// LoadServerList 读取服务器列表配置
func (s *Storage) LoadServerList() ([]model.ServerConfig, error) {
	mu.Lock()
	defer mu.Unlock()
	return s.loadServerList()
}

// SaveServerList 保存服务器列表配置到 ServerList.json
func (s *Storage) SaveServerList(list []model.ServerConfig) error {
	mu.Lock()
	defer mu.Unlock()
	return s.saveServerList(list)
}

// UpdateServerList 在锁内执行"读取-修改-写回"
// 保证并发下读-改-写整体原子，防止丢失更新（apply 返回修改后的完整列表）
func (s *Storage) UpdateServerList(apply func(list []model.ServerConfig) ([]model.ServerConfig, error)) error {
	mu.Lock()
	defer mu.Unlock()

	list, err := s.loadServerList()
	if err != nil {
		return err
	}
	updated, err := apply(list)
	if err != nil {
		return err
	}
	return s.saveServerList(updated)
}

// findServerByName 在服务器列表中找到指定服务器名的下标
// 返回下标索引与是否找到；未找到时返回 -1, false
func findServerByName(list []model.ServerConfig, serverName string) (int, bool) {
	// 遍历列表，逐个比对服务器名
	for i := range list {
		if list[i].ServerName == serverName {
			// 找到匹配的服务器，返回其下标
			return i, true
		}
	}
	// 未找到，返回 -1
	return -1, false
}

// SetActiveServer 设置当前活动服务器
// 将该服务器的 isActive 置为 true，其余置为 false，并持久化
func (s *Storage) SetActiveServer(serverName string) error {
	return s.UpdateServerList(func(list []model.ServerConfig) ([]model.ServerConfig, error) {
		// 先判断该服务器是否已存在于列表
		_, found := findServerByName(list, serverName)
		// 遍历列表：匹配的服务器置为活动，其余置为非活动
		for i := range list {
			list[i].Config.IsActive = (list[i].ServerName == serverName)
		}

		// 若服务器不在列表中，追加一个新活动条目
		if !found {
			list = append(list, model.ServerConfig{
				ServerName: serverName,
				Config: model.ServerConfigExtra{
					IsActive: true,
				},
			})
		}
		return list, nil
	})
}

// GetActiveServer 获取当前活动服务器名称
// 遍历配置列表，返回 isActive 为 true 的服务器名
// 若没有任何服务器标记为活动，则将第一个服务器设为默认活动并持久化
func (s *Storage) GetActiveServer() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	list, err := s.loadServerList()
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", nil
	}

	// 查找活动服务器
	for _, item := range list {
		if item.Config.IsActive {
			return item.ServerName, nil
		}
	}

	// 没有活动服务器，默认将第一个设为活动并写回 json
	list[0].Config.IsActive = true
	if err := s.saveServerList(list); err != nil {
		return "", err
	}
	return list[0].ServerName, nil
}

// GetServerPort 获取指定服务器的 server-port 端口
// 从 server.properties 中读取端口配置
func (s *Storage) GetServerPort(serverName string) (string, error) {
	props := NewServerProperties()
	return props.GetPort(serverName)
}

func (s *Storage) GetQueryPort(serverName string) (string, error) {
	props := NewServerProperties()
	return props.GetQueryPort(serverName)
}
