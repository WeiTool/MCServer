// Package config 提供应用配置的持久化能力
// 配置存储在 exe 同级的 config 文件夹中，不存在会自动创建
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"MCServer/backend/model"
	"MCServer/backend/utils"
)

// configFileName 服务器列表配置文件
// 存储所有服务器的配置信息（JSON 数组）
const configFileName = "ServerList.json"

// ConfigService 配置服务
// 负责读写 exe 同级 config 文件夹中的配置文件
type ConfigService struct {
	// config 文件夹的绝对路径
	dir string
}

// NewConfigService 创建配置服务实例
// 初始化时会确保 config 文件夹存在
func NewConfigService() *ConfigService {
	s := &ConfigService{}
	s.dir = s.getConfigDir()
	// 确保 config 文件夹存在
	_ = os.MkdirAll(s.dir, 0755)
	return s
}

// getConfigDir 获取 config 文件夹路径（exe 同级）
// 统一由 utils.GetConfigDir 提供，避免各处重复拼接
func (s *ConfigService) getConfigDir() string {
	return utils.GetConfigDir()
}

// getConfigPath 获取 ServerList.json 的完整路径
func (s *ConfigService) getConfigPath() string {
	return filepath.Join(s.dir, configFileName)
}

// LoadServerList 读取服务器列表配置
// 若文件不存在或解析失败，返回空数组
func (s *ConfigService) LoadServerList() ([]model.ServerConfig, error) {
	data, err := os.ReadFile(s.getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []model.ServerConfig{}, nil
		}
		return nil, err
	}

	var list []model.ServerConfig
	if err := json.Unmarshal(data, &list); err != nil {
		// 解析失败返回空数组，避免影响主流程
		return []model.ServerConfig{}, nil
	}
	return list, nil
}

// SaveServerList 保存服务器列表配置到 ServerList.json
func (s *ConfigService) SaveServerList(list []model.ServerConfig) error {
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

// SetServerJava 设置指定服务器使用的 java.exe 路径
// 每个服务器可配置不同的 Java，持久化到 ServerList.json
func (s *ConfigService) SetServerJava(serverName, javaPath string) error {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return err
	}

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

	return s.SaveServerList(list)
}

// GetServerJava 获取指定服务器配置的 java.exe 路径
// 未配置时返回空字符串
func (s *ConfigService) GetServerJava(serverName string) string {
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

// SetServerExtensionsCount 同时更新指定服务器的 mod 与插件数量并写回 ServerList.json
// 若服务器不在列表中，追加一个新条目（仅含 mod/pluginCount）
func (s *ConfigService) SetServerExtensionsCount(serverName string, modCount, pluginCount int) error {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return err
	}

	// 按服务器名查找，若存在则更新其 mod 数量
	idx, found := findServerByName(list, serverName)
	if found {
		list[idx].Info.ModCount = strconv.Itoa(modCount)
		list[idx].Info.PluginCount = strconv.Itoa(pluginCount)
	} else {
		// 若服务器不在列表中，追加一个新条目
		list = append(list, model.ServerConfig{
			ServerName: serverName,
			Info: model.ServerInfo{
				ModCount:    strconv.Itoa(modCount),
				PluginCount: strconv.Itoa(pluginCount),
			},
		})
	}

	return s.SaveServerList(list)
}

// GetServerModCount 读取指定服务器配置的 mod 数量（info.modCount）
// 未配置时返回 0
func (s *ConfigService) GetServerModCount(serverName string) int {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return 0
	}
	// 按服务器名查找
	idx, found := findServerByName(list, serverName)
	if found {
		// 解析 mod 数量
		count, _ := strconv.Atoi(list[idx].Info.ModCount)
		return count
	}
	return 0
}

// GetServerPluginCount 读取指定服务器配置的插件数量（info.pluginCount）
// 未配置时返回 0
func (s *ConfigService) GetServerPluginCount(serverName string) int {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return 0
	}
	// 按服务器名查找
	idx, found := findServerByName(list, serverName)
	if found {
		// 解析插件数量
		count, _ := strconv.Atoi(list[idx].Info.PluginCount)
		return count
	}
	return 0
}

// setServerInfoField 更新指定服务器的 Info 字段值并写回 ServerList.json
// 参数 apply 为具体的字段赋值回调（更新 Type 或 Version 等）
// 服务器不存在时追加一个新条目（仅含该字段）
func (s *ConfigService) setServerInfoField(serverName, value string, apply func(*model.ServerInfo, string)) error {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return err
	}

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

	return s.SaveServerList(list)
}

// SetServerType 设置指定服务器的类型（info.type）并写回 ServerList.json
// 类型如：Vanilla / Fabric / Forge / Paper / Mohist / NeoForge / Youer / Purpur
func (s *ConfigService) SetServerType(serverName, serverType string) error {
	// 复用通用字段写入逻辑，更新 Info.Type
	return s.setServerInfoField(serverName, serverType, func(info *model.ServerInfo, v string) {
		info.Type = v
	})
}

// SetServerVersion 设置指定服务器的版本（info.version）并写回 ServerList.json
func (s *ConfigService) SetServerVersion(serverName, version string) error {
	// 复用通用字段写入逻辑，更新 Info.Version
	return s.setServerInfoField(serverName, version, func(info *model.ServerInfo, v string) {
		info.Version = v
	})
}

// GetServerType 读取指定服务器的类型（info.type）
// 未配置时返回空字符串
func (s *ConfigService) GetServerType(serverName string) string {
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

// GetServerVersion 读取指定服务器的版本（info.version）
// 未配置时返回空字符串
func (s *ConfigService) GetServerVersion(serverName string) string {
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

// SetActiveServer 设置当前活动服务器
// 将该服务器的 isActive 置为 true，其余置为 false，并持久化
func (s *ConfigService) SetActiveServer(serverName string) error {
	// 读取现有服务器列表
	list, err := s.LoadServerList()
	if err != nil {
		return err
	}

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

	return s.SaveServerList(list)
}

// GetActiveServer 获取当前活动服务器名称
// 遍历配置列表，返回 isActive 为 true 的服务器名
// 若没有任何服务器标记为活动，则将第一个服务器设为默认活动并持久化
func (s *ConfigService) GetActiveServer() (string, error) {
	list, err := s.LoadServerList()
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
	if err := s.SaveServerList(list); err != nil {
		return "", err
	}
	return list[0].ServerName, nil
}
