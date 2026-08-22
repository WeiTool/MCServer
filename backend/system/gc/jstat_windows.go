//go:build windows

package gc

// jstatExecutable Windows 平台 jstat 可执行文件名
func jstatExecutable() string {
	return "jstat.exe"
}
