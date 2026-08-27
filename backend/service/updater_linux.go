//go:build linux

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// launchUpdater 生成并启动 Linux 替换脚本（.sh）
// 由应用关闭回调触发，脚本在独立进程（Setsid 分离）中运行：
// 探测主进程退出 → 删除旧 exe → 用临时文件替换（重试）→ 写 config/update.json 结果。
// 不自动重启——替换完成后由用户下次手动打开应用。
func launchUpdater(exeDir, tempFile, currentExeName, targetExeName, configDir, version string) error {
	scriptPath := filepath.Join(exeDir, "update.sh")
	stateFile := filepath.Join(configDir, updateFileName)
	procName := filepath.Base(currentExeName)

	escVersion := strings.ReplaceAll(version, `"`, `\"`)
	script := `#!/bin/sh
# wait for main process exit (max 30s, every 0.5s)
i=0
while [ $i -lt 60 ]; do
  if ! pidof "` + procName + `" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
  i=$((i+1))
done
# delete old exe (current name)
rm -f "` + filepath.Join(exeDir, currentExeName) + `"
# replace with temp file (retry 10 x 0.5s)
tries=0
while [ $tries -lt 10 ]; do
  if mv -f "` + tempFile + `" "` + filepath.Join(exeDir, targetExeName) + `" 2>/dev/null; then
    break
  fi
  sleep 0.5
  tries=$((tries+1))
done
chmod +x "` + filepath.Join(exeDir, targetExeName) + `"
if [ -f "` + filepath.Join(exeDir, targetExeName) + `" ]; then
  echo '{"status":"updated","version":"` + escVersion + `","error":""}' > "` + stateFile + `"
else
  echo '{"status":"error","version":"` + escVersion + `","error":"replace failed"}' > "` + stateFile + `"
fi
# delete self
rm -f "` + scriptPath + `"
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("写入更新脚本失败: %w", err)
	}

	cmd := exec.Command("/bin/sh", scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新脚本失败: %w", err)
	}
	_ = cmd.Process.Release()
	return nil
}
