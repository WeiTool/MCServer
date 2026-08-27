package service

import (
	"MCServer/backend/model"
	"MCServer/backend/storage"
	"MCServer/backend/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var currentAppVersion = "0.0.1"

func SetCurrentVersion(v string) {
	if v != "" {
		currentAppVersion = v
	}
}

var versionRegex = regexp.MustCompile(`v(\d+\.\d+\.\d+)`)

// DownloadUpdate 下载新版本压缩包到 cache 目录
func DownloadUpdate(downloadURL string) error {
	cacheDir := utils.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	finalTemp := filepath.Join(cacheDir, "MCServer-update.tmp")
	partTemp := filepath.Join(cacheDir, "MCServer-update.part")

	_ = os.Remove(partTemp)
	_ = os.Remove(finalTemp)

	if err := downloadFile(downloadURL, partTemp); err != nil {
		return err
	}

	const minSize = 1 * 1024 * 1024
	info, err := os.Stat(partTemp)
	if err != nil {
		return fmt.Errorf("无法读取临时文件信息: %w", err)
	}
	if info.Size() < minSize {
		_ = os.Remove(partTemp)
		return fmt.Errorf("下载的文件过小（%d 字节），可能无效", info.Size())
	}

	if err := os.Rename(partTemp, finalTemp); err != nil {
		_ = os.Remove(partTemp)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	version := extractVersion(downloadURL)
	return writeUpdateState(model.UpdateState{
		Status:  model.UpdatePending,
		Version: version,
	})
}

// ApplyPendingUpdate 应用待执行的更新（从 cache 目录读取状态和压缩包）
// 流程：解压压缩包 → 提取可执行文件 → 调用平台替换脚本
func ApplyPendingUpdate() error {
	state, err := readUpdateState()
	if err != nil {
		return nil
	}
	if state.Status != model.UpdatePending {
		return nil
	}

	cacheDir := utils.GetCacheDir()
	archivePath := filepath.Join(cacheDir, "MCServer-update.tmp")

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		_ = os.Remove(updateStatePath())
		return nil
	}

	extractDir := filepath.Join(cacheDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		_ = writeErrorState("创建解压目录失败: " + err.Error())
		return err
	}
	defer os.RemoveAll(extractDir)

	if err := utils.ExtractArchive(archivePath, extractDir); err != nil {
		_ = writeErrorState("解压失败: " + err.Error())
		return err
	}

	// 查找可执行文件（逻辑不变）
	var exePath string
	if runtime.GOOS == "windows" {
		err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".exe") {
				exePath = path
				return filepath.SkipAll
			}
			return nil
		})
	} else {
		err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "MCServer" && info.Mode()&0111 != 0 {
				exePath = path
				return filepath.SkipAll
			}
			return nil
		})
	}
	if err != nil || exePath == "" {
		_ = writeErrorState("未找到可执行文件")
		return fmt.Errorf("未找到可执行文件")
	}

	finalExe := filepath.Join(cacheDir, "MCServer-update.bin")
	if err := os.Rename(exePath, finalExe); err != nil {
		_ = writeErrorState("移动可执行文件失败: " + err.Error())
		return err
	}

	_ = os.Remove(archivePath)

	exeDir := utils.GetExeDir()
	if exeDir == "" {
		return fmt.Errorf("无法获取可执行文件目录")
	}

	// 获取当前运行的可执行文件名称（用于删除旧文件）
	curExePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取当前可执行文件路径: %w", err)
	}
	currentExeName := filepath.Base(curExePath)

	// 固定目标文件名
	var targetExeName string
	if runtime.GOOS == "windows" {
		targetExeName = "MCServer.exe"
	} else {
		targetExeName = "MCServer"
	}

	// 调用平台替换脚本（传入当前文件名和目标文件名）
	return launchUpdater(exeDir, finalExe, currentExeName, targetExeName, cacheDir, state.Version)
}

// writeErrorState 写入 error 状态（辅助函数）
func writeErrorState(errMsg string) error {
	return writeUpdateState(model.UpdateState{
		Status: model.UpdateError,
		Error:  errMsg,
	})
}

const updateFileName = "update.json"

// updateStatePath 返回状态文件路径（位于 cache 目录）
func updateStatePath() string {
	return filepath.Join(utils.GetCacheDir(), updateFileName)
}

func readUpdateState() (model.UpdateState, error) {
	data, err := os.ReadFile(updateStatePath())
	if err != nil {
		return model.UpdateState{}, err
	}
	var state model.UpdateState
	if err := json.Unmarshal(data, &state); err != nil {
		return model.UpdateState{}, err
	}
	return state, nil
}

func writeUpdateState(state model.UpdateState) error {
	cacheDir := utils.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(updateStatePath(), data, 0644)
}

