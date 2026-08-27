package server

import (
	"MCServer/backend/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// 当前应用版本（用于与最新 release 的版本号比较，判断是否有新版本）
// 单一来源：wails.json 的 info.productVersion（main 包嵌入读取后注入），
// 无需 ldflags；改版只改 wails.json 一处，wails build -clean 即生效。
var currentAppVersion = "0.0.1"

// SetCurrentVersion 设置当前应用版本
// 由 main 包启动时从嵌入的 wails.json 读取 productVersion 注入
func SetCurrentVersion(v string) {
	if v != "" {
		currentAppVersion = v
	}
}

// versionRegex 匹配 release 文件名中的版本号：MCServer-v0.0.1-beta1-windows-amd64.exe → 0.0.1
var versionRegex = regexp.MustCompile(`v(\d+\.\d+\.\d+)`)

// DownloadFile 从 URL 下载文件到当前目录
func DownloadAndRename(downloadURL string) error {
	// 定义文件名
	newFileName := "MCServer.exe"              // 目标文件名
	tempFileName := filepath.Base(downloadURL) // 临时文件名（从 URL 提取）

	// 删除旧的 MCServer.exe（如果存在）
	if _, err := os.Stat(newFileName); err == nil {
		fmt.Printf("删除旧文件: %s\n", newFileName)
		if err := os.Remove(newFileName); err != nil {
			return fmt.Errorf("删除旧文件失败: %w", err)
		}
	}

	// 下载文件到临时文件名
	fmt.Printf("正在下载: %s\n", tempFileName)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP 状态码: %s", resp.Status)
	}

	out, err := os.Create(tempFileName)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	fmt.Printf("下载完成: %s\n", tempFileName)

	// 将临时文件重命名为 MCServer.exe
	fmt.Printf("重命名: %s -> %s\n", tempFileName, newFileName)
	if err := os.Rename(tempFileName, newFileName); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	fmt.Printf("完成！新文件: %s\n", newFileName)
	return nil
}

// CheckVersion 检查是否有新版本
// 流程：
//  1. 调用 GetNameAndUrl 获取最新 release 的两个下载文件名与地址
//  2. 根据当前系统（runtime.GOOS/GOARCH，如 windows-amd64）与文件名关键词匹配
//     找到属于当前平台的下载包
//  3. 从文件名中提取版本号（如 MCServer-v0.0.1-beta1-windows-amd64.exe → 0.0.1）
//  4. 与当前版本 currentAppVersion 比较，不一致则视为有新版本
func CheckVersion() (model.VersionResponse, error) {
	// 当前平台关键词：如 windows-amd64 / linux-amd64 / darwin-arm64
	platform := runtime.GOOS + "-" + runtime.GOARCH

	name1, url1, name2, url2 := GetNameAndUrl()
	assets := []struct{ name, url string }{
		{name1, url1},
		{name2, url2},
	}

	// 匹配当前平台的下载包（文件名中包含平台关键词）
	var matchedName, matchedURL string
	for _, a := range assets {
		if a.name == "" {
			continue
		}
		// 文件名形如 MCServer-v0.0.1-beta1-windows-amd64.exe
		if strings.Contains(a.name, platform) {
			matchedName = a.name
			matchedURL = a.url
			break
		}
	}

	resp := model.VersionResponse{
		Current: currentAppVersion,
	}

	// 未找到当前平台的下载包（如 release 只发布了 Windows），视为无更新，不报错
	if matchedName == "" {
		return resp, nil
	}

	// 提取版本号 vX.Y.Z
	m := versionRegex.FindStringSubmatch(matchedName)
	if len(m) < 2 {
		return resp, fmt.Errorf("无法从文件名解析版本号: %s", matchedName)
	}
	latestVersion := m[1]

	resp.Latest = latestVersion
	resp.AssetName = matchedName
	resp.DownloadURL = matchedURL
	// 文件名含 beta 即视为预发布版本
	resp.IsBeta = strings.Contains(strings.ToLower(matchedName), "beta")
	// 与当前版本不一致 → 有新版本
	resp.HasUpdate = latestVersion != currentAppVersion

	return resp, nil
}

func GetNameAndUrl() (string, string, string, string) {
	release, err := GetLatestRelease()
	if err != nil {
		return "", "", "", ""
	}

	// 安全处理：如果 Assets 为空或数量不足，返回空字符串
	if len(release.Assets) < 2 {
		return "", "", "", ""
	}

	name1 := release.Assets[0].Name
	url1 := release.Assets[0].BrowserDownloadURL
	name2 := release.Assets[1].Name
	url2 := release.Assets[1].BrowserDownloadURL

	return name1, url1, name2, url2
}

// GetLatestRelease 获取最新 Release，直接返回 Release 结构体
func GetLatestRelease() (*model.Release, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := "https://gitee.com/api/v5/repos/weitool/MCServer/releases/latest"

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var release model.Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	return &release, nil
}
