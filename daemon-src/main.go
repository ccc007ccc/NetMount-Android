package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// --- 数据结构 ---

type MountConfig struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Remote     string            `json:"remote"`
	MountPoint string            `json:"mountPoint"`
	Parameters map[string]string `json:"parameters"`
	AuthType   string            `json:"authType,omitempty"`
	User       string            `json:"user,omitempty"`
	Pass       string            `json:"pass,omitempty"`
}

type AppConfig struct {
	Mounts []MountConfig `json:"mounts"`
}

type MountStatus struct {
	Name      string `json:"name"`
	IsMounted bool   `json:"isMounted"`
	Error     string `json:"error,omitempty"`
	Pid       int    `json:"pid,omitempty"`
}

// --- 全局变量 ---

var (
	config        AppConfig
	mountStatus   = make(map[string]*MountStatus)
	configLock    = &sync.RWMutex{}
	configPath    string // 将在 main 函数中通过 flag 设置
	rclonePath    string
	logStore      []string
	logStoreLock  = &sync.RWMutex{}
	maxLogEntries = 500 // 增加日志存储上限
)

// --- 日志系统 ---

// obscurePassword 使用 rclone obscure 命令模糊化密码
func obscurePassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	
	cmd := exec.Command(rclonePath, "obscure", password)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("rclone obscure 命令执行失败: %v", err)
	}
	
	obscured := strings.TrimSpace(string(output))
	if obscured == "" {
		return "", fmt.Errorf("rclone obscure 返回空结果")
	}
	
	return obscured, nil
}

// --- 日志系统 ---

// addLog 添加一条带时间戳的日志到内存存储区
func addLog(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullMessage := fmt.Sprintf("%s - %s", timestamp, message)

	// 同时输出到标准日志，方便调试
	log.Println(fullMessage)

	logStoreLock.Lock()
	defer logStoreLock.Unlock()
	logStore = append(logStore, fullMessage)
	if len(logStore) > maxLogEntries {
		// 保持日志列表大小
		logStore = logStore[len(logStore)-maxLogEntries:]
	}
}

// --- 核心逻辑 ---

func createRcloneConfig() {
	rcloneConfigPath := filepath.Join(filepath.Dir(configPath), "rclone.conf")
	
	// 检查配置文件是否已存在
	if _, err := os.Stat(rcloneConfigPath); err == nil {
		addLog("rclone 配置文件已存在: %s", rcloneConfigPath)
		return
	}
	
	addLog("创建 rclone 配置文件: %s", rcloneConfigPath)
	
	// 创建基本的 rclone 配置
	configContent := `# rclone 配置文件 - 由 NetMount 自动生成
# 这个文件用于存储 rclone 的基本配置

`
	
	// 确保配置目录存在
	configDir := filepath.Dir(rcloneConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		addLog("错误: 创建配置目录失败: %v", err)
		return
	}
	
	// 写入配置文件
	if err := ioutil.WriteFile(rcloneConfigPath, []byte(configContent), 0644); err != nil {
		addLog("错误: 创建 rclone 配置文件失败: %v", err)
		return
	}
	
	addLog("rclone 配置文件创建成功")
}

