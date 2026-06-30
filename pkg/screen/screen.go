package screen

import (
	imageutil "Interview_Assistant/pkg/ImageUtil"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"

	"github.com/kbinani/screenshot"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type PreviewResult struct {
	ImgBytes []byte `json:"imgBytes"`
	Base64   string `json:"base64"`
	Size     string `json:"size"`
}

type Service struct {
	ctx context.Context
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
}

// CapturePreview 获取当前截图的预览（Base64）
func (s *Service) CapturePreview(quality int, sharpen float64, grayscale bool, noCompression bool, mode string) (PreviewResult, error) {
	var x, y, w, h int

	if mode == "fullscreen" {
		// 全屏模式：获取主屏幕尺寸
		bounds := screenshot.GetDisplayBounds(0)
		x, y, w, h = bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	} else {
		// 窗口模式：获取当前窗口位置和大小
		if s.ctx == nil {
			return PreviewResult{}, fmt.Errorf("context not initialized")
		}
		x, y = wailsruntime.WindowGetPosition(s.ctx)
		w, h = wailsruntime.WindowGetSize(s.ctx)
	}

	// 截图
	img, err := screenshot.Capture(x, y, w, h)
	if err != nil {
		return PreviewResult{}, fmt.Errorf("截图失败: %v", err)
	}

	return encodePreview(img, quality, sharpen, grayscale, noCompression)
}

// CaptureInteractiveSelection 调用系统框选截图：按下后拖拽选区，松手生成截图。
func (s *Service) CaptureInteractiveSelection(quality int, sharpen float64, grayscale bool, noCompression bool) (PreviewResult, error) {
	if _, err := exec.LookPath("screencapture"); err != nil {
		return PreviewResult{}, fmt.Errorf("当前系统不支持交互式框选截图")
	}

	tmp, err := os.CreateTemp("", "ia_selection_*.png")
	if err != nil {
		return PreviewResult{}, fmt.Errorf("创建临时截图文件失败: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("screencapture", "-i", "-s", "-x", "-t", "png", tmpPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		if _, statErr := os.Stat(tmpPath); statErr != nil {
			return PreviewResult{}, fmt.Errorf("截图已取消")
		}
		return PreviewResult{}, fmt.Errorf("框选截图失败: %v %s", err, string(output))
	}

	file, err := os.Open(tmpPath)
	if err != nil {
		return PreviewResult{}, fmt.Errorf("读取截图失败: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return PreviewResult{}, fmt.Errorf("解析截图失败: %w", err)
	}

	return encodePreview(img, quality, sharpen, grayscale, noCompression)
}

func encodePreview(img image.Image, quality int, sharpen float64, grayscale bool, noCompression bool) (PreviewResult, error) {
	var imgBytes []byte
	var ImageBase64 string
	if noCompression {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return PreviewResult{}, fmt.Errorf("图片编码失败: %v", err)
		}
		imgBytes = buf.Bytes()
		ImageBase64 = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(imgBytes))
	} else {
		var err error
		imgBytes, err = imageutil.CompressForOCR(img, quality, sharpen, grayscale)
		if err != nil {
			return PreviewResult{}, fmt.Errorf("图片处理失败: %v", err)
		}
		ImageBase64 = fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(imgBytes))
	}

	// 计算大小
	sizeKB := float64(len(ImageBase64)) / 1024.0
	sizeStr := fmt.Sprintf("%.2f KB", sizeKB)

	// 转 Base64
	return PreviewResult{
		ImgBytes: imgBytes,
		Base64:   ImageBase64,
		Size:     sizeStr,
	}, nil
}
