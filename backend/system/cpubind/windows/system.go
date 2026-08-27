//go:build windows

package windows

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	procGetNativeSystemInfo = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetNativeSystemInfo")
)

// systemInfo 对应 SYSTEM_INFO 结构（64 位布局）
type systemInfo struct {
	oemID                     uint32
	pageSize                  uint32
	minimumApplicationAddress uintptr
	maximumApplicationAddress uintptr
	activeProcessorMask       uintptr
	numberOfProcessors        uint32
	processorType             uint32
	allocationGranularity     uint32
	processorLevel            uint16
	processorRevision         uint16
}

// GetTotalCores 获取总逻辑核心数
// 直接调用 kernel32!GetNativeSystemInfo，读取 dwNumberOfProcessors
func GetTotalCores() int {
	var si systemInfo
	r, _, _ := procGetNativeSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	if r == 0 {
		return 0
	}
	return int(si.numberOfProcessors)
}

// getCPUName 获取 CPU 型号名称
// 直接读取注册表 HKLM\HARDWARE\DESCRIPTION\System\CentralProcessor\0 的
// ProcessorNameString 值（与 Win32_Processor.Name 同源，无需 WMI）
func getCPUName() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()

	name, _, err := k.GetStringValue("ProcessorNameString")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}