// createSMBConfig 为 SMB 挂载创建临时配置段，返回配置段名称和share路径
func createSMBConfig(mount MountConfig) (string, string) {
	rcloneConfigPath := filepath.Join(filepath.Dir(configPath), "rclone.conf")
	
	// 生成唯一的配置段名称
	sectionName := fmt.Sprintf("smb_%s", mount.Name)
	
	// 解析远程地址
	parts := strings.SplitN(mount.Remote, "/", 2)
	host := parts[0]
	sharePath := ""
	if len(parts) > 1 && parts[1] != "" {
		sharePath = parts[1]
	}
	
	addLog("📋 SMB地址解析: 原始='%s' -> host='%s', share路径='%s'", mount.Remote, host, sharePath)
	
	// 使用 rclone 密码模糊化命令处理密码
	obscuredPass, err := obscurePassword(mount.Pass)
	if err != nil {
		addLog("🔐 密码模糊化失败，使用明文密码: %v", err)
		obscuredPass = mount.Pass
	} else {
		addLog("🔐 使用模糊化密码")
	}
	
	// 对于SMB，总是创建基础配置（不指定share），然后在挂载时指定路径
	configSection := fmt.Sprintf(`
[%s]
type = smb
host = %s
user = %s
pass = %s
`, sectionName, host, mount.User, obscuredPass)
	
	if sharePath != "" {
		addLog("📋 将在挂载时直接访问: %s/%s", host, sharePath)
	} else {
		addLog("📋 将挂载SMB根目录")
	}
	
	// 读取现有配置
	existingContent := ""
	if data, err := ioutil.ReadFile(rcloneConfigPath); err == nil {
		existingContent = string(data)
	}
	
	// 检查是否已存在该配置段，如果存在则替换
	startMarker := fmt.Sprintf("[%s]", sectionName)
	if strings.Contains(existingContent, startMarker) {
		// 删除旧的配置段
		lines := strings.Split(existingContent, "\n")
		var newLines []string
		inSection := false
		
		for _, line := range lines {
			if strings.TrimSpace(line) == startMarker {
				inSection = true
				continue
			}
			if inSection && strings.HasPrefix(line, "[") && line != startMarker {
				inSection = false
			}
			if !inSection {
				newLines = append(newLines, line)
			}
		}
		existingContent = strings.Join(newLines, "\n")
		addLog("已删除旧的配置段 '%s'", sectionName)
	}
	
	// 追加新的配置段
	newContent := existingContent + configSection
	
	if err := ioutil.WriteFile(rcloneConfigPath, []byte(newContent), 0644); err != nil {
		addLog("错误: 写入 rclone 配置失败: %v", err)
		return "", ""
	}
	
	addLog("已创建 rclone SMB 配置段: %s", sectionName)
	addLog("📋 配置内容预览:\n%s", configSection)
	
	// 简单测试配置是否能连接
	addLog("🔍 测试SMB配置连接...")
	testCmd := exec.Command(rclonePath, "lsd", fmt.Sprintf("%s:", sectionName), "--config", rcloneConfigPath, "--timeout", "5s")
	if output, err := testCmd.CombinedOutput(); err != nil {
		addLog("⚠️  配置测试失败: %v", err)
		addLog("📋 测试输出: %s", string(output))
	} else {
		addLog("✅ 配置测试成功，可以访问共享")
	}
	
	return sectionName, sharePath
}

func loadConfig() {
	configLock.Lock()
	defer configLock.Unlock()
	addLog("正在从 %s 加载配置...", configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		addLog("配置文件 %s 不存在，将创建一个新的。", configPath)
		config = AppConfig{Mounts: []MountConfig{}}
		saveConfig_nolock()
	} else {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			addLog("致命错误: 读取配置文件失败: %v", err)
			log.Fatalf("读取配置文件失败: %v", err)
		}
		if err := json.Unmarshal(data, &config); err != nil {
			addLog("致命错误: 解析配置文件失败: %v", err)
			log.Fatalf("解析配置文件失败: %v", err)
		}
	}

	activeMounts := make(map[string]bool)
	for _, mount := range config.Mounts {
		activeMounts[mount.Name] = true
		if _, exists := mountStatus[mount.Name]; !exists {
			mountStatus[mount.Name] = &MountStatus{Name: mount.Name, IsMounted: false}
		}
	}
	for name := range mountStatus {
		if !activeMounts[name] {
			if status := mountStatus[name]; status.IsMounted {
				addLog("配置中移除了挂载点 '%s'，尝试卸载...", name)
				unmountSingle_nolock(name)
			}
			delete(mountStatus, name)
		}
	}
	addLog("配置加载完成。")
}

func saveConfig_nolock() {
	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		addLog("错误: 创建配置目录 '%s' 失败: %v", configDir, err)
		return
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		addLog("错误: 序列化配置失败: %v", err)
		return
	}
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		addLog("错误: 写入配置文件 '%s' 失败: %v", configPath, err)
	}
}

// isDeviceUnlocked 检测Android设备是否已解锁(用户数据可访问)
func isDeviceUnlocked() bool {
	// 尝试在用户数据目录创建测试文件
	testPaths := []string{
		"/data/media/0",
		"/storage/emulated/0", 
		"/sdcard",
	}
	
	for _, path := range testPaths {
		// 先检查目录是否存在
		if _, err := os.Stat(path); err != nil {
			continue // 目录不存在，尝试下一个
		}
		
		// 尝试创建测试文件来验证写入权限
		testFile := filepath.Join(path, ".netmount_unlock_test")
		file, err := os.Create(testFile)
		if err != nil {
			// 检查具体错误类型
			if strings.Contains(err.Error(), "required key not available") {
				addLog("🔒 检测到加密错误，设备未完全解锁")
				return false
			}
			continue // 其他错误，尝试下一个路径
		}
		
		// 成功创建文件，立即清理
		file.Close()
		os.Remove(testFile)
		addLog("📱 检测到设备已解锁，路径 %s 可写入", path)
		return true
	}
	
	addLog("🔒 所有用户数据路径均不可访问，设备未解锁")
	return false
}

