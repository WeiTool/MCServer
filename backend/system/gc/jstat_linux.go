//go:build linux

package gc

// jstatExecutable Linux 平台 jstat 可执行文件名
func jstatExecutable() string {
	return "jstat"
}
