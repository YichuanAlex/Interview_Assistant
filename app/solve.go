package app

import (
	"Interview_Assistant/pkg/logger"
	"Interview_Assistant/pkg/solution"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const MaxScreenshots = 3

var screenshotBuffer []string
var ocrTextBuffer string

func (a *App) TriggerScreenshot() {
	cfg := a.configManager.Get()

	if cfg.APIKey == "" {
		a.EmitEvent("require-api-key")
		return
	}

	if cfg.Model == "" {
		a.EmitEvent("toast", "请先选择模型")
		a.EmitEvent("open-settings", "model")
		return
	}

	if a.taskManager.HasRunningTask() {
		logger.Println("忽略截图：当前有任务正在运行")
		a.EmitEvent("toast", "正在处理中，请稍候...")
		return
	}

	if len(screenshotBuffer) >= MaxScreenshots {
		a.EmitEvent("toast", "最多截图 3 张图片，请先发送或删除")
		return
	}

	previewResult, err := a.GetScreenshotPreview(
		cfg.CompressionQuality,
		cfg.Sharpening,
		cfg.Grayscale,
		cfg.NoCompression,
		cfg.ScreenshotMode,
	)
	if err != nil {
		logger.Printf("截图失败: %v\n", err)
		a.EmitEvent("toast", "截图失败: "+err.Error())
		return
	}

	screenshotBuffer = append(screenshotBuffer, previewResult.Base64)

	// 异步 OCR：将截图保存为临时文件并识别文字
	go func(b64 string, bufIdx int) {
		tmpPath, err := saveBase64ToTemp(b64)
		if err != nil {
			logger.Printf("保存截图临时文件失败: %v", err)
			return
		}
		defer os.Remove(tmpPath)

		text, err := a.ocrService.Recognize(tmpPath)
		if err != nil {
			logger.Printf("OCR 识别失败: %v", err)
			return
		}

		// 只保留最近一次截图的 OCR 文本到编辑缓冲
		ocrTextBuffer = text
		a.EmitEvent("ocr-text", text, bufIdx)
	}(previewResult.Base64, len(screenshotBuffer)-1)

	a.EmitEvent("screenshot-taken", previewResult.Base64, len(screenshotBuffer))
}

func (a *App) RemoveScreenshot(index int) {
	if index < 0 || index >= len(screenshotBuffer) {
		return
	}
	screenshotBuffer = append(screenshotBuffer[:index], screenshotBuffer[index+1:]...)
	a.EmitEvent("screenshot-removed", index, len(screenshotBuffer))
}

func (a *App) RemoveLastScreenshot() {
	if len(screenshotBuffer) == 0 {
		return
	}
	index := len(screenshotBuffer) - 1
	screenshotBuffer = screenshotBuffer[:index]
	a.EmitEvent("screenshot-removed", index, len(screenshotBuffer))
}

func (a *App) ClearScreenshots() {
	screenshotBuffer = nil
	ocrTextBuffer = ""
	a.EmitEvent("screenshots-cleared")
}

// GetOCRText 返回最近一次截图识别出的文字
func (a *App) GetOCRText() string {
	return ocrTextBuffer
}

// SetOCRText 设置/覆盖当前 OCR 文本（用户可编辑）
func (a *App) SetOCRText(text string) {
	ocrTextBuffer = text
}

// SendTextMessage 直接发送用户输入的文本给当前模型（默认作为当前对话的追问）
func (a *App) SendTextMessage(text string) {
	cfg := a.configManager.Get()

	if cfg.APIKey == "" {
		a.EmitEvent("require-api-key")
		return
	}

	if cfg.Model == "" {
		a.EmitEvent("toast", "请先选择模型")
		a.EmitEvent("open-settings", "model")
		return
	}

	if strings.TrimSpace(text) == "" {
		a.EmitEvent("toast", "消息内容不能为空")
		return
	}

	if a.taskManager.HasRunningTask() {
		a.EmitEvent("toast", "正在处理中，请稍候...")
		return
	}

	ocrTextBuffer = text
	a.EmitEvent("user-text", text)
	a.triggerSendInternal(true)
}

func (a *App) TriggerSend() {
	a.triggerSendInternal(false)
}

func (a *App) triggerSendInternal(isFollowUp bool) {
	cfg := a.configManager.Get()

	if cfg.APIKey == "" {
		a.EmitEvent("require-api-key")
		return
	}

	if cfg.Model == "" {
		a.EmitEvent("toast", "请先选择模型")
		a.EmitEvent("open-settings", "model")
		return
	}

	// 如果没有截图且没有 OCR 文本，先补一次截图
	if len(screenshotBuffer) == 0 && strings.TrimSpace(ocrTextBuffer) == "" {
		previewResult, err := a.GetScreenshotPreview(
			cfg.CompressionQuality,
			cfg.Sharpening,
			cfg.Grayscale,
			cfg.NoCompression,
			cfg.ScreenshotMode,
		)
		if err != nil {
			logger.Printf("截图失败: %v\n", err)
			a.EmitEvent("toast", "截图失败: "+err.Error())
			return
		}
		screenshotBuffer = append(screenshotBuffer, previewResult.Base64)

		go func(b64 string) {
			tmpPath, err := saveBase64ToTemp(b64)
			if err != nil {
				logger.Printf("保存截图临时文件失败: %v", err)
				return
			}
			defer os.Remove(tmpPath)
			text, err := a.ocrService.Recognize(tmpPath)
			if err != nil {
				logger.Printf("OCR 识别失败: %v", err)
				return
			}
			ocrTextBuffer = text
			a.EmitEvent("ocr-text", text, 0)
		}(previewResult.Base64)
	}

	if a.taskManager.HasRunningTask() {
		logger.Println("忽略重复触发：当前有任务正在运行")
		a.EmitEvent("toast", "正在处理中，请稍候...")
		return
	}

	screenshots := make([]string, len(screenshotBuffer))
	copy(screenshots, screenshotBuffer)
	ocrText := ocrTextBuffer

	screenshotBuffer = nil
	ocrTextBuffer = ""

	a.EmitEvent("start-solving", isFollowUp)
	if len(screenshots) > 0 {
		a.EmitEvent("user-message", screenshots[0])
	}

	ctx, taskID := a.taskManager.StartTask("solve")
	go func() {
		defer a.taskManager.CompleteTask(taskID)
		a.solveInternal(ctx, screenshots, ocrText, isFollowUp)
	}()
}

func (a *App) TriggerSolve() {
	a.TriggerSend()
}

func (a *App) TriggerDeleteScreenshot() {
	a.RemoveLastScreenshot()
}

func (a *App) solveInternal(ctx context.Context, screenshots []string, ocrText string, isFollowUp bool) bool {
	cfg := a.configManager.Get()

	if cfg.APIKey == "" {
		a.EmitEvent("require-api-key")
		return false
	}

	// 针对 DeepSeek 等不支持图片输入的模型，使用 OCR 文本
	useTextMode := isTextOnlyModel(cfg.BaseURL, cfg.Model)

	req := solution.Request{
		Config:      cfg,
		Screenshots: screenshots,
		OCRText:     ocrText,
		UseTextMode: useTextMode,
		IsFollowUp:  isFollowUp,
	}

	cb := solution.Callbacks{
		EmitEvent: a.EmitEvent,
	}

	return a.solver.Solve(ctx, req, cb)
}

// isTextOnlyModel 判断当前配置是否应使用纯文本模式发送
func isTextOnlyModel(baseURL, model string) bool {
	lower := strings.ToLower(baseURL + " " + model)
	return strings.Contains(lower, "deepseek")
}

// saveBase64ToTemp 将 base64 图片数据保存为临时文件，返回文件路径
func saveBase64ToTemp(b64 string) (string, error) {
	data := b64
	commaIdx := strings.Index(b64, ",")
	if commaIdx != -1 {
		data = b64[commaIdx+1:]
	}

	imgBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}

	ext := "png"
	if strings.Contains(b64, "image/jpeg") {
		ext = "jpg"
	}

	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("ia_screenshot_%d.%s", os.Getpid(), ext))
	if err := os.WriteFile(tmpPath, imgBytes, 0644); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	return tmpPath, nil
}

func (a *App) CancelRunningTask() bool {
	return a.taskManager.CancelCurrentTask()
}

func (a *App) IsInterruptThinkingEnabled() bool {
	return a.configManager.Get().InterruptThinking
}