// waitForDeviceUnlock 等待设备解锁
func waitForDeviceUnlock() {
	addLog("🔒 设备未解锁，开始等待用户解锁设备...")
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()
	
	for range ticker.C {
		if isDeviceUnlocked() {
			addLog("🔓 设备已解锁！用户数据区域现在可以访问")
			return
		}
		addLog("⏱️  继续等待设备解锁...")
	}
}

func mountSingle(mount MountConfig) {
	configLock.Lock()
	status := mountStatus[mount.Name]
	if status.IsMounted {
		addLog("挂载点 '%s' 已经挂载，跳过。", mount.Name)
		configLock.Unlock()
		return
	}
	configLock.Unlock()

	addLog("=== 开始挂载 '%s' ===", mount.Name)
	addLog("挂载类型: %s", mount.Type)
	addLog("远程地址: %s", mount.Remote)
	addLog("挂载路径: %s", mount.MountPoint)
	addLog("认证类型: %s", mount.AuthType)

	// 如果挂载路径在用户数据区域，等待设备解锁
	if strings.HasPrefix(mount.MountPoint, "/data/media/") || 
	   strings.HasPrefix(mount.MountPoint, "/storage/emulated/") ||
	   strings.HasPrefix(mount.MountPoint, "/sdcard/") {
		addLog("🔒 挂载路径位于用户数据区域，等待设备解锁...")
		waitForDeviceUnlock()
	}

	// 创建挂载点目录
	if err := os.MkdirAll(mount.MountPoint, 0755); err != nil {
		errMsg := fmt.Sprintf("创建挂载点目录 '%s' 失败: %v", mount.MountPoint, err)
		addLog("❌ 错误: %s", errMsg)
		configLock.Lock()
		status.Error = errMsg
		configLock.Unlock()
		return
	}

	// 根据类型构建正确的 remote 路径
	var remoteAddr string
	var useConfigSection bool
	rcloneConfigPath := filepath.Join(filepath.Dir(configPath), "rclone.conf")
	
	switch mount.Type {
	case "smb":
		// 为 SMB 创建配置段并使用配置文件方式
		sectionName, sharePath := createSMBConfig(mount)
		if sectionName == "" {
			errMsg := fmt.Sprintf("创建 SMB 配置段失败 '%s'", mount.Name)
			addLog("❌ 错误: %s", errMsg)
			configLock.Lock()
			status.Error = errMsg
			configLock.Unlock()
			return
		}
		if sharePath != "" {
			remoteAddr = fmt.Sprintf("%s:%s", sectionName, sharePath)
			addLog("📂 挂载路径: %s (指向共享: %s)", remoteAddr, sharePath)
		} else {
			remoteAddr = fmt.Sprintf("%s:", sectionName)
			addLog("📂 挂载路径: %s (根目录)", remoteAddr)
		}
		useConfigSection = true
	case "ftp":
		if !strings.HasPrefix(mount.Remote, "ftp://") {
			remoteAddr = fmt.Sprintf("ftp://%s", mount.Remote)
		} else {
			remoteAddr = mount.Remote
		}
		useConfigSection = false
	case "webdav":
		if !strings.HasPrefix(mount.Remote, "http://") && !strings.HasPrefix(mount.Remote, "https://") {
			remoteAddr = fmt.Sprintf("http://%s", mount.Remote)
		} else {
			remoteAddr = mount.Remote
		}
		useConfigSection = false
	default:
		remoteAddr = mount.Remote
		useConfigSection = true
	}

	var args []string
	if useConfigSection {
		// 使用配置文件中的命名段
		args = []string{
			"mount",
			remoteAddr,
			mount.MountPoint,
			"--config", rcloneConfigPath,
			"--allow-other",
			"--allow-non-empty", // 允许挂载到非空目录
			"--log-level", "INFO",
			"--uid", "0",
			"--gid", "9997",
			"--umask", "0007",
			"--vfs-cache-mode", "writes",
		}
	} else {
		// 直接使用URL，避免配置文件相关的问题，使用基础兼容参数
		args = []string{
			"mount",
			remoteAddr,
			mount.MountPoint,
			"--allow-other",
			"--allow-non-empty", // 允许挂载到非空目录
			"--log-level", "INFO",
			"--uid", "0",
			"--gid", "9997",
			"--umask", "0007",
			"--vfs-cache-mode", "writes",
			"--no-check-certificate",
		}
	}

	// 准备环境变量
	env := os.Environ()
	env = append(env, "HOME=/data/adb/netmount") // 设置 HOME 目录避免配置目录查找问题

	// 根据类型和认证类型添加认证参数
	switch mount.Type {
	case "webdav":
		if mount.AuthType == "password" {
			args = append(args, "--webdav-user", mount.User, "--webdav-pass", mount.Pass)
		}
	case "ftp":
		if mount.AuthType == "password" {
			args = append(args, "--ftp-user", mount.User, "--ftp-pass", mount.Pass)
		} else if mount.AuthType == "anonymous" {
			args = append(args, "--ftp-user", "anonymous", "--ftp-pass", "rclone@rclone.org")
		}
	case "smb":
		// SMB 认证已在配置文件中处理
	}

	// 启动挂载命令
	addLog("🚀 启动挂载: %s -> %s", mount.Remote, mount.MountPoint)

	cmd := exec.Command(rclonePath, args...)
	
	// 设置环境变量（已在前面准备好）
	cmd.Env = env
	
	// 为跨平台兼容性设置进程属性
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	setProcessGroupID(cmd.SysProcAttr)

	// 捕获 stdout 和 stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("启动 rclone 进程失败 '%s': %v", mount.Name, err)
		addLog("❌ 错误: %s", errMsg)
		configLock.Lock()
		status.IsMounted = false
		status.Error = errMsg
		configLock.Unlock()
		return
	}

	addLog("✅ 成功启动 '%s' 的挂载进程, PID: %d", mount.Name, cmd.Process.Pid)
	configLock.Lock()
	status.IsMounted = true
	status.Error = ""
	status.Pid = cmd.Process.Pid
	configLock.Unlock()

	// 等待一段时间后检查挂载点是否有内容
	go func() {
		time.Sleep(5 * time.Second) // 等待5秒让挂载稳定
		
		// 检查挂载点是否真的有内容
		files, err := os.ReadDir(mount.MountPoint)
		if err != nil {
			addLog("⚠️  警告: 无法读取挂载点 '%s': %v", mount.MountPoint, err)
		} else {
			if len(files) == 0 {
				addLog("📂 挂载点 '%s' 为空，可能的原因:", mount.MountPoint)
				addLog("   1. 网络连接问题")
				addLog("   2. 认证失败")
				addLog("   3. 远程路径不存在")
				addLog("   4. SMB 共享路径错误")
			} else {
				addLog("✅ 挂载验证成功: '%s' 包含 %d 个项目", mount.MountPoint, len(files))
			}
		}
	}()

	// 实时读取并优化日志输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// 过滤和美化 rclone 输出
			if strings.Contains(line, "INFO") {
				addLog("ℹ️  [%s] %s", mount.Name, line)
			} else if strings.Contains(line, "NOTICE") {
				addLog("📢 [%s] %s", mount.Name, line)
			} else {
				addLog("📝 [%s] %s", mount.Name, line)
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// 过滤和美化错误输出
			if strings.Contains(line, "CRITICAL") || strings.Contains(line, "ERROR") {
				addLog("❌ [%s] %s", mount.Name, line)
			} else if strings.Contains(line, "WARNING") {
				addLog("⚠️  [%s] %s", mount.Name, line)
			} else if strings.Contains(line, "INFO") {
				addLog("ℹ️  [%s] %s", mount.Name, line)
			} else {
				addLog("🔍 [%s] %s", mount.Name, line)
			}
		}
	}()

	go func() {
		err := cmd.Wait()
		configLock.Lock()
		if s, ok := mountStatus[mount.Name]; ok && s.Pid == cmd.Process.Pid {
			s.IsMounted = false
			s.Pid = 0
			if err != nil {
				addLog("❌ 挂载进程 '%s' (PID: %d) 异常退出: %v", mount.Name, cmd.Process.Pid, err)
			} else {
				addLog("ℹ️  挂载进程 '%s' (PID: %d) 正常退出", mount.Name, cmd.Process.Pid)
			}
		}
		configLock.Unlock()
	}()
}

