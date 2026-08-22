// Package process 提供 Java 服务端进程的生命周期管理
// 本文件实现 ProcessManager 核心逻辑：启动/停止/命令发送/日志流/类型检测编排
package process

import (
	"MCServer/backend/launch"
	"MCServer/backend/launch/detector"
	eula2 "MCServer/backend/launch/eula"
	"MCServer/backend/launch/stream"
	"MCServer/backend/launch/sysproc"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"MCServer/backend/storage"
	"MCServer/backend/system/cpubind"
	"MCServer/backend/system/gc"
	"MCServer/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProcessManager 进程管理器
// 管理 Java 服务端进程的生命周期与日志推送
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
	cfg *storage.Storage
	// 类型检测状态机，启动时创建，停止时清空
	detector *detector.DetectionState
	// 服务器启动时间（用于计算运行时长），停止时清零
	startTime time.Time
	// GC 统计服务（jstat 采集），启动时记录 Java bin 目录供其定位 jstat
	gc *gc.GCService
}

// NewProcessManager 创建进程管理器实例
func NewProcessManager(cfg *storage.Storage, gcService *gc.GCService) *ProcessManager {
	return &ProcessManager{
		// 使用外部注入的共享存储实例，读取服务器 Java/内存配置
		cfg: cfg,
		gc:  gcService,
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
// 流程: 检查 eula -> 构造命令 -> 启动进程 -> 接管 -> 绑定 CPU -> 收尾
func (s *ProcessManager) StartServer(serverName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 若已有进程在运行，先提示
	if s.cmd != nil && s.cmd.Process != nil {
		return fmt.Errorf("服务器已在运行")
	}

	// 1. 启动前检查：eula 未同意则自动改为 true
	eulaStatus, err := s.checkEulaBeforeStart(serverName)
	if err != nil {
		return err
	}

	// 2. 构造启动命令：定位 jar + 校验 Java + 组装 JVM 参数
	javaExe, cmd, err := s.buildLaunchCommand(serverName)
	if err != nil {
		return err
	}

	// 3. 创建进程管道（stdout/stderr/stdin）
	stdout, stderr, stdin, err := s.setupProcessPipes(cmd)
	if err != nil {
		return err
	}

	// 4. 启动 Java 进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Java 失败: %v", err)
	}

	// 5. 接管进程：登记状态、初始化检测状态机、启动日志流、加入作业对象
	needType := s.attachProcess(serverName, cmd, stdin, stdout, stderr, javaExe)

	// 6. 绑定 CPU 到所有性能核心（P-Core），避免落到 E 核
	//    放在接管之后，保证"启动服务器"相关日志先于绑定日志输出
	s.bindToPerformanceCores(cmd.Process.Pid)

	// 7. 启动后收尾：自动检测提示 + eula 缺失兜底
	s.afterStart(serverName, eulaStatus, needType)

	return nil
}

// checkEulaBeforeStart 启动前检查 eula.txt 并自动处理
// - 存在但未同意（eula=false）：自动改为 true 后继续
// - 缺失：返回 EulaMissing，由 afterStart 启动"等文件出现→关服→改true→重启"流程
func (s *ProcessManager) checkEulaBeforeStart(serverName string) (eula2.EulaStatus, error) {
	eulaStatus := eula2.CheckEula(serverName)
	if eulaStatus == eula2.EulaNotAgreed {
		// 存在但未同意，先改为 true
		if err := eula2.SetEulaAgreed(serverName); err != nil {
			return eulaStatus, fmt.Errorf("修改 eula.txt 失败: %v", err)
		}
		runtime.EventsEmit(s.ctx, "server:log", "> 已自动同意 EULA（eula=true）")
	}
	return eulaStatus, nil
}

// buildLaunchCommand 构造 Java 启动命令
// 流程: 定位 jar 文件 -> 校验该服务器配置的 Java -> 读取内存配置 -> 组装命令
// 必须手动选择 Java 才能启动：未配置时直接报错，绝不回退到系统默认 java
func (s *ProcessManager) buildLaunchCommand(serverName string) (string, *exec.Cmd, error) {
	// 定位 jar 文件
	jarPath, err := s.findJarPath(serverName)
	if err != nil {
		return "", nil, err
	}

	// 确定 java 可执行文件路径（从 ServerList.json 读取该服务器配置的 Java）
	javaExe := s.cfg.GetServerJava(serverName)
	if javaExe == "" {
		return "", nil, fmt.Errorf("服务器 [%s] 未配置 Java，请先在右侧手动选择一个 Java 后再启动", serverName)
	}

	// 读取该服务器配置的内存（MB）
	xmxMB, xmsMB := s.cfg.GetServerMemory(serverName)

	// 组装启动命令（工作目录设为服务器文件夹，以便正确加载 mods 等）
	cmd := exec.Command(javaExe, buildJvmArgs(jarPath, xmxMB, xmsMB)...)
	cmd.Dir = filepath.Dir(jarPath)
	// 设置进程生命周期管理：
	// Linux 使用 Pdeathsig（父进程退出时自动终止子进程），Windows 使用 Job Object
	cmd.SysProcAttr = sysproc.NewSysProcAttr()

	return javaExe, cmd, nil
}

// buildJvmArgs 组装 JVM 启动参数
// 注意：不使用 -XX:+ForceUnbufferedStdout（部分新 JDK 不支持会报错）
func buildJvmArgs(jarPath string, xmxMB, xmsMB int) []string {
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
	// 禁用插件/模组触发的 System.gc()，防止强制 Full GC
	args = append(args, "-XX:+DisableExplicitGC")
	// 强制使用 UTF-8 编码读写，避免 Windows 中文环境下 Java 用 GBK 输出导致终端中文乱码
	args = append(args, "-Dfile.encoding=UTF-8")
	args = append(args, "-Dsun.stdout.encoding=UTF-8")
	args = append(args, "-Dsun.stderr.encoding=UTF-8")
	// 最终发起
	args = append(args, "-jar", jarPath, "nogui")
	return args
}

// setupProcessPipes 创建进程的三条管道
// stdout/stderr：读取 Java 进程的输出（日志/JVM 警告），数据流 Java -> Go
// stdin：向 Java 进程写入控制台命令（stop、/version 等），数据流 Go -> Java
func (s *ProcessManager) setupProcessPipes(cmd *exec.Cmd) (io.Reader, io.Reader, io.WriteCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取 stdout 管道失败: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取 stderr 管道失败: %v", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取 stdin 管道失败: %v", err)
	}
	return stdout, stderr, stdin, nil
}