func extractVersion(downloadURL string) string {
	m := versionRegex.FindStringSubmatch(downloadURL)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// GetUpdateState 读取并清除更新状态（从 cache 目录）
func GetUpdateState() (model.UpdateState, error) {
	state, err := readUpdateState()
	if err != nil {
		return model.UpdateState{}, err
	}
	if state.Status != model.UpdatePending {
		_ = os.Remove(updateStatePath())
	}
	return state, nil
}

func downloadFile(downloadURL, destPath string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP 状态码: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	if err := out.Sync(); err != nil {
		return fmt.Errorf("同步文件失败: %w", err)
	}
	return nil
}

// launchUpdater 由各平台实现（updater_windows.go / updater_linux.go）
// 参数 configDir 现在是 cache 目录，脚本将状态文件写入该目录
// 注意：脚本自身仍然生成在 exeDir 中，以便执行

// 资产文件名匹配（与版本号、平台、压缩包后缀强匹配）
// 正式版：MCServer-v{major}.{minor}.{patch}-{os}-{arch}.zip / .tar.gz / .tgz
var stableAssetRegex = regexp.MustCompile(`^MCServer-v(\d+)\.(\d+)\.(\d+)-(\w+)-(\w+)\.(zip|tar\.gz|tgz)$`)

// 预览版：MCServer-v{major}.{minor}.{patch}-beta{num}-{os}-{arch}.zip / .tar.gz / .tgz
var previewAssetRegex = regexp.MustCompile(`^MCServer-v(\d+)\.(\d+)\.(\d+)-beta(\d+)-(\w+)-(\w+)\.(zip|tar\.gz|tgz)$`)

// stableBetaRank 正式版 beta 排名标记：视为高于任何 beta 编号
const stableBetaRank = 1 << 30

// parsedAsset 解析后的发布资产
type parsedAsset struct {
	version [3]int // major.minor.patch
	betaNum int    // 正式版为 stableBetaRank，预览版为 beta 编号
	name    string
	url     string
}

// parseAssetName 解析资产文件名；不匹配返回 nil
func parseAssetName(name, url string) *parsedAsset {
	if m := previewAssetRegex.FindStringSubmatch(name); len(m) >= 7 {
		if m[5] != runtime.GOOS || m[6] != runtime.GOARCH {
			return nil
		}
		beta, _ := strconv.Atoi(m[4])
		return &parsedAsset{
			version: [3]int{atoiOrZero(m[1]), atoiOrZero(m[2]), atoiOrZero(m[3])},
			betaNum: beta,
			name:    name,
			url:     url,
		}
	}
	if m := stableAssetRegex.FindStringSubmatch(name); len(m) >= 6 {
		if m[4] != runtime.GOOS || m[5] != runtime.GOARCH {
			return nil
		}
		return &parsedAsset{
			version: [3]int{atoiOrZero(m[1]), atoiOrZero(m[2]), atoiOrZero(m[3])},
			betaNum: stableBetaRank,
			name:    name,
			url:     url,
		}
	}
	return nil
}

// compareAssets 判断 a 是否比 b 新
// 规则：先比较 v 版本（major.minor.patch），相同则比较 beta 编号，正式版视为最终版（高于任何 beta）
func compareAssets(a, b *parsedAsset) bool {
	for i := range 3 {
		if a.version[i] != b.version[i] {
			return a.version[i] > b.version[i]
		}
	}
	return a.betaNum > b.betaNum
}

func atoiOrZero(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// CheckVersion 检查新版本，匹配压缩包（Windows: .zip, Linux: .tar.gz/.tgz）
// 依据全局配置决定是否纳入预览版：预览版关闭时只比较正式版
func CheckVersion() (model.VersionResponse, error) {
	resp := model.VersionResponse{Current: currentAppVersion}

	cfg, err := storage.GetGlobalConfig()
	if err != nil {
		return resp, err
	}

	releases, err := GetReleases()
	if err != nil {
		return resp, err
	}

	var best *parsedAsset
	for _, rel := range releases {
		for _, asset := range rel.Assets {
			if asset.Name == "" {
				continue
			}
			parsed := parseAssetName(asset.Name, asset.BrowserDownloadURL)
			if parsed == nil {
				continue
			}
			// 预览版未开启时跳过预览资产
			if !cfg.PreviewEnabled && parsed.betaNum != stableBetaRank {
				continue
			}
			if best == nil || compareAssets(parsed, best) {
				best = parsed
			}
		}
	}

	if best == nil {
		return resp, nil
	}

	resp.AssetName = best.name
	resp.DownloadURL = best.url
	resp.Latest = fmt.Sprintf("%d.%d.%d", best.version[0], best.version[1], best.version[2])
	resp.IsBeta = best.betaNum != stableBetaRank
	resp.IsPreview = resp.IsBeta
	resp.HasUpdate = resp.Latest != currentAppVersion
	return resp, nil
}

// GetReleases 获取全部 release（Gitee API），用于跨版本比较
func GetReleases() ([]model.Release, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://gitee.com/api/v5/repos/weitool/MCServer/releases?per_page=50"

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

	var releases []model.Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}
	return releases, nil
}
