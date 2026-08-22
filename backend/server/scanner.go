// Package server 提供服务器目录的扫描与信息查询能力
// 本文件实现服务器文件夹列表的扫描与持久化
package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"MCServer/backend/model"
	"MCServer/backend/storage"
	"MCServer/backend/utils"
)

// Scanner 服务器扫描器
// 扫描 servers 根目录下的服务器文件夹列表
type Scanner struct{}

// NewScanner 创建扫描器实例
func NewScanner() *Scanner { return &Scanner{} }

// ScanServers 扫描 exe 目录下的 servers 文件夹
// 校验 servers 主文件夹名必须是全小写的 "servers"，否则不检测
// 返回服务器实例列表
func (s *Scanner) ScanServers() (*model.ServerListResult, error) {
	// 1. 获取 exe 所在目录
	exeDir := utils.GetExeDir()
	if exeDir == "" {
		return nil, fmt.Errorf("获取 exe 目录失败")
	}

	// 2. 校验主文件夹名：必须全小写且为 "servers"
	//    若 exe 目录下存在非全小写的同名文件夹（如 Servers、SERVER），则不检测
	if !s.isValidServersDir(exeDir) {
		return &model.ServerListResult{
			Servers:  []model.ServerInstance{},
			BasePath: utils.GetServersRoot(),
			Total:    0,
		}, nil
	}

	// 3. 构建 servers 目录路径
	serversRootPath := utils.GetServersRoot()

	// 4. 检查目录是否存在
	if _, err := os.Stat(serversRootPath); os.IsNotExist(err) {
		return &model.ServerListResult{
			Servers:  []model.ServerInstance{},
			BasePath: serversRootPath,
			Total:    0,
		}, nil
	}

	// 5. 读取目录下的所有子目录
	entries, err := os.ReadDir(serversRootPath)
	if err != nil {
		return nil, err
	}

	var servers []model.ServerInstance
	var serverConfigs []model.ServerConfig

	for _, entry := range entries {
		if !entry.IsDir() {
			continue // 跳过文件，只处理文件夹
		}

		serverName := entry.Name()

		serverFolderPath := filepath.Join(serversRootPath, serverName)

		instance := s.scanServerFolder(serverName, serverFolderPath)
		if instance.HasJar {
			servers = append(servers, instance)
			// 统计该服务器的 mod/插件数量
			serverConfigs = append(serverConfigs, model.ServerConfig{
				ServerName: serverName,
				Info: model.ServerInfo{
					ModCount:    ScanMods(serverFolderPath),
					PluginCount: ScanPlugins(serverFolderPath),
				},
			})
		}
	}

	// 保留活动服务器标记并持久化到 config/ServerList.json
	s.persistServerConfigs(serverConfigs)

	return &model.ServerListResult{
		Servers:  servers,
		BasePath: serversRootPath,
		Total:    len(servers),
	}, nil
}

// persistServerConfigs 将服务器配置写入 config/ServerList.json
// 保留原有的用户配置（Java 路径、内存、活动标记等），避免每次扫描覆盖丢失
// 在 ConfigService 锁内整体读-改-写，防止与其他写者（检测/内存配置）并发丢数据
func (s *Scanner) persistServerConfigs(configs []model.ServerConfig) {
	cfg := storage.NewStorage()
	_ = cfg.UpdateServerList(func(existing []model.ServerConfig) ([]model.ServerConfig, error) {
		// 建立现有服务器配置映射（按服务器名）
		existingMap := make(map[string]model.ServerConfig)
		for _, e := range existing {
			existingMap[e.ServerName] = e
		}

		// 将现有配置合并到新配置
		// - 统计字段（Info: version/type/modCount/pluginCount）优先用新扫描值，为空时回退旧值避免丢数据
		// - 运行时配置字段（Config: 活动标记/Java/内存）保留旧值，避免被覆盖
		hasActive := false
		for i := range configs {
			if old, ok := existingMap[configs[i].ServerName]; ok {
				if configs[i].Info.Version == "" {
					configs[i].Info.Version = old.Info.Version
				}
				if configs[i].Info.Type == "" {
					configs[i].Info.Type = old.Info.Type
				}
				if configs[i].Info.ModCount == 0 {
					configs[i].Info.ModCount = old.Info.ModCount
				}
				if configs[i].Info.PluginCount == 0 {
					configs[i].Info.PluginCount = old.Info.PluginCount
				}
				configs[i].Config.IsActive = old.Config.IsActive
				configs[i].Config.JavaPath = old.Config.JavaPath
				configs[i].Config.Xmx = old.Config.Xmx
				configs[i].Config.Xms = old.Config.Xms
			}
			if configs[i].Config.IsActive {
				hasActive = true
			}
		}

		// 若没有任何活动服务器，默认将第一个设为活动
		if !hasActive && len(configs) > 0 {
			configs[0].Config.IsActive = true
		}

		return configs, nil
	})
}

// isValidServersDir 校验 exe 目录下的 servers 主文件夹名是否合法
// 规则：文件夹名必须为全小写的 "servers"
// 若存在同名但非全小写的文件夹（如 Servers、SERVER），返回 false 不检测
func (s *Scanner) isValidServersDir(exeDir string) bool {
	entries, err := os.ReadDir(exeDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 找到名为 servers 的文件夹，校验是否全小写
		if strings.EqualFold(name, "servers") {
			return utils.IsAllLowercase(name)
		}
	}
	return false
}

// scanServerFolder 扫描单个服务器文件夹
// 参数 serverName 为服务器名（文件夹名），serverFolderPath 为服务器文件夹绝对路径
// 返回包含 jar 文件、eula 状态的服务器实例信息
func (s *Scanner) scanServerFolder(serverName, serverFolderPath string) model.ServerInstance {
	// 初始化服务器实例，记录服务器名与路径
	instance := model.ServerInstance{
		Name: serverName,
		Path: serverFolderPath,
	}

	// 读取服务器文件夹下的所有条目
	entries, err := os.ReadDir(serverFolderPath)
	if err != nil {
		// 无法读取目录，返回空结果
		return instance
	}

	// 收集所有 .jar 文件的文件名
	var jarFiles []string
	for _, entry := range entries {
		// 跳过子目录，只处理文件
		if entry.IsDir() {
			continue
		}
		// 统一用 utils.IsJarFile 判断是否为 jar 文件
		if utils.IsJarFile(entry.Name()) {
			jarFiles = append(jarFiles, entry.Name())
		}
	}

	// 只要有 .jar 文件就算有效服务器，记录 jar 相关信息
	if len(jarFiles) > 0 {
		instance.HasJar = true
		instance.JarCount = len(jarFiles)
		instance.JarFiles = jarFiles
	}

	// 检查 eula.txt 是否存在及是否已同意
	eulaPath := filepath.Join(serverFolderPath, "eula.txt")
	if _, err := os.Stat(eulaPath); err == nil {
		instance.HasEula = true
		instance.EulaAgreed = s.checkEulaAgreed(eulaPath)
	}

	return instance
}

// checkEulaAgreed 检查 eula.txt 是否已同意
func (s *Scanner) checkEulaAgreed(eulaPath string) bool {
	content, err := os.ReadFile(eulaPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "eula=true")
}

// GetServerPath 获取 servers 目录的绝对路径
func (s *Scanner) GetServerPath() (string, error) {
	// 统一由 utils 提供 servers 根目录
	root := utils.GetServersRoot()
	if root == "" {
		return "", fmt.Errorf("获取 servers 目录失败")
	}
	return root, nil
}
