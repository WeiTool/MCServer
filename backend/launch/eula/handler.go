// 本文件实现 eula.txt 缺失时的"关服→改true→自动重启"流程
package eula

import (
	"fmt"
	"time"
)

// WatchEulaAndRestart 等待 eula.txt 出现，处理"关服→改true→自动重启"流程
// 仅在启动时 eula.txt 缺失的情况下由进程管理器调用
// 参数 log 推送日志，stop 关停当前进程并清空状态，restart 自动重启
func WatchEulaAndRestart(serverName string, log func(string), stop func(), restart func() error) {
	log("> 未检测到 eula.txt，等待服务端生成...")

	// 持续轮询 eula.txt，最多等待 5 分钟
	appeared := WaitEulaAppear(serverName, 2*time.Second, 5*time.Minute)
	if !appeared {
		log("[警告] 等待 eula.txt 超时，请手动处理")
		return
	}

	log("> 检测到 eula.txt 出现，2 秒后自动重启以同意 EULA...")

	// 等待 2 秒
	time.Sleep(2 * time.Second)

	// 关服（忽略错误，因为服务端可能已因 eula 未同意而自行退出）
	stop()

	// 修改 eula.txt 为 eula=true
	if err := SetEulaAgreed(serverName); err != nil {
		log(fmt.Sprintf("[错误] 修改 eula.txt 失败: %v", err))
		return
	}
	log("> 已同意 EULA，自动重新启动服务器...")

	// 自动重启（继续类型检测）
	if err := restart(); err != nil {
		log(fmt.Sprintf("[错误] 自动重启失败: %v", err))
	}
}
