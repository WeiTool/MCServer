package model

type Release struct {
	Assets []Asset `json:"assets"`
}

// Asset 嵌套结构体
type Asset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	Name               string `json:"name"`
}

// VersionResponse 版本检查结果（供前端提示是否有新版本）
type VersionResponse struct {
	// 是否有新版本（最新 release 版本号与当前版本不一致）
	HasUpdate bool `json:"hasUpdate"`
	// 当前运行版本号（如 0.0.1）
	Current string `json:"current"`
	// 最新 release 版本号（如 0.0.1）
	Latest string `json:"latest"`
	// 匹配当前平台的下载文件完整名称（如 MCServer-v0.0.1-beta1-windows-amd64.exe）
	AssetName string `json:"assetName"`
	// 下载地址
	DownloadURL string `json:"downloadUrl"`
	// 是否为 beta 预发布版本
	IsBeta bool `json:"isBeta"`
	// 是否为预览版（按当前全局配置是否启用预览版决定）
	IsPreview bool `json:"isPreview"`
}

// GlobalConfig 全局配置（config/global_config.json）
type GlobalConfig struct {
	// 是否获取预览版（beta）更新
	PreviewEnabled bool `json:"previewEnabled"`
}

// UpdateState 静默更新状态（config/update.json）
// 下载完成后写入 pending；替换脚本执行后改为 updated 或 error；
// 下次启动前端读取后清除。
type UpdateState struct {
	// 状态：pending（已下载待替换）/ updated（已替换成功）/ error（替换失败）
	Status string `json:"status"`
	// 目标版本号（如 0.0.1）
	Version string `json:"version"`
	// 错误信息（仅 status=error 时非空）
	Error string `json:"error"`
}

// 更新状态常量
const (
	UpdatePending = "pending"
	UpdateUpdated = "updated"
	UpdateError   = "error"
)