// attachProcess 启动成功后接管进程
// 登记运行状态、初始化类型检测状态机、启动日志流、加入作业对象
// 返回类型检测需求（供 afterStart 提示用）
func (s *ProcessManager) attachProcess(serverName string, cmd *exec.Cmd, stdin io.WriteCloser, stdout, stderr io.Reader, javaExe string) bool {
	// 登记进程状态与启动时间（用于计算运行时长）
	s.cmd = cmd
	s.stdin = stdin
	s.currentServer = serverName
	s.startTime = time.Now()

	// 记录所用 Java 的 bin 目录，供 GC 服务定位同版本 jstat
	if s.gc != nil {
		s.gc.SetJavaBin(filepath.Dir(javaExe))
	}

	// 初始化类型检测状态机（必须在 streamLines 之前）
	// 否则最早几行日志（含类型关键字）到达时 detector 仍为空，会被丢弃
	needType := s.cfg.GetServerType(serverName) == ""
	needVersion := s.cfg.GetServerVersion(serverName) == ""
	s.detector = detector.NewDetectionState(serverName, needType, needVersion)

	// 先推送启动信息，再启动日志流
	// 否则 Java 自己的输出会先于启动提示到达，打乱日志顺序
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 正在启动服务器: %s", serverName))
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 使用 Java: %s", javaExe))
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> MCServer PID=%d, Java PID=%d", os.Getpid(), cmd.Process.Pid))

	// 后台逐行读取并推送 stdout 和 stderr
	go stream.StreamLines(stdout, s.handleLogLine)
	go stream.StreamLines(stderr, s.handleLogLine)

	// 将 Java 进程加入作业对象
	// 使其成为 MCServer 的受管子进程，退出时随主进程自动终止
	_ = launch.AssignProcessToJob(cmd.Process.Pid)

	// 如果需要版本检测，启动版本检测 goroutine
	if needVersion {
		go s.detectVersion()
	}

	return needType
}

// bindToPerformanceCores 将 Java 进程绑定到全部性能核心（P-Core）
// 说明：Minecraft 服务端为单线程主循环（仅 Folia 为多线程），绑定到全 P 核
// 后 OS 会在 P 核之间调度主线程，避免跑在 E 核上；纯 P 核 CPU 则不产生实际限制
func (s *ProcessManager) bindToPerformanceCores(pid int) {
	cores, err := cpubind.GetPerformanceCores()
	if err != nil || len(cores) == 0 {
		runtime.EventsEmit(s.ctx, "server:log", "> 无法获取性能核心，跳过 CPU 绑定")
		return
	}

	// 先打印绑定的核心列表
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 将进程绑定到 P-Core: %v", cores))

	if err := cpubind.BindToPerformanceCores(pid); err != nil {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[警告] CPU 绑定失败: %v", err))
	} else {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 已绑定全 P-Core (PID=%d)", pid))
	}
}

// afterStart 启动后的收尾
// 提示类型自动检测；eula.txt 缺失时启动"等文件出现→关服→改true→重启"流程
func (s *ProcessManager) afterStart(serverName string, eulaStatus eula2.EulaStatus, needType bool) {
	if needType {
		runtime.EventsEmit(s.ctx, "server:log", "> 检测到类型缺失，启动自动检测...")
	}
	if eulaStatus == eula2.EulaMissing {
		// 服务端首次启动会生成 eula.txt 并退出，检测到文件出现后：
		// 等 2 秒 -> 关服 -> 改 eula=true -> 自动重启 -> 继续检测
		go eula2.WatchEulaAndRestart(serverName,
			func(msg string) { runtime.EventsEmit(s.ctx, "server:log", msg) },
			s.stopAndClear,
			func() error { return s.StartServer(serverName) },
		)
	}
}

