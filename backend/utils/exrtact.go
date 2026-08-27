package utils

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
)

// 支持的压缩格式常量
const (
	ArchiveZip    = "zip"
	ArchiveTar    = "tar"
	ArchiveTarGz  = "tar.gz"
	ArchiveTarBz2 = "tar.bz2"
)

// ExtractArchive 自动识别压缩包格式并解压到指定目录
// 支持 .zip, .tar, .tar.gz, .tar.bz2
func ExtractArchive(archivePath, destDir string) error {
	kind, err := DetectArchiveType(archivePath)
	if err != nil {
		return err
	}

	switch kind {
	case ArchiveZip:
		return extractZip(archivePath, destDir)
	case ArchiveTar:
		return extractTarFile(archivePath, destDir)
	case ArchiveTarGz:
		return extractTarGz(archivePath, destDir)
	case ArchiveTarBz2:
		return extractTarBz2(archivePath, destDir)
	default:
		return fmt.Errorf("不支持的压缩格式: %s", kind)
	}
}

// DetectArchiveType 读取文件头魔数识别压缩格式
func DetectArchiveType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开压缩包失败: %v", err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := io.ReadFull(f, buf)

	// zip 魔数
	if n >= 4 && buf[0] == 'P' && buf[1] == 'K' && (buf[2] == 3 || buf[2] == 5 || buf[2] == 7) && buf[3] == 4 {
		return ArchiveZip, nil
	}
	// gzip
	if n >= 2 && buf[0] == 0x1F && buf[1] == 0x8B {
		return ArchiveTarGz, nil
	}
	// bzip2
	if n >= 3 && buf[0] == 'B' && buf[1] == 'Z' && buf[2] == 'h' {
		return ArchiveTarBz2, nil
	}
	// tar (ustar)
	if n >= 262 && string(buf[257:262]) == "ustar" {
		return ArchiveTar, nil
	}

	// 回退到扩展名
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return ArchiveZip, nil
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return ArchiveTarGz, nil
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return ArchiveTarBz2, nil
	case strings.HasSuffix(lower, ".tar"):
		return ArchiveTar, nil
	}
	return "", fmt.Errorf("无法识别的压缩格式，支持: zip / tar / tar.gz / tar.bz2")
}

// extractZip 解压 zip 到目标目录（含路径穿越防护）
func extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(destDir, file.Name)
		if !withinDir(destDir, filePath) {
			return fmt.Errorf("非法路径穿越: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败: %v", err)
		}
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
func extractTarFile(path, destDir string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return extractTar(f, destDir)
}

// extractTarGz 解压 tar.gz
func extractTarGz(path, destDir string) error {
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
	return extractTar(gz, destDir)
}

// extractTarBz2 解压 tar.bz2
func extractTarBz2(path, destDir string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return extractTar(bzip2.NewReader(f), destDir)
}

// extractTar 从 tar reader 解压到目标目录（含路径穿越防护）
func extractTar(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		filePath := filepath.Join(destDir, hdr.Name)
		if !withinDir(destDir, filePath) {
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
