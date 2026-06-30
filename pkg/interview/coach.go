package interview

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"Interview_Assistant/pkg/knowledge"
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
	knowledge        *knowledge.Service
	interviewerTexts []string
	intervieweeTexts []string

	emitFunc     func(eventName string, data ...interface{})
	silenceTimer *time.Timer
	lastHintTime time.Time
}

// NewCoach 创建面试提示生成器
func NewCoach(provider llm.Provider, emitFunc func(eventName string, data ...interface{}), knowledgeService *knowledge.Service) *Coach {
	return &Coach{
		provider:         provider,
		knowledge:        knowledgeService,
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
	latestQuestion := latestNonEmptyLine(interviewerContext)
	if latestQuestion == "" {
		latestQuestion = "（当前转录没有完整问题，请根据已有上下文和候选人材料推断最可能的面试问题）"
	}
	personalContext := ""
	if c.knowledge != nil {
		query := latestQuestion + "\n" + interviewerContext + "\n" + intervieweeContext
		if strings.TrimSpace(query) == "" {
			query = "自我介绍 项目经历 实习经历 多模态 数据闭环 机器人 RAG 模型训练 系统架构"
		}
		personalContext = c.knowledge.BuildContext(query, 5, 4200)
	}

	systemPrompt := `你是一位资深技术面试官的实时辅助教练。请根据下面面试官的提问和候选人的回答，给出对候选人有帮助的简短提示。

要求：
1. 重点针对面试官的最新问题给出回答思路。
2. 必须优先结合“候选人个性化材料”，把回答落到候选人的项目、实习、论文、数据平台/多模态/机器人数据经验上。
3. 候选人的回答只作为上下文参考，不要基于面试者的话触发新的提示。
4. 即使转录不完整，也不要拒答、不要说“问题不完整”、不要要求继续聆听；请基于上下文推断最可能的问题并直接给出可用回答。
5. 回答要简洁，控制在 220 字以内，给出 2-3 条要点即可；尽量使用候选人第一人称可直接复述的话术。

---
最新面试官问题：
` + latestQuestion + `

候选人个性化材料（从本地 material/preparation 与 material/mine 检索得到，可能为空）：
` + emptyPlaceholder(personalContext) + `

面试官说的话：
` + interviewerContext + `

候选人（面试者）说的话（仅作参考）：
` + intervieweeContext

	messages := []llm.Message{
		llm.NewSystemMessage(systemPrompt),
		llm.NewUserMessage("请直接给出当前可回答的话术提示，不要输出“问题不完整”。"),
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	msg, err := c.provider.GenerateContent(timeoutCtx, "", messages)
	if err != nil {
		return "", err
	}

	return msg.Content, nil
}

func latestNonEmptyLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

func emptyPlaceholder(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "（未检索到足够相关的个性化材料）"
	}
	return text
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

	if strings.TrimSpace(interviewerContext) == "" && strings.TrimSpace(intervieweeContext) == "" {
		interviewerContext = "当前还没有清晰转录。请基于候选人个性化材料，给出一段可用于技术面试开场、自我介绍或项目追问的通用回答提示。"
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
