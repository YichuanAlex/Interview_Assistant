package ocr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"Interview_Assistant/pkg/logger"
)

// Service 提供跨平台的 OCR 能力
type Service struct{}

func NewService() *Service {
	return &Service{}
}

// Recognize 对给定图片路径执行 OCR，返回识别到的文本
func (s *Service) Recognize(imagePath string) (string, error) {
	if imagePath == "" {
		return "", fmt.Errorf("图片路径为空")
	}
	if _, err := os.Stat(imagePath); err != nil {
		return "", fmt.Errorf("图片不存在: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return s.recognizeMac(imagePath)
	case "windows":
		return s.recognizeWindows(imagePath)
	default:
		return "", fmt.Errorf("当前平台 %s 暂不支持 OCR", runtime.GOOS)
	}
}

// recognizeMac 使用 macOS Vision 框架识别文字
func (s *Service) recognizeMac(imagePath string) (string, error) {
	scriptPath := filepath.Join(getProjectRoot(), "scripts", "ocr_mac.swift")
	if _, err := os.Stat(scriptPath); err != nil {
		return "", fmt.Errorf("找不到 OCR 脚本: %s", scriptPath)
	}

	cmd := exec.Command("swift", scriptPath, imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("macOS OCR 失败: %v, output: %s", err, string(output))
		return "", fmt.Errorf("OCR 识别失败: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// recognizeWindows 使用 Wechat_OCR 识别文字
func (s *Service) recognizeWindows(imagePath string) (string, error) {
	ocrDir := filepath.Join(getProjectRoot(), "Wechat_OCR")
	script := filepath.Join(ocrDir, "ocr_wrapper.py")

	// 如果 wrapper 不存在，创建一个最小 wrapper
	if _, err := os.Stat(script); os.IsNotExist(err) {
		wrapper := `import sys
import os
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
import OCR

texts = OCR.wechat_ocr(sys.argv[1])
print("\n".join(texts))
`
		if err := os.WriteFile(script, []byte(wrapper), 0644); err != nil {
			return "", fmt.Errorf("创建 OCR wrapper 失败: %w", err)
		}
	}

	cmd := exec.Command("python", script, imagePath)
	cmd.Dir = ocrDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Windows Wechat_OCR 失败: %v, output: %s", err, string(output))
		return "", fmt.Errorf("OCR 识别失败: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getProjectRoot 尝试定位项目根目录
func getProjectRoot() string {
	// 1. 通过环境变量
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}

	// 2. 通过可执行文件路径向上查找 go.mod
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for i := 0; i < 6; i++ {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// 3. 当前工作目录
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	return "."
}
