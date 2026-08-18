// Package launch 提供 Java 服务端进程的启动能力
// 负责启动 Java 进程，并将其绑定到负载最低的性能核心
package launch

// BindToLowestLoadCore 将指定 PID 进程绑定到负载最低的性能核心
// 流程: 识别 P-Core -> 找负载最低核心 -> 设置亲和性
// 具体实现由各平台文件提供
func BindToLowestLoadCore(pid int) error {
	return bindToLowestLoadCore(pid)
}
