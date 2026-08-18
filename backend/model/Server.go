package model

// ServerInstance 服务器实例信息
type ServerInstance struct {
	Name       string   `json:"name"`       // 服务器名称（文件夹名）
	Path       string   `json:"path"`       // 绝对路径
	HasJar     bool     `json:"hasJar"`     // 是否有 .jar 文件
	JarCount   int      `json:"jarCount"`   // .jar 文件数量（用于显示）
	JarFiles   []string `json:"jarFiles"`   // 所有 .jar 文件列表
	HasEula    bool     `json:"hasEula"`    // 是否有 eula.txt
	EulaAgreed bool     `json:"eulaAgreed"` // eula.txt 是否已同意
	IsRunning  bool     `json:"isRunning"`  // 是否正在运行
	PID        int      `json:"pid"`        // 进程 PID
}

// ServerListResult 服务器列表查询结果
type ServerListResult struct {
	Servers  []ServerInstance `json:"servers"`
	BasePath string           `json:"basePath"`
	Total    int              `json:"total"`
}

// ServerConfig 服务器配置（持久化到 config/ServerList.json）
// serverName 为服务器文件夹名称
type ServerConfig struct {
	// 服务器名称（文件夹名）
	ServerName string `json:"serverName"`
	// 服务器统计信息（版本/类型/mod/插件）
	Info ServerInfo `json:"info"`
	// 服务器运行时配置（活动标记/Java/内存）
	Config ServerConfigExtra `json:"config"`
}

// ServerInfo 服务器统计信息
type ServerInfo struct {
	// 服务器版本
	Version string `json:"version"`
	// 服务器类型（Forge/Fabric 等）
	Type string `json:"type"`
	// 模组数量
	ModCount string `json:"modCount"`
	// 插件数量
	PluginCount string `json:"pluginCount"`
}

// ServerConfigExtra 服务器运行时配置
type ServerConfigExtra struct {
	// 是否为当前活动服务器
	IsActive bool `json:"isActive"`
	// 该服务器使用的 java.exe 路径
	JavaPath string `json:"javaPath"`
	// 最大内存（MB）
	Xmx int `json:"xmx"`
	// 最小内存（MB）
	Xms int `json:"xms"`
}