// stopAndClear 关停当前进程并清空状态（供 eula 重启流程回调）
func (s *ProcessManager) stopAndClear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		s.cmd = nil
		s.stdin = nil
	}
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

// handleLogLine 处理一行日志：推送给前端，并交给类型检测
// 由 stream 包的 StreamLines 逐行回调
func (s *ProcessManager) handleLogLine(line string) {
	runtime.EventsEmit(s.ctx, "server:log", line)
	s.handleDetectionLine(line)
}

// handleDetectionLine 分析一行服务器日志，执行类型检测
func (s *ProcessManager) handleDetectionLine(line string) {
	// 检测状态机为空时直接返回（无检测需求）
	if s.detector == nil {
		return
	}

	// 类型检测：若仍需类型检测，匹配类型关键字
	if s.detector.NeedType() {
		if serverType, ok := detector.DetectType(line); ok {
			// 命中类型关键字，写入 json 并标记完成；失败时推送错误，便于排查
			if err := s.cfg.SetServerType(s.detector.ServerName(), serverType); err != nil {
				runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[错误] 写入服务器类型失败: %v", err))
			} else {
				s.detector.MarkTypeDone()
				runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 已检测到服务器类型: %s", serverType))
				// 推送类型更新事件给前端（payload 为类型名）
				runtime.EventsEmit(s.ctx, "server:type", serverType)
			}
		}
	}
}

// detectVersion 执行版本检测（在独立 goroutine 中运行）
func (s *ProcessManager) detectVersion() {
	// 获取端口
	port, err := s.cfg.GetServerPort(s.detector.ServerName())
	if err != nil {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[错误] 获取端口失败: %v", err))
		return
	}
	if port == "" {
		runtime.EventsEmit(s.ctx, "server:log", "[错误] 未配置服务器端口，无法检测版本")
		return
	}

	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 开始检测服务器版本 (端口: %s)...", port))

	// 调用版本检测（会持续尝试直到成功）
	version, err := detector.GetVersionWithState(s.detector, port)
	if err != nil {
		runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[错误] 获取版本失败: %v", err))
		return
	}

	if version != "" {
		// 保存版本到配置
		if err := s.cfg.SetServerVersion(s.detector.ServerName(), version); err != nil {
			runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("[错误] 写入服务器版本失败: %v", err))
		} else {
			runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> 已检测到服务器版本: %s", version))
			// 推送版本更新事件给前端
			runtime.EventsEmit(s.ctx, "server:version", version)
		}
	}
}

// SendCommand 向运行中的服务器发送控制台命令
func (s *ProcessManager) SendCommand(command string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("服务器未在运行")
	}

	// 回显命令
	runtime.EventsEmit(s.ctx, "server:log", fmt.Sprintf("> %s", command))

	// 写入命令并换行
	if _, err := io.WriteString(s.stdin, command+"\n"); err != nil {
		return fmt.Errorf("发送命令失败: %v", err)
	}
	return nil
}

// StopServer 停止正在运行的服务器进程
func (s *ProcessManager) StopServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("服务器未在运行")
	}

	runtime.EventsEmit(s.ctx, "server:log", "> 正在停止服务器...")

	// 先尝试命令停止（发送 stop 命令给 Minecraft）
	if s.stdin != nil {
		_, _ = io.WriteString(s.stdin, "stop\n")
	}

	// 终止进程
	err := s.cmd.Process.Kill()
	s.cmd = nil
	s.stdin = nil
	// 清空检测状态机
	s.detector = nil
	// 清空启动时间（运行时长归零）
	s.startTime = time.Time{}

	if err != nil {
		return fmt.Errorf("停止服务器失败: %v", err)
	}
	runtime.EventsEmit(s.ctx, "server:log", "> 服务器已停止")
	return nil
}

// GetServerUptimeFor 返回指定服务器已运行秒数
// 仅当该服务器当前正在运行时返回实际时长，否则返回 0
func (s *ProcessManager) GetServerUptimeFor(serverName string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 进程未运行、不是目标服务器或未记录启动时间，返回 0
	if s.cmd == nil || s.cmd.Process == nil || s.currentServer != serverName || s.startTime.IsZero() {
		return 0
	}
	// 计算距启动时间的秒数
	return int(time.Since(s.startTime).Seconds())
}

// GetRunningPid 返回当前运行中的 Java 进程 PID
// 未运行时返回 0，供状态服务采集 JVM GC 统计使用
func (s *ProcessManager) GetRunningPid() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd == nil || s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}

// ShutdownAll 停止所有正在运行的服务器进程
// 在应用退出时调用，确保关闭 APP 时所有服务端都被终止
func (s *ProcessManager) ShutdownAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return
	}

	// 尝试优雅停止（发送 stop 命令）
	if s.stdin != nil {
		_, _ = io.WriteString(s.stdin, "stop\n")
	}

	// 终止进程
	_ = s.cmd.Process.Kill()
	s.cmd = nil
	s.stdin = nil
}
