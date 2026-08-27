package server

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"MCServer/backend/model"
	"MCServer/backend/utils"
)

// 支持识别的压缩包格式（按魔数检测，不依赖扩展名）
const (
	archiveZip    = "zip"
	archiveTar    = "tar"
	archiveTarGz  = "tar.gz"
	archiveTarBz2 = "tar.bz2"
)

// ExtractServerZip 解压 Minecraft 服务器压缩包到 servers 目录
// 支持 zip / tar / tar.gz / tar.bz2 四种格式（按文件头魔数自动识别，Windows/Linux 通用）
func ExtractServerZip(zipPath, serverName string) (model.FileOperationResponse, error) {
	// 参数校验
	if zipPath == "" {
		return model.FileOperationResponse{Success: false, Message: "压缩包路径不能为空"}, nil
	}

	// 检查压缩包是否存在
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("压缩包不存在: %s", zipPath)}, nil
	}

	// 识别压缩格式
	kind, err := detectArchiveType(zipPath)
	if err != nil {
		return model.FileOperationResponse{Success: false, Message: err.Error()}, nil
	}

	// 获取 servers 服务器根目录的绝对路径（exe 同级）
	serversRootDir := utils.GetServersRoot()
	serverDir := filepath.Join(serversRootDir, serverName)

	// 创建目标目录
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("创建目录失败: %v", err)}, nil
	}

	// 按格式分发解压
	var extractErr error
	switch kind {
	case archiveZip:
		extractErr = extractZip(zipPath, serverDir)
	case archiveTar:
		extractErr = extractTarFile(zipPath, serverDir)
	case archiveTarGz:
		extractErr = extractTarGz(zipPath, serverDir)
	case archiveTarBz2:
		extractErr = extractTarBz2(zipPath, serverDir)
	default:
		extractErr = fmt.Errorf("不支持的压缩格式")
	}

	if extractErr != nil {
		return model.FileOperationResponse{Success: false, Message: fmt.Sprintf("解压失败: %v", extractErr)}, nil
	}

	return model.FileOperationResponse{
		Success: true,
		Message: fmt.Sprintf("解压成功！服务器已安装在: %s", serverDir),
		Path:    serverDir,
	}, nil
}

// detectArchiveType 读取文件头魔数识别压缩格式
// zip: "PK\x03\x04"；gzip: 0x1F 0x8B；bzip2: "BZh"；tar: offset 257 处 "ustar"
// 魔数无法命中时回退到扩展名判断
func detectArchiveType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开压缩包失败: %v", err)
	}
	defer f.Close()

	// 读取前 512 字节（tar 的 ustar 标记在 257 偏移处）
	buf := make([]byte, 512)
	n, _ := io.ReadFull(f, buf)

	// zip 魔数
	if n >= 4 && buf[0] == 'P' && buf[1] == 'K' && (buf[2] == 3 || buf[2] == 5 || buf[2] == 7) && buf[3] == 4 {
		return archiveZip, nil
	}
	// gzip 魔数（tar.gz / tgz）
	if n >= 2 && buf[0] == 0x1F && buf[1] == 0x8B {
		return archiveTarGz, nil
	}
	// bzip2 魔数（tar.bz2）
	if n >= 3 && buf[0] == 'B' && buf[1] == 'Z' && buf[2] == 'h' {
		return archiveTarBz2, nil
	}
	// tar 魔数（offset 257 处 "ustar"）
	if n >= 262 && string(buf[257:262]) == "ustar" {
		return archiveTar, nil
	}

	// 回退：按扩展名判断（魔数读取失败等场景）
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return archiveZip, nil
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return archiveTarGz, nil
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return archiveTarBz2, nil
	case strings.HasSuffix(lower, ".tar"):
		return archiveTar, nil
	}

	return "", fmt.Errorf("无法识别的压缩格式，支持: zip / tar / tar.gz / tar.bz2")
}

// extractZip 解压 zip 到目标目录（含路径穿越防护）
func extractZip(zipPath, serverDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(serverDir, file.Name)
		// 安全检查 - 防止路径穿越攻击
		if !withinDir(serverDir, filePath) {
			return fmt.Errorf("非法路径穿越: %s", file.Name)
		}

		// 目录
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
			continue
		}

		// 创建文件所在目录
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败: %v", err)
		}

		// 限制文件权限（去掉可执行位）
		perm := file.Mode() & 0666
		if perm == 0 {
			perm = 0644
		}
		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return fmt.Errorf("读取zip文件失败: %v", err)
		}

		_, err = io.Copy(destFile, srcFile)
		destFile.Close()
		srcFile.Close()
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}
	return nil
}

// extractTarFile 解压纯 tar 包
func extractTarFile(path, serverDir string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return extractTar(f, serverDir)
}

// extractTarGz 解压 tar.gz
func extractTarGz(path, serverDir string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	return extractTar(gz, serverDir)
}

// extractTarBz2 解压 tar.bz2
func extractTarBz2(path, serverDir string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return extractTar(bzip2.NewReader(f), serverDir)
}

// extractTar 从 tar reader 解压到目标目录（含路径穿越防护）
func extractTar(r io.Reader, serverDir string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		filePath := filepath.Join(serverDir, hdr.Name)
		// 安全检查 - 防止路径穿越攻击
		if !withinDir(serverDir, filePath) {
			return fmt.Errorf("非法路径穿越: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return fmt.Errorf("创建父目录失败: %v", err)
			}
			// 限制文件权限（去掉可执行位）
			perm := os.FileMode(hdr.Mode) & 0666
			if perm == 0 {
				perm = 0644
			}
			destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
			if err != nil {
				return fmt.Errorf("创建文件失败: %v", err)
			}
			_, err = io.Copy(destFile, tr)
			destFile.Close()
			if err != nil {
				return fmt.Errorf("写入文件失败: %v", err)
			}
		}
		// 符号链接等特殊类型忽略（Windows 创建链接需特权，跳过最安全）
	}
	return nil
}

// withinDir 检查 target 是否位于 dir 目录内（路径穿越防护）
func withinDir(dir, target string) bool {
	cleanDir := filepath.Clean(dir)
	cleanTarget := filepath.Clean(target)
	if cleanTarget == cleanDir {
		return true
	}
	return strings.HasPrefix(cleanTarget, cleanDir+string(os.PathSeparator))
}
