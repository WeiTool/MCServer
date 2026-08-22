package cpubind

// BindToPerformanceCores 将指定 PID 进程绑定到所有性能核心（P-Core）
// 说明：Minecraft 服务端是单线程主循环（仅 Folia 为多线程），
// 绑定到全 P 核后，OS 调度器可在任意 P 核之间迁移主线程，避免落到 E 核上；
// AMD 全核 / Intel Xeon 等无 E 核的 CPU 则等效"不限制"。
// 具体实现由各平台文件提供
func BindToPerformanceCores(pid int) error {
	return bindToPerformanceCores(pid)
}
