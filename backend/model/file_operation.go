package model

type FileOperationResponse struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 提示信息
	Path    string `json:"path"`    // 解压到了哪里
}
