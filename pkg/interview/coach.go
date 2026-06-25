package interview

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"Interview_Assistant/pkg/llm"
	"Interview_Assistant/pkg/logger"
)

const (
	// interviewerSilenceThreshold 判定面试官说完话的静音间隔
	interviewerSilenceThreshold = 2 * time.Second
)

// Coach 根据实时转录文本生成面试提示
type Coach struct {
	mu               sync.Mutex
	provider         llm.Provider
	interviewerTexts []string
	intervieweeTexts []string

	emitFunc      func(eventName string, data ...interface{})
	silenceTimer  *time.Timer
	lastHintTime  time.Time
}

// NewCoach 创建面试提示生成器
func NewCoach(provider llm.Provider, emitFunc func(eventName string, data ...interface{})) *Coach {
	return &Coach{
		provider:         provider,
		interviewerTexts: make([]string, 0),
		intervieweeTexts: make([]string, 0),
		emitFunc:         emitFunc,
	}
}

// SetProvider 更新 LLM provider
func (c *Coach) SetProvider(provider llm.Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.provider = provider
}

// AddTranscript 添加一条转录文本
// role: "interviewer" 表示面试官，"interviewee" 表示面试者
func (c *Coach) AddTranscript(text string, role string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if role == "interviewee" {
		c.intervieweeTexts = append(c.intervieweeTexts, text)
		if len(c.intervieweeTexts) > 30 {
			c.intervieweeTexts = c.intervieweeTexts[len(c.intervieweeTexts)-30:]
		}
		return
	}

	// 面试官文本：追加并触发静音检测
	c.interviewerTexts = append(c.interviewerTexts, text)
	if len(c.interviewerTexts) > 30 {
		c.interviewerTexts = c.interviewerTexts[len(c.interviewerTexts)-30:]
	}

	c.resetSilenceTimerLocked()
}

func (c *Coach) resetSilenceTimerLocked() {
	if c.silenceTimer != nil {
		c.silenceTimer.Stop()
	}
	c.silenceTimer = time.AfterFunc(interviewerSilenceThreshold, c.onInterviewerSilence)
}

func (c *Coach) onInterviewerSilence() {
	c.mu.Lock()
	provider := c.provider
	if provider == nil {
		c.mu.Unlock()
		return
	}

	// 防止过于频繁调用 API，两次提示间隔至少 3 秒
	if time.Since(c.lastHintTime) < 3*time.Second {
		c.mu.Unlock()
		return
	}
	c.lastHintTime = time.Now()

	interviewerContext := strings.Join(c.interviewerTexts, "\n")
	intervieweeContext := strings.Join(c.intervieweeTexts, "\n")
	c.mu.Unlock()

	if strings.TrimSpace(interviewerContext) == "" {
		return
	}

	hint, err := c.generateHintWithContext(interviewerContext, intervieweeContext)
	if err != nil {
		logger.Printf("生成面试提示失败: %v\n", err)
		return
	}

	if c.emitFunc != nil {
		c.emitFunc("interview-hint", hint)
	}
}

func (c *Coach) generateHintWithContext(interviewerContext, intervieweeContext string) (string, error) {
	systemPrompt := `你是一位资深技术面试官的实时辅助教练。请根据下面面试官的提问和候选人的回答，给出对候选人有帮助的简短提示。

要求：
1. 重点针对面试官的最新问题给出回答思路。
2. 候选人的回答只作为上下文参考，不要基于面试者的话触发新的提示。
3. 回答要简洁，控制在 150 字以内，给出 2-3 条要点即可。
4. 如果面试官问题不完整，回复“问题不完整，请继续聆听”。

---
面试官说的话：
` + interviewerContext + `

候选人（面试者）说的话（仅作参考）：
` + intervieweeContext

	messages := []llm.Message{
		llm.NewSystemMessage(systemPrompt),
		llm.NewUserMessage("请给出当前面试提示。"),
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	msg, err := c.provider.GenerateContent(timeoutCtx, "", messages)
	if err != nil {
		return "", err
	}

	return msg.Content, nil
}

// GenerateHint 手动基于当前上下文生成面试提示
func (c *Coach) GenerateHint(ctx context.Context) (string, error) {
	c.mu.Lock()
	provider := c.provider
	interviewerContext := strings.Join(c.interviewerTexts, "\n")
	intervieweeContext := strings.Join(c.intervieweeTexts, "\n")
	c.mu.Unlock()

	if provider == nil {
		return "", fmt.Errorf("LLM provider 未初始化")
	}

	if strings.TrimSpace(interviewerContext) == "" {
		return "", fmt.Errorf("暂无面试官转录内容，请先开始语音转录")
	}

	return c.generateHintWithContext(interviewerContext, intervieweeContext)
}

// ClearHistory 清空历史
func (c *Coach) ClearHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interviewerTexts = c.interviewerTexts[:0]
	c.intervieweeTexts = c.intervieweeTexts[:0]
	if c.silenceTimer != nil {
		c.silenceTimer.Stop()
		c.silenceTimer = nil
	}
}

// GetContext 返回当前上下文文本
func (c *Coach) GetContext() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var sb strings.Builder
	if len(c.interviewerTexts) > 0 {
		sb.WriteString("面试官:\n")
		sb.WriteString(strings.Join(c.interviewerTexts, "\n"))
	}
	if len(c.intervieweeTexts) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("面试者:\n")
		sb.WriteString(strings.Join(c.intervieweeTexts, "\n"))
	}
	return sb.String()
}
