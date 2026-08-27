package server

import (
	"MCServer/backend/model"
	"MCServer/backend/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyJarFile 复制 JAR 文件到服务器目录
func CopyJarFile(jarPath, serverName string) (model.FileOperationResponse, error) {
	// 参数校验
	if jarPath == "" {
		return model.FileOperationResponse{Success: false, Message: "JAR路径不能为空"}, nil
	}

	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("JAR文件不存在: %s", jarPath)}, nil
	}

	// 检查是否为 .jar 文件
	if !strings.HasSuffix(strings.ToLower(jarPath), ".jar") {
		return model.FileOperationResponse{Success: false, Message: "文件不是 JAR 格式"}, nil
	}

	// 获取服务器根目录
	serversRootDir := utils.GetServersRoot()
	serverDir := filepath.Join(serversRootDir, serverName)

	// 创建服务器目录
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("创建目录失败: %v", err)}, nil
	}

	// 目标路径
	fileName := filepath.Base(jarPath)
	destPath := filepath.Join(serverDir, fileName)

	// 复制文件
	srcFile, err := os.Open(jarPath)
	if err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("打开源文件失败: %v", err)}, nil
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("创建目标文件失败: %v", err)}, nil
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("复制文件失败: %v", err)}, nil
	}

	return model.FileOperationResponse{
		Success: true,
		Message: fmt.Sprintf("JAR 文件已复制到: %s", destPath),
		Path:    serverDir,
	}, nil
}
