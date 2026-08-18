package model

// JavaInfo Java 环境信息
type JavaInfo struct {
	// Java 安装目录（bin 的上一级文件夹）
	Path string `json:"path"`
	// Java 可执行文件完整路径
	Executable string `json:"executable"`
	// Java 主版本号（如 8、17、21）
	Version int `json:"version"`
	// 完整版本号字符串（如 1.8.0_291、17.0.1）
	VersionName string `json:"versionName"`
}
