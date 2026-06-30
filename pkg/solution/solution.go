package solution

import (
	"Interview_Assistant/pkg/config"
	"Interview_Assistant/pkg/llm"
	"Interview_Assistant/pkg/logger"
	"context"
	"errors"
	"fmt"
	"strings"
)

// MaxConversationRounds is the maximum number of conversation rounds to keep.
const MaxConversationRounds = 10

type Callbacks struct {
	EmitEvent func(event string, data ...interface{})
}

type Request struct {
	Config      config.Config
	Screenshots []string
	UserText    string
	OCRText     string
	UseTextMode bool
	IsFollowUp  bool
}

type Solver struct {
	llmProvider llm.Provider
	chatHistory []llm.Message
}

func NewSolver(provider llm.Provider) *Solver {
	return &Solver{
		llmProvider: provider,
		chatHistory: make([]llm.Message, 0),
	}
}

func (s *Solver) SetProvider(provider llm.Provider) {
	s.llmProvider = provider
}

func (s *Solver) ClearHistory() {
	s.chatHistory = make([]llm.Message, 0)
}

func (s *Solver) Solve(ctx context.Context, req Request, cb Callbacks) bool {
	if req.Config.APIKey == "" {
		if cb.EmitEvent != nil {
			cb.EmitEvent("require-api-key")
		}
		return false
	}

	logger.Println("开始 AI 对话流程")

	currentUserMsg, err := buildUserMessage(req)
	if err != nil {
		logger.Printf("构造用户消息失败: %v\n", err)
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution-error", err.Error())
		}
		return false
	}

	// 新题目清空历史；追问保留历史
	if !req.IsFollowUp {
		s.chatHistory = make([]llm.Message, 0)
	}

	s.chatHistory = append(s.chatHistory, currentUserMsg)
	s.trimChatHistory()

	messagesToSend := make([]llm.Message, len(s.chatHistory))
	copy(messagesToSend, s.chatHistory)

	if cb.EmitEvent != nil {
		cb.EmitEvent("solution-stream-start")
	}

	response, err := s.llmProvider.GenerateContentStream(ctx, messagesToSend, func(chunk llm.StreamChunk) {
		if cb.EmitEvent == nil {
			return
		}

		switch chunk.Type {
		case llm.ChunkThinking:
			cb.EmitEvent("solution-stream-thinking", chunk.Content)
		case llm.ChunkContent:
			cb.EmitEvent("solution-stream-chunk", chunk.Content)
		}
	})

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			logger.Println("当前任务已中断（用户产生新输入）")
			if cb.EmitEvent != nil {
				cb.EmitEvent("solution-error", "context canceled")
			}
			return false
		}

		logger.Printf("LLM 请求失败: %v\n", err)
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution-error", err.Error())
		}
		return false
	}

	logger.Printf("[解题] 模型返回内容长度: %d", len(response.Content))
	logger.Printf("[解题] 模型返回内容: %s", response.Content)
	logger.Printf("[解题] 模型返回思考链长度: %d", len(response.Thinking))

	if response.Content == "" && response.Thinking == "" {
		logger.Println("[解题] 警告: 模型返回内容为空")
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution-error", "模型返回内容为空，请检查模型配置或稍后重试")
		}
		return false
	}

	if cb.EmitEvent != nil {
		cb.EmitEvent("solution", response.Content)
	}

	// 将模型回复加入历史，支持后续追问
	s.chatHistory = append(s.chatHistory, llm.NewAssistantMessage(response.Content))
	return true
}

func buildUserMessage(req Request) (llm.Message, error) {
	userText := strings.TrimSpace(req.UserText)
	ocrText := strings.TrimSpace(req.OCRText)

	if req.UseTextMode {
		parts := make([]string, 0, 2)
		if userText != "" {
			parts = append(parts, userText)
		}
		if len(req.Screenshots) > 0 {
			if ocrText == "" && userText == "" {
				return llm.Message{}, fmt.Errorf("当前模型不支持图片输入，截图 OCR 尚未得到可发送文本；请稍等 OCR 完成，或手动输入问题")
			}
			if userText == "" && ocrText != "" {
				parts = append(parts, ocrText)
			}
		}
		if len(parts) == 0 {
			return llm.Message{}, fmt.Errorf("请输入消息或先添加截图附件")
		}
		return llm.NewUserMessage(strings.Join(parts, "\n\n")), nil
	}

	if userText == "" && len(req.Screenshots) == 0 {
		return llm.Message{}, fmt.Errorf("请输入消息或先添加截图附件")
	}

	userParts := make([]llm.ContentPart, 0, len(req.Screenshots)+1)
	if userText != "" {
		userParts = append(userParts, llm.TextPart(userText))
	}
	for _, screenshot := range req.Screenshots {
		if strings.TrimSpace(screenshot) == "" {
			continue
		}
		userParts = append(userParts, llm.ImagePart(screenshot))
	}
	if len(userParts) == 0 {
		return llm.Message{}, fmt.Errorf("请输入消息或先添加截图附件")
	}
	return llm.NewMultiPartMessage(llm.RoleUser, userParts), nil
}

// trimChatHistory keeps only the most recent conversation rounds.
func (s *Solver) trimChatHistory() {
	maxNonSystemMsgs := MaxConversationRounds * 2
	if len(s.chatHistory) <= maxNonSystemMsgs {
		return
	}

	oldLen := len(s.chatHistory)
	s.chatHistory = append([]llm.Message(nil), s.chatHistory[len(s.chatHistory)-maxNonSystemMsgs:]...)
	logger.Printf("裁剪对话历史: %d -> %d 条消息 (保留最近 %d 轮对话)", oldLen, len(s.chatHistory), MaxConversationRounds)
}
