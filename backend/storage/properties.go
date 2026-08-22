// Package storage 提供应用配置与服务器配置的持久化与读取能力
// 本文件实现 server.properties 的解析（非 JSON，但同属配置读写）
package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magiconair/properties"

	"MCServer/backend/utils"
)

// serverPropertiesFileName Minecraft 服务端配置文件名
const serverPropertiesFileName = "server.properties"

// ServerProperties server.properties 文件解析与编辑服务
type ServerProperties struct {
	// 缓存已加载的 properties 对象，避免重复读取
	cache map[string]*properties.Properties
}

// NewServerProperties 创建 ServerProperties 服务实例
func NewServerProperties() *ServerProperties {
	return &ServerProperties{
		cache: make(map[string]*properties.Properties),
	}
}

// getFilePath 获取指定服务器的 server.properties 完整路径
func (s *ServerProperties) getFilePath(serverName string) string {
	return filepath.Join(utils.GetServerFolderPath(serverName), serverPropertiesFileName)
}

// loadProperties 加载指定服务器的 properties 对象（带缓存）
func (s *ServerProperties) loadProperties(serverName string) (*properties.Properties, error) {
	// 检查缓存
	if p, ok := s.cache[serverName]; ok {
		return p, nil
	}

	path := s.getFilePath(serverName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// 文件不存在：返回空 properties，不报错
			return properties.NewProperties(), nil
		}
		return nil, fmt.Errorf("读取 server.properties 失败: %v", err)
	}

	p, err := properties.LoadFile(path, properties.UTF8)
	if err != nil {
		return nil, fmt.Errorf("解析 server.properties 失败: %v", err)
	}

	s.cache[serverName] = p
	return p, nil
}

// Load 读取指定服务器的 server.properties 全部字段
func (s *ServerProperties) Load(serverName string) (map[string]string, error) {
	p, err := s.loadProperties(serverName)
	if err != nil {
		return nil, err
	}
	return p.Map(), nil
}

// Save 将指定服务器的配置保存到文件
func (s *ServerProperties) Save(serverName string) error {
	p, ok := s.cache[serverName]
	if !ok {
		var err error
		p, err = s.loadProperties(serverName)
		if err != nil {
			return err
		}
	}

	path := s.getFilePath(serverName)
	// 创建或覆盖文件
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer f.Close()

	// 使用 Write 方法将配置写入文件，指定 UTF-8 编码
	_, err = p.Write(f, properties.UTF8)
	return err
}

// Update 更新指定服务器的配置项并立即保存到文件 (新增在 Save 下面)
func (s *ServerProperties) Update(serverName string, updates map[string]string) error {
	// 1. 获取（或加载）缓存中的 properties 对象
	p, err := s.loadProperties(serverName)
	if err != nil {
		return err
	}

	// 2. 遍历前端传来的更新数据，写入到 properties 对象中
	for key, value := range updates {
		p.Set(key, value)
	}

	// 3. 将修改后的内容保存到磁盘
	return s.Save(serverName)
}

func (s *ServerProperties) GetPort(serverName string) (string, error) {
	p, err := s.loadProperties(serverName)
	if err != nil {
		return "", err
	}

	// 获取 server-port 字段的值
	port := p.GetString("server-port", "")
	if port == "" {
		return "", fmt.Errorf("server-port 未配置")
	}

	return port, nil
}

func (s *ServerProperties) GetQueryPort(serverName string) (string, error) {
	p, err := s.loadProperties(serverName)
	if err != nil {
		return "", err
	}

	queryPort := p.GetString("query.port", "")
	if queryPort == "" {
		return "", fmt.Errorf("query.port 未配置")
	}
	return queryPort, nil
}

// ClearCache 清除缓存（用于服务器配置更新后强制重新加载）
func (s *ServerProperties) ClearCache(serverName string) {
	delete(s.cache, serverName)
}
