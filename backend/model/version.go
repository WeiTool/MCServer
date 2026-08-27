package model

// Release 映射 Gitee API 的 Release 响应
type Release struct {
	Assets []Asset `json:"assets"`
}

// Asset 嵌套结构体
type Asset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	Name               string `json:"name"`
}
