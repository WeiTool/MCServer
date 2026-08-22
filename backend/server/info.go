// 本文件实现服务器信息查询：类型、版本、运行时长
package server

import (
	"MCServer/backend/storage"
	"MCServer/backend/utils"
	"fmt"
	"strconv"
)

// uptimeProvider 提供指定服务器的已运行秒数（由进程服务实现）
type uptimeProvider interface {
	GetServerUptimeFor(serverName string) int
}

// ServerInfo 服务器信息查询服务
// 类型/版本读取启动检测写入的 ServerList.json；运行时长委托给进程服务
type ServerInfo struct {
	cfg  *storage.Storage
	proc uptimeProvider
}

// NewServerInfo 创建信息查询服务实例
// cfg 为共享存储实例；proc 为运行时长提供者（进程服务），无进程上下文时传 nil
func NewServerInfo(cfg *storage.Storage, proc uptimeProvider) *ServerInfo {
	return &ServerInfo{
		cfg:  cfg,
		proc: proc,
	}
}

// GetType 获取指定服务器的类型（info.type，由启动日志检测写入）
// 未检测到时返回空字符串
func (i *ServerInfo) GetType(serverName string) string {
	return i.cfg.GetServerType(serverName)
}

// GetVersion 获取指定服务器的版本（info.version）
// 未检测到时返回空字符串
func (i *ServerInfo) GetVersion(serverName string) string {
	return i.cfg.GetServerVersion(serverName)
}

// GetUptime 获取指定服务器的已运行秒数
// 仅当该服务器当前正在运行时返回实际时长，否则返回 0
func (i *ServerInfo) GetUptime(serverName string) int {
	if i.proc == nil {
		return 0
	}
	return i.proc.GetServerUptimeFor(serverName)
}

// GetOnlinePlayers 获取当前活动服务器的在线玩家数量
func (i *ServerInfo) GetOnlinePlayers() (int, error) {
	serverName, err := i.cfg.GetActiveServer()
	if err != nil {
		return 0, err
	}
	if serverName == "" {
		return 0, nil
	}

	queryPortStr, err := i.cfg.GetQueryPort(serverName)
	if err != nil {
		return 0, err
	}
	if queryPortStr == "" {
		return 0, nil
	}

	port, err := strconv.ParseUint(queryPortStr, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("无效的端口号: %w", err)
	}

	online, _, _, err := utils.GetFullPlayerInfo(uint16(port))
	if err != nil {
		return 0, err
	}

	return online, nil
}

// GetMaxPlayers 获取当前活动服务器的最大玩家数量
func (i *ServerInfo) GetMaxPlayers() (int, error) {
	serverName, err := i.cfg.GetActiveServer()
	if err != nil {
		return 0, err
	}
	if serverName == "" {
		return 0, nil
	}

	queryPortStr, err := i.cfg.GetQueryPort(serverName)
	if err != nil {
		return 0, err
	}
	if queryPortStr == "" {
		return 0, nil
	}

	port, err := strconv.ParseUint(queryPortStr, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("无效的端口号: %w", err)
	}

	_, max, _, err := utils.GetFullPlayerInfo(uint16(port))
	if err != nil {
		return 0, err
	}

	return max, nil
}

// GetPlayerList 获取当前活动服务器的完整玩家列表
func (i *ServerInfo) GetPlayerList() ([]string, error) {
	serverName, err := i.cfg.GetActiveServer()
	if err != nil {
		return nil, err
	}
	if serverName == "" {
		return []string{}, nil
	}

	queryPortStr, err := i.cfg.GetQueryPort(serverName)
	if err != nil {
		return nil, err
	}
	if queryPortStr == "" {
		return []string{}, nil
	}

	port, err := strconv.ParseUint(queryPortStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("无效的端口号: %w", err)
	}

	_, _, players, err := utils.GetFullPlayerInfo(uint16(port))
	if err != nil {
		return nil, err
	}

	return players, nil
}
