//go:build windows

package windows

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetTotalCores 获取总逻辑核心数
func GetTotalCores() int {
	psCmd := `(Get-CimInstance -ClassName Win32_ComputerSystem).NumberOfLogicalProcessors`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	var total int
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &total)
	return total
}

// getCPUName 获取 CPU 型号名称
func getCPUName() string {
	psCmd := `(Get-CimInstance -ClassName Win32_Processor).Name`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
