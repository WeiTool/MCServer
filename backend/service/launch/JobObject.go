//go:build windows

// 使用 Windows 作业对象（Job Object）管理 Java 子进程
// 当 MCServer 进程退出时，Job 句柄关闭，所有 Java 子进程被系统自动终止
// 从而确保关闭 APP 时所有服务端进程都被清理
package launch

import (
	"syscall"
	"unsafe"
)

// Windows Job Object API
var (
	jobProcCreateJobObject  = kernel32.NewProc("CreateJobObjectW")
	jobProcSetInfoJobObject = kernel32.NewProc("SetInformationJobObject")
	jobProcAssignProcess    = kernel32.NewProc("AssignProcessToJobObject")
)

// JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
// 当 Job 的最后一个句柄被关闭时，终止 Job 内所有进程
const jobObjectLimitKillOnJobClose = 0x2000

// JOBOBJECT_EXTENDED_LIMIT_INFORMATION 结构体
// 仅需要 LimitFlags 字段，其余字段占位
type jobObjectExtendedLimitInformation struct {
	BasicLimitInformation struct {
		PerProcessUserTimeLimit uint64
		PerJobUserTimeLimit     uint64
		LimitFlags              uint32
		MinWorkingSetSize       uintptr
		MaxWorkingSetSize       uintptr
		ActiveProcessLimit      uint32
		Affinity                uintptr
		PriorityClass           uint32
		SchedulingClass         uint32
	}
	IoInfo                 [3]uint64
	ProcessMemoryLimit     uintptr
	JobMemoryLimit         uintptr
	PeakProcessMemoryUsed  uintptr
	PeakJobMemoryUsed      uintptr
}

// jobHandle 保存 Job 对象句柄
var jobHandle syscall.Handle

// InitJobObject 初始化作业对象
// 创建一个 Job 对象并设置 KILL_ON_JOB_CLOSE，
// 使得 MCServer 退出时所有加入该 Job 的进程自动终止
func InitJobObject() error {
	// 创建 Job 对象（无名称）
	handle, _, err := jobProcCreateJobObject.Call(0, 0)
	if handle == 0 {
		return err
	}
	jobHandle = syscall.Handle(handle)

	// 设置 KILL_ON_JOB_CLOSE 标志
	info := &jobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose

	r, _, err := jobProcSetInfoJobObject.Call(
		uintptr(handle),
		9, // JobObjectExtendedLimitInformation
		uintptr(unsafe.Pointer(info)),
		unsafe.Sizeof(*info),
	)
	if r == 0 {
		// 设置失败时关闭句柄并返回错误
		procCloseHandle.Call(uintptr(handle))
		return err
	}
	return nil
}

// AssignProcessToJob 将指定 PID 的进程加入作业对象
// 加入后，进程将受 Job 的 KILL_ON_JOB_CLOSE 管理
func AssignProcessToJob(pid int) error {
	if jobHandle == 0 {
		return nil
	}

	// 打开进程句柄（需要 PROCESS_SET_QUOTA | PROCESS_TERMINATE）
	processHandle, err := openProcess(pid)
	if err != nil {
		return err
	}
	defer procCloseHandle.Call(uintptr(processHandle))

	r, _, err := jobProcAssignProcess.Call(uintptr(jobHandle), uintptr(processHandle))
	if r == 0 {
		return err
	}
	return nil
}
