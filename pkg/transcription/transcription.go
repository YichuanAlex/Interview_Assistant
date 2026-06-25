package transcription

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"Interview_Assistant/pkg/logger"
)

// Result 表示一条实时转录结果
type Result struct {
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
}

// TranscriptionService 管理 faster-whisper 子进程
type TranscriptionService struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	running  bool
	emitFunc func(eventName string, data ...interface{})

	scriptPath string
	pythonPath string
}

// NewTranscriptionService 创建转录服务
func NewTranscriptionService(emitFunc func(eventName string, data ...interface{})) *TranscriptionService {
	return &TranscriptionService{
		emitFunc:   emitFunc,
		pythonPath: findPython(),
		scriptPath: findScript(),
	}
}

// IsRunning 返回是否正在转录
func (s *TranscriptionService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Start 启动实时转录
func (s *TranscriptionService) Start(device int, model string, language string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("转录已经在运行")
	}

	if s.pythonPath == "" {
		return fmt.Errorf("未找到 python3，请确认已安装 Python 环境")
	}
	if s.scriptPath == "" {
		return fmt.Errorf("未找到 realtime_transcribe.py 脚本")
	}

	if model == "" {
		model = "./models/small"
	}
	if language == "" {
		language = "zh"
	}

	args := []string{
		s.scriptPath,
		"--model", model,
		"--device", fmt.Sprintf("%d", device),
		"--language", language,
		"--json-output",
	}

	cmd := exec.Command(s.pythonPath, args...)
	cmd.Dir = filepath.Dir(s.scriptPath)
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建 stdout pipe 失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr pipe 失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动转录进程失败: %w", err)
	}

	s.cmd = cmd
	s.running = true

	go s.readStdout(stdout)
	go s.readStderr(stderr)
	go s.waitExit()

	logger.Printf("实时转录已启动: device=%d model=%s language=%s\n", device, model, language)
	if s.emitFunc != nil {
		s.emitFunc("transcription-status", "started")
	}
	return nil
}

// Stop 停止实时转录
func (s *TranscriptionService) Stop() error {
	s.mu.Lock()
	cmd := s.cmd
	s.running = false
	s.cmd = nil
	s.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(os.Interrupt)
		// 等待进程退出，最多 3 秒
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
		}
	}

	logger.Println("实时转录已停止")
	if s.emitFunc != nil {
		s.emitFunc("transcription-status", "stopped")
	}
	return nil
}

func (s *TranscriptionService) readStdout(stdout io.ReadCloser) {
	defer stdout.Close()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var res Result
		if err := json.Unmarshal([]byte(line), &res); err == nil && res.Text != "" {
			if s.emitFunc != nil {
				s.emitFunc("transcription", res.Timestamp, res.Text)
			}
		} else {
			logger.Printf("转录输出: %s\n", line)
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("读取转录输出失败: %v\n", err)
	}
}

func (s *TranscriptionService) readStderr(stderr io.ReadCloser) {
	defer stderr.Close()

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		logger.Printf("转录 stderr: %s\n", scanner.Text())
	}
}

func (s *TranscriptionService) waitExit() {
	s.mu.Lock()
	cmd := s.cmd
	s.mu.Unlock()
	if cmd == nil {
		return
	}

	err := cmd.Wait()
	s.mu.Lock()
	wasRunning := s.running
	s.running = false
	s.cmd = nil
	s.mu.Unlock()

	if err != nil && wasRunning {
		logger.Printf("转录进程退出: %v\n", err)
		if s.emitFunc != nil {
			s.emitFunc("transcription-status", "error", err.Error())
		}
	} else if s.emitFunc != nil {
		s.emitFunc("transcription-status", "stopped")
	}
}

func findPython() string {
	candidates := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		candidates = []string{"python.exe", "python3.exe"}
	}
	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func findScript() string {
	// 1. 相对于当前工作目录（开发模式）
	cwd, err := os.Getwd()
	if err == nil {
		path := filepath.Join(cwd, "BUZZ", "faster-whisper", "realtime_transcribe.py")
		if fileExists(path) {
			return path
		}
	}

	// 2. 相对于可执行文件（生产模式 .app/Contents/MacOS/..）
	ex, err := os.Executable()
	if err == nil {
		exDir := filepath.Dir(ex)
		// macOS: Interview_Assistant.app/Contents/MacOS/ -> ../Resources/BUZZ/...
		path := filepath.Join(exDir, "..", "Resources", "BUZZ", "faster-whisper", "realtime_transcribe.py")
		if fileExists(path) {
			return path
		}
		// Windows/Linux: 与可执行文件同级
		path = filepath.Join(exDir, "BUZZ", "faster-whisper", "realtime_transcribe.py")
		if fileExists(path) {
			return path
		}
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
