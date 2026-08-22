//go:build linux

package linux

import (
	"os"
)

// readFile 读取文件内容，供 sysfs 和 /proc 读取使用
func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
