package xray

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"

	"x-ui/config"
	"x-ui/logger"
	"x-ui/util/common"
)

func GetBinaryName() string {
	return fmt.Sprintf("xray-linux-%s", runtime.GOARCH)
}

func GetBinaryPath() string {
	return config.GetBinFolderPath() + "/" + GetBinaryName()
}

func GetConfigPath() string {
	return config.GetBinFolderPath() + "/config.json"
}

func GetGeositePath() string {
	return config.GetBinFolderPath() + "/geosite.dat"
}

func GetGeoipPath() string {
	return config.GetBinFolderPath() + "/geoip.dat"
}

func GetIPLimitLogPath() string {
	return config.GetLogFolder() + "/3xipl.log"
}

func GetIPLimitBannedLogPath() string {
	return config.GetLogFolder() + "/3xipl-banned.log"
}

func GetIPLimitBannedPrevLogPath() string {
	return config.GetLogFolder() + "/3xipl-banned.prev.log"
}

// IsXrayProcessRunning 检查系统中是否有 xray 进程在运行
// 这个函数用于检测"仅关闭面板"后遗留的 xray 进程
func IsXrayProcessRunning() bool {
	// 使用 pgrep 查找 xray 进程
	cmd := exec.Command("pgrep", "-f", "xray-linux.*config.json")
	err := cmd.Run()
	return err == nil
}

// KillXrayProcess 强制终止 xray 进程
func KillXrayProcess() error {
	cmd := exec.Command("pkill", "-f", "xray-linux.*config.json")
	return cmd.Run()
}

// GetApiPortFromConfig 从配置文件读取 API 端口
func GetApiPortFromConfig() int {
	configData, err := os.ReadFile(GetConfigPath())
	if err != nil {
		return 0
	}
	
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return 0
	}
	
	for _, inbound := range config.InboundConfigs {
		if inbound.Tag == "api" {
			return inbound.Port
		}
	}
	return 0
}

func GetAccessPersistentLogPath() string {
	return config.GetLogFolder() + "/3xipl-ap.log"
}

func GetAccessPersistentPrevLogPath() string {
	return config.GetLogFolder() + "/3xipl-ap.prev.log"
}

func GetAccessLogPath() (string, error) {
	config, err := os.ReadFile(GetConfigPath())
	if err != nil {
		logger.Warningf("Failed to read configuration file: %s", err)
		return "", err
	}

	jsonConfig := map[string]any{}
	err = json.Unmarshal([]byte(config), &jsonConfig)
	if err != nil {
		logger.Warningf("Failed to parse JSON configuration: %s", err)
		return "", err
	}

	if jsonConfig["log"] != nil {
		jsonLog := jsonConfig["log"].(map[string]any)
		if jsonLog["access"] != nil {
			accessLogPath := jsonLog["access"].(string)
			return accessLogPath, nil
		}
	}
	return "", err
}

func stopProcess(p *Process) {
	p.Stop()
}

type Process struct {
	*process
}

func NewProcess(xrayConfig *Config) *Process {
	p := &Process{newProcess(xrayConfig)}
	runtime.SetFinalizer(p, stopProcess)
	return p
}

type process struct {
	cmd *exec.Cmd

	version string
	apiPort int

	onlineClients []string
	mutex         sync.RWMutex

	config    *Config
	logWriter *LogWriter
	exitErr   error
	startTime time.Time
}

func newProcess(config *Config) *process {
	return &process{
		version:   "Unknown",
		config:    config,
		logWriter: NewLogWriter(),
		startTime: time.Now(),
	}
}

func (p *process) IsRunning() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	if p.cmd.ProcessState == nil {
		// 额外检查进程是否真实存在（防止僵尸进程）
		// 发送信号 0 不会杀死进程，只检查进程是否存在
		err := p.cmd.Process.Signal(syscall.Signal(0))
		return err == nil
	}
	return false
}

func (p *process) GetErr() error {
	return p.exitErr
}

func (p *process) GetResult() string {
	if len(p.logWriter.lastLine) == 0 && p.exitErr != nil {
		return p.exitErr.Error()
	}
	return p.logWriter.lastLine
}

func (p *process) GetVersion() string {
	return p.version
}

func (p *Process) GetAPIPort() int {
	return p.apiPort
}

func (p *Process) GetConfig() *Config {
	return p.config
}

func (p *Process) GetOnlineClients() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	clientsCopy := make([]string, len(p.onlineClients))
	copy(clientsCopy, p.onlineClients)
	return clientsCopy
}

func (p *Process) SetOnlineClients(clients []string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.onlineClients = clients
}

func (p *Process) GetUptime() uint64 {
	return uint64(time.Since(p.startTime).Seconds())
}

func (p *process) refreshAPIPort() {
	for _, inbound := range p.config.InboundConfigs {
		if inbound.Tag == "api" {
			p.apiPort = inbound.Port
			break
		}
	}
}

func (p *process) refreshVersion() {
	cmd := exec.Command(GetBinaryPath(), "-version")
	data, err := cmd.Output()
	if err != nil {
		p.version = "Unknown"
	} else {
		datas := bytes.Split(data, []byte(" "))
		if len(datas) <= 1 {
			p.version = "Unknown"
		} else {
			p.version = string(datas[1])
		}
	}
}

func (p *process) Start() (err error) {
	if p.IsRunning() {
		return errors.New("xray is already running")
	}

	defer func() {
		if err != nil {
			logger.Error("Failure in running xray-core process: ", err)
			p.exitErr = err
		}
	}()

	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		return common.NewErrorf("Failed to generate XRAY configuration files: %v", err)
	}

	err = os.MkdirAll(config.GetLogFolder(), 0o770)
	if err != nil {
		logger.Warningf("Failed to create log folder: %s", err)
	}

	configPath := GetConfigPath()
	err = os.WriteFile(configPath, data, fs.ModePerm)
	if err != nil {
		return common.NewErrorf("Failed to write configuration file: %v", err)
	}

	cmd := exec.Command(GetBinaryPath(), "-c", configPath)
	p.cmd = cmd

	// 设置Xray进程在独立的进程组中运行，避免父进程退出时被一起杀死
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 创建新的进程组
	}

	cmd.Stdout = p.logWriter
	cmd.Stderr = p.logWriter

	go func() {
		err := cmd.Run()
		if err != nil {
			logger.Error("Failure in running xray-core:", err)
			p.exitErr = err
		}
	}()

	p.refreshVersion()
	p.refreshAPIPort()

	return nil
}

func (p *process) Stop() error {
	if !p.IsRunning() {
		return errors.New("xray is not running")
	}
	
	return p.cmd.Process.Signal(syscall.SIGTERM)
}

func writeCrashReport(m []byte) error {
	crashReportPath := config.GetBinFolderPath() + "/core_crash_" + time.Now().Format("20060102_150405") + ".log"
	return os.WriteFile(crashReportPath, m, os.ModePerm)
}
