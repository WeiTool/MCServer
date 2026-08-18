// Package core 提供 Java 服务端进程的核心实现
// 负责启动服务器、实时推送日志、发送命令、停止进程
// 对外由 process 包的门面 ProcessService 提供统一接口
package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"MCServer/backend/service/config"
	"MCServer/backend/service/launch"
	"MCServer/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProcessManager 进程管理器
// 管理 Java 服务端进程的生命周期与日志推送
// 对外通过 process.ProcessService 门面调用
type ProcessManager struct {
	// 应用上下文（用于事件推送）
	ctx context.Context
	// 正在运行的 Java 进程
	cmd *exec.Cmd
	// stdin 写入管道（用于发送命令）
	stdin io.Writer
	// 当前正在运行的服务器名（用于 StopServer 定位需要刷新的统计）
	currentServer string
	// 进程操作互斥锁，防止并发启动/停止
	mu sync.Mutex
	// 配置服务：统一读取服务器持久化配置（Java 路径、内存等）
	cfg *config.ConfigService
	// 检测状态机（类型/版本检测），启动时创建，停止时清空
	detector *DetectionState
	// 服务器启动时间（用于计算运行时长），停止时清零
	startTime time.Time
}

// NewProcessManager 创建进程管理器实例
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		// 初始化配置服务，供读取服务器 Java/内存配置使用
		cfg: config.NewConfigService(),
	}
}

// Startup 保存应用上下文
func (s *ProcessManager) Startup(ctx context.Context) {
	s.ctx = ctx
	// 初始化作业对象：MCServer 退出时自动终止所有 Java 子进程
	_ = launch.InitJobObject()
}

