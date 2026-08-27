package service

import (
	"fmt"
	"os"
	"path/filepath"

	"MCServer/backend/model"
	"MCServer/backend/utils"
)

// ExtractServerZip 解压 Minecraft 服务器压缩包到 servers 目录
// 支持 zip / tar / tar.gz / tar.bz2 四种格式（按文件头魔数自动识别）
func ExtractServerZip(zipPath, serverName string) (model.FileOperationResponse, error) {
	if zipPath == "" {
		return model.FileOperationResponse{Success: false, Message: "压缩包路径不能为空"}, nil
	}
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("压缩包不存在: %s", zipPath)}, nil
	}

	serversRootDir := utils.GetServersRoot()
	serverDir := filepath.Join(serversRootDir, serverName)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("创建目录失败: %v", err)}, nil
	}

	// 直接调用 utils.ExtractArchive，它内部会自动识别格式
	if err := utils.ExtractArchive(zipPath, serverDir); err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("解压失败: %v", err)}, nil
	}

	return model.FileOperationResponse{
		Success: true,
		Message: fmt.Sprintf("解压成功！服务器已安装在: %s", serverDir),
		Path:    serverDir,
	}, nil
}
