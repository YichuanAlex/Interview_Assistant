package interview

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"Interview_Assistant/pkg/llm"
)

// Coach 根据实时转录文本生成面试提示
type Coach struct {
	mu       sync.Mutex
	provider llm.Provider
	history  []string
}

// NewCoach 创建面试提示生成器
func NewCoach(provider llm.Provider) *Coach {
	return &Coach{
		provider: provider,
		history:  make([]string, 0),
	}
}

// SetProvider 更新 LLM provider
func (c *Coach) SetProvider(provider llm.Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.provider = provider
}

// AddTranscript 添加一条转录文本
func (c *Coach) AddTranscript(text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = append(c.history, text)
	// 保留最近 30 条，避免上下文过长
	if len(c.history) > 30 {
		c.history = c.history[len(c.history)-30:]
	}
}

// ClearHistory 清空历史
func (c *Coach) ClearHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = c.history[:0]
}

// GetContext 返回当前上下文文本
func (c *Coach) GetContext() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.history, "\n")
}

// GenerateHint 基于当前上下文生成面试提示
func (c *Coach) GenerateHint(ctx context.Context) (string, error) {
	c.mu.Lock()
	provider := c.provider
	contextText := strings.Join(c.history, "\n")
	c.mu.Unlock()

	if provider == nil {
		return "", fmt.Errorf("LLM provider 未初始化")
	}

	if strings.TrimSpace(contextText) == "" {
		return "", fmt.Errorf("暂无转录内容，请先开始语音转录")
	}

	systemPrompt := `你是一位资深技术面试官的实时辅助教练。请根据下面面试官和候选人的实时对话转录文本，给出对候选人有帮助的简短提示。

要求：
1. 先判断当前是面试官提问还是候选人回答。
2. 如果是面试官提问：总结问题核心，并给出 2-3 条回答思路或要点（中文）。
3. 如果是候选人回答：不打断，仅给出 1 条可能的补充方向（可选）。
4. 回答要简洁，控制在 100 字以内，不要写完整长篇答案。
5. 如果问题不明确，回复“问题不完整，请继续聆听”。

实时转录文本：
` + contextText

	messages := []llm.Message{
		llm.NewSystemMessage(systemPrompt),
		llm.NewUserMessage("请给出当前面试提示。"),
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	msg, err := provider.GenerateContent(timeoutCtx, "", messages)
	if err != nil {
		return "", err
	}

	return msg.Content, nil
}
