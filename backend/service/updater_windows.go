//go:build windows

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"MCServer/backend/utils"
)

// launchUpdater 生成并启动 Windows 替换脚本（.bat）
// 由应用关闭回调触发，脚本在独立进程（schtasks 计划任务）中运行：
// 探测主进程退出 → 删除旧 exe → 用临时文件替换（重试）→ 写 config/update.json 结果。
// 不自动重启——替换完成后由用户下次手动打开应用。
//
// 重要：bat 内容必须全部 ASCII（无中文注释/路径），否则 cmd 解析错乱。
func launchUpdater(exeDir, tempFile, currentExeName, targetExeName, configDir, version string) error {
	scriptPath := filepath.Join(exeDir, "update.bat")
	stateFile := filepath.Join(configDir, updateFileName)

	// 删除模式：使用当前文件名（可能带 .exe）
	delPattern := filepath.Join(exeDir, strings.TrimSuffix(currentExeName, ".exe")+"*.exe")

	escVersion := strings.ReplaceAll(version, `"`, `\"`)
	script := `@echo off
rem poll for main process exit (max 30s, every 0.5s), proceed as soon as it exits
set /a count=0
:wait
tasklist /FI "IMAGENAME eq ` + currentExeName + `" 2>nul | find /I "` + currentExeName + `" >nul
if errorlevel 1 goto replace
ping -n 1 -w 500 127.0.0.1 >nul
set /a count+=1
if %count% lss 60 goto wait
:replace
rem delete old exe using current name pattern
del /Q "` + delPattern + `" 2>nul
rem replace with temp file (retry 20 x 0.5s = 10s window)
set /a tries=0
:move_retry
move /Y "` + tempFile + `" "` + filepath.Join(exeDir, targetExeName) + `" >nul 2>nul
if not errorlevel 1 goto ok
ping -n 1 -w 500 127.0.0.1 >nul
set /a tries+=1
if %tries% lss 20 goto move_retry
rem move failed: record error state
echo {"status":"error","version":"` + escVersion + `","error":"replace failed after retries"} > "` + stateFile + `"
goto end
:ok
rem replaced successfully: record updated state
echo {"status":"updated","version":"` + escVersion + `","error":""} > "` + stateFile + `"
:end
rem delete self
del /Q "` + scriptPath + `" 2>nul
`

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("写入更新脚本失败: %w", err)
	}

	return scheduleUpdater(scriptPath)
}

// scheduleUpdater 通过计划任务启动更新脚本
// 分两步：先创建一次性任务（/f 覆盖同名），再立即运行。
// 任务由系统 Task Scheduler 服务执行，与当前进程完全解耦。
func scheduleUpdater(scriptPath string) error {
	taskName := "MCServerUpdater"
	// 任务命令：cmd /c "update.bat"，路径带空格时需引号包裹
	cmdLine := fmt.Sprintf(`cmd /c "%s"`, scriptPath)
	// /st 需要 HH:MM 未来时间（+24h 避免跨天导致"开始时间早于当前"被拒），
	// 随后用 /run 立即触发，时间仅作占位
	st := time.Now().Add(24 * time.Hour).Format("15:04")

	create := exec.Command("schtasks",
		"/create", "/tn", taskName,
		"/tr", cmdLine,
		"/sc", "once", "/st", st,
		"/f",
	)
	create.SysProcAttr = utils.NewHiddenSysProcAttr()
	if out, err := create.CombinedOutput(); err != nil {
		return fmt.Errorf("创建计划任务失败: %v (%s)", err, string(out))
	}

	run := exec.Command("schtasks", "/run", "/tn", taskName)
	run.SysProcAttr = utils.NewHiddenSysProcAttr()
	if out, err := run.CombinedOutput(); err != nil {
		return fmt.Errorf("触发计划任务失败: %v (%s)", err, string(out))
	}
	return nil
}