// StartServer 启动指定服务器文件夹中的 Java 服务端
// 使用该服务器在 ServerList.json 中配置的 Java 路径启动
// 流程: 定位 jar -> 构造命令 -> 启动进程 -> 绑定 CPU -> 逐行推送日志
func (s *ProcessManager) StartServer(serverName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 若已有进程在运行，先提示
	if s.cmd != nil && s.cmd.Process != nil {
		return fmt.Errorf("服务器已在运行")
	}

	// 0. eula 检查（启动前）：
	//    - eula.txt 存在但未同意 -> 直接改为 true，再继续启动
	//    - eula.txt 缺失 -> 记录状态，启动后由 eula 检测 goroutine 处理（等文件出现→关服→改true→重启）
	eulaStatus := CheckEula(serverName)
	if eulaStatus == EulaNotAgreed {
		// 存在但未同意，先改为 true
		if err := SetEulaAgreed(serverName); err != nil {
			return fmt.Errorf("修改 eula.txt 失败: %v", err)
		}
		runtime.EventsEmit(s.ctx, "server:log", "> 已自动同意 EULA（eula=true）")
	}

	// 1. 定位 jar 文件
	jarPath, err := s.findJarPath(serverName)
	if err != nil {
		return err
	}

	// 2. 确定 java 可执行文件路径（从 ServerList.json 读取该服务器配置的 Java）
	//    必须手动选择 Java 才能启动：未配置时直接报错，绝不回退到系统默认 java
	javaExe := s.cfg.GetServerJava(serverName)
	if javaExe == "" {
		return fmt.Errorf("服务器 [%s] 未配置 Java，请先在右侧手动选择一个 Java 后再启动", serverName)
	}

	// 3. 读取该服务器配置的内存（MB）
	xmxMB, xmsMB := s.cfg.GetServerMemory(serverName)

	// 4. 构造启动命令参数（java [JVM参数] -Xmx{max}M -Xms{min}M -jar {jar} nogui）
	//    注意：不使用 -XX:+ForceUnbufferedStdout（部分新 JDK 不支持会报错）
	var args []string
	if xmxMB > 0 {
		args = append(args, fmt.Sprintf("-Xmx%dM", xmxMB))
	}
	if xmsMB > 0 {
		args = append(args, fmt.Sprintf("-Xms%dM", xmsMB))
	}
	// 启用服务端 JIT 编译模式
	args = append(args, "-server")
	// 禁用图形环境
	args = append(args, "-Djava.awt.headless=true")
	// 使用 G1 垃圾收集器
	args = append(args, "-XX:+UseG1GC")
	// G1 目标最大停顿时间
	args = append(args, "-XX:MaxGCPauseMillis=240")
	// 限制并行 GC 线程数为 1，防止与 MC 主线程抢核
	args = append(args, "-XX:ParallelGCThreads=1")
	// 限制 G1 并发标记线程数为 1
	args = append(args, "-XX:ConcGCThreads=1")
	// 禁用插件/模组触发的 System.gc()，防止强制 Full GC
	args = append(args, "-XX:+DisableExplicitGC")
	// 强制使用 UTF-8 编码读写，避免 Windows 中文环境下 Java 用 GBK 输出导致终端中文乱码
	args = append(args, "-Dfile.encoding=UTF-8")
	args = append(args, "-Dsun.stdout.encoding=UTF-8")
	args = append(args, "-Dsun.stderr.encoding=UTF-8")
	// 最终发起
	args = append(args, "-jar", jarPath, "nogui")

	// 工作目录设为服务器文件夹，以便正确加载 mods 等
	workDir := filepath.Dir(jarPath)
	cmd := exec.Command(javaExe, args...)
	cmd.Dir = workDir
	// 设置进程生命周期管理：
	// Linux 使用 Pdeathsig（父进程退出时自动终止子进程），Windows 使用 Job Object
	cmd.SysProcAttr = newSysProcAttr()

	// 5. 设置 stdout/stderr 管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取 stdout 管道失败: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取 stderr 管道失败: %v", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取 stdin 管道失败: %v", err)
	}

	// 6. 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Java 失败: %v", err)
	}
	s.cmd = cmd
	s.stdin = stdin
	s.currentServer = serverName
	// 记录启动时间，用于计算运行时长
	s.startTime = time.Now()

	// 将 Java 进程加入作业对象
	// 使其成为 MCServer 的受管子进程，退出时随主进程自动终止
	_ = launch.AssignProcessToJob(cmd.Process.Pid)

	// 推送启动提示
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 正在启动服务器: %s", serverName))
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 使用 Java: %s", javaExe))
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> MCServer PID=%d, Java PID=%d", os.Getpid(), cmd.Process.Pid))
	runtime.EventsEmit(s.ctx, "server:log", "> 绑定到负载最低的性能核心...")

	// 7. 绑定 CPU 到负载最低的性能核心
	if err := launch.BindToLowestLoadCore(cmd.Process.Pid); err != nil {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[警告] CPU 绑定失败: %v", err))
	} else {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 已绑定性能核心 (PID=%d)", cmd.Process.Pid))
	}

	// 8. 后台逐行读取并推送 stdout 和 stderr
	go s.streamLines(stdout)
	go s.streamLines(stderr)

	// 9. 初始化类型/版本检测状态机
	//    根据 json 中 type/version 是否为空，决定需要检测哪些项
	needType := s.cfg.GetServerType(serverName) == ""
	needVersion := s.cfg.GetServerVersion(serverName) == ""
	s.detector = NewDetectionState(serverName, needType, needVersion)
	if needType || needVersion {
		runtime.EventsEmit(s.ctx, "server:log", "> 检测到类型/版本缺失，启动自动检测...")
	}

	// 10. 若 eula.txt 缺失，启动 eula 检测 goroutine
	//     服务端首次启动会生成 eula.txt 并退出，检测到文件出现后：
	//     等 2 秒 -> 关服 -> 改 eula=true -> 自动重启 -> 继续检测
	if eulaStatus == EulaMissing {
		go s.watchEulaAndRestart(serverName)
	}

	return nil
}

// findJarPath 在服务器文件夹中找到 jar 文件路径
func (s *ProcessManager) findJarPath(serverName string) (string, error) {
	// 1. 通过 utils 获取该服务器的文件夹绝对路径
	folderPath := utils.GetServerFolderPath(serverName)

	// 2. 读取服务器文件夹目录
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return "", fmt.Errorf("读取服务器文件夹失败: %v", err)
	}

	// 3. 遍历条目，返回第一个 jar 文件的完整路径
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if utils.IsJarFile(entry.Name()) {
			return filepath.Join(folderPath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("服务器文件夹中未找到 jar 文件: %s", serverName)
}
