package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mcstatus-io/mcutil/v4/query"
)

// GetFullPlayerInfo 使用 query.Full 获取在线玩家数量、最大玩家数量和完整玩家列表
// 参数:
//   - port: 服务器查询端口（注意：不是游戏端口，是 server.properties 中 query.port 设置的端口）
//
// 返回值:
//   - online: 当前在线玩家数量
//   - max: 最大玩家数量
//   - players: 完整的在线玩家昵称列表
//   - err: 错误信息
func GetFullPlayerInfo(port uint16) (online int, max int, players []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行完整查询，地址固定为 127.0.0.1
	resp, err := query.Full(ctx, "127.0.0.1", port)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("query.Full 查询失败: %w", err)
	}

	// 从 Data map 中提取在线玩家数量
	if onlineStr, ok := resp.Data["numplayers"]; ok {
		if val, err := strconv.Atoi(onlineStr); err == nil {
			online = val
		}
	}

	// 从 Data map 中提取最大玩家数量
	if maxStr, ok := resp.Data["maxplayers"]; ok {
		if val, err := strconv.Atoi(maxStr); err == nil {
			max = val
		}
	}

	// 获取完整玩家列表
	players = resp.Players

	return online, max, players, nil
}
