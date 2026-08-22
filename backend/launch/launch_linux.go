//go:build linux

package launch

// InitJobObject Linux 平台占位实现
// Linux 下无 Windows Job Object 概念，子进程随父进程退出由系统管理
func InitJobObject() error {
	return nil
}

// AssignProcessToJob Linux 平台占位实现
// 子进程天然受父进程生命周期管理，无需额外处理
func AssignProcessToJob(pid int) error {
	return nil
}