func unmountSingle_nolock(name string) error {
	status, ok := mountStatus[name]
	if !ok || !status.IsMounted || status.Pid == 0 {
		errMsg := fmt.Sprintf("无法卸载 '%s': 未挂载或 PID 未知。", name)
		addLog("警告: %s", errMsg)
		return fmt.Errorf(errMsg)
	}

	addLog("正在尝试卸载 '%s' (PID: %d)...", name, status.Pid)
	process, err := os.FindProcess(status.Pid)
	if err != nil {
		addLog("找不到进程 (PID: %d): %v", status.Pid, err)
	} else {
		if err := process.Signal(syscall.SIGTERM); err != nil {
			addLog("向 PID %d 发送 SIGTERM 失败: %v。尝试强制终止...", status.Pid, err)
			process.Kill()
		}
	}

	status.IsMounted = false
	status.Pid = 0
	status.Error = ""
	addLog("卸载信号已发送到 '%s'。", name)
	return nil
}

// --- HTTP 处理函数 ---

func logHandler(w http.ResponseWriter, r *http.Request) {
	logStoreLock.RLock()
	defer logStoreLock.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logStore)
}

// (其他 handler 保持不变, 但需要用 addLog 替换 log.Printf)
func configHandler(w http.ResponseWriter, r *http.Request) {
	configLock.Lock()
	defer configLock.Unlock()
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	case "POST":
		var newConfig AppConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		config = newConfig
		saveConfig_nolock()
		go loadConfig()
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	configLock.RLock()
	defer configLock.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mountStatus)
}

func mountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
		return
	}
	name := filepath.Base(r.URL.Path)
	configLock.RLock()
	var mountToRun *MountConfig
	for i := range config.Mounts {
		if config.Mounts[i].Name == name {
			mountToRun = &config.Mounts[i]
			break
		}
	}
	configLock.RUnlock()
	if mountToRun == nil {
		http.Error(w, "未找到指定的挂载配置", http.StatusNotFound)
		return
	}
	go mountSingle(*mountToRun)
	w.WriteHeader(http.StatusAccepted)
}

func unmountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
		return
	}
	name := filepath.Base(r.URL.Path)
	configLock.Lock()
	defer configLock.Unlock()
	if err := unmountSingle_nolock(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}


func precheckHandler(w http.ResponseWriter, r *http.Request) {
	precheckLogPath := filepath.Join(filepath.Dir(configPath), "precheck.log")
	data, err := ioutil.ReadFile(precheckLogPath)
	if err != nil {
		http.Error(w, "无法读取环境检查日志: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

// --- main 函数 ---

func main() {
	// 设置时区，避免时间显示错误
	os.Setenv("TZ", "CST-8") // 使用 POSIX 时区格式 CST-8 (中国标准时间)
	
	// 定义命令行 flag
	defaultConfigPath := "/data/adb/netmount/config.json"
	configPathPtr := flag.String("config", defaultConfigPath, "Path to the configuration file.")
	flag.Parse()
	configPath = *configPathPtr

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取可执行文件路径: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	rclonePath = filepath.Join(exeDir, "rclone")

	addLog("--- NetMount 守护进程启动 ---")
	addLog("使用配置文件: %s", configPath)
	
	// 创建 rclone 配置文件（如果不存在）
	createRcloneConfig()
	
	loadConfig()

	webrootDir := filepath.Join(exeDir, "webroot")
	fs := http.FileServer(http.Dir(webrootDir))
	http.Handle("/", fs)
	http.HandleFunc("/api/config", configHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/api/logs", logHandler)
	http.HandleFunc("/api/precheck", precheckHandler) // 新增: 环境检查 API
	http.HandleFunc("/api/mount/", mountHandler)
	http.HandleFunc("/api/unmount/", unmountHandler)

	go func() {
		time.Sleep(2 * time.Second) // 稍微延迟，等待网络稳定
		configLock.RLock()
		defer configLock.RUnlock()
		
		addLog("🚀 开始智能挂载策略...")
		
		// 分类挂载点：需要设备解锁的和不需要的
		var needUnlockMounts []MountConfig
		var normalMounts []MountConfig
		
		for _, mount := range config.Mounts {
			if strings.HasPrefix(mount.MountPoint, "/data/media/") || 
			   strings.HasPrefix(mount.MountPoint, "/storage/emulated/") ||
			   strings.HasPrefix(mount.MountPoint, "/sdcard/") {
				needUnlockMounts = append(needUnlockMounts, mount)
			} else {
				normalMounts = append(normalMounts, mount)
			}
		}
		
		// 先挂载不需要解锁的
		if len(normalMounts) > 0 {
			addLog("📁 挂载系统区域挂载点 (%d个)...", len(normalMounts))
			for _, mount := range normalMounts {
				go mountSingle(mount)
			}
		}
		
		// 对于需要解锁的挂载点，启动单独的协程处理
		if len(needUnlockMounts) > 0 {
			go func() {
				addLog("🔒 检测到需要设备解锁的挂载点 (%d个)", len(needUnlockMounts))
				addLog("📱 开始挂载用户数据区域...")
				for _, mount := range needUnlockMounts {
					go mountSingle(mount)
				}
			}()
		}
	}()

	addLog("服务器正在监听 :8088...")
	if err := http.ListenAndServe(":8088", nil); err != nil {
		addLog("致命错误: 服务器启动失败: %v", err)
		log.Fatal(err)
	}
}
