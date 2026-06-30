package interview

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"Interview_Assistant/pkg/knowledge"
	"Interview_Assistant/pkg/llm"
)

type captureProvider struct {
	messages []llm.Message
	response string
}

func (p *captureProvider) GenerateContentStream(ctx context.Context, messages []llm.Message, onChunk llm.StreamCallback) (llm.Message, error) {
	return llm.Message{}, nil
}

func (p *captureProvider) GenerateContent(ctx context.Context, model string, messages []llm.Message) (llm.Message, error) {
	p.messages = append([]llm.Message(nil), messages...)
	if p.response == "" {
		p.response = "可以结合项目经历直接回答。"
	}
	return llm.NewAssistantMessage(p.response), nil
}

func (p *captureProvider) GetModels(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (p *captureProvider) TestChat(ctx context.Context) error {
	return nil
}

func TestGenerateHintWithoutClearTranscriptStillCallsProvider(t *testing.T) {
	provider := &captureProvider{}
	coach := NewCoach(provider, nil, nil)

	got, err := coach.GenerateHint(context.Background())
	if err != nil {
		t.Fatalf("GenerateHint returned error: %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Fatal("expected non-empty hint")
	}
	if len(provider.messages) != 2 {
		t.Fatalf("expected 2 prompt messages, got %d", len(provider.messages))
	}

	systemPrompt := provider.messages[0].Content
	if !strings.Contains(systemPrompt, "当前还没有清晰转录") {
		t.Fatalf("expected fallback context in prompt, got: %s", systemPrompt)
	}
	if strings.Contains(systemPrompt, "问题不完整，请继续聆听") {
		t.Fatalf("prompt still contains old refusal phrase: %s", systemPrompt)
	}
}

func TestGenerateHintUsesPersonalKnowledge(t *testing.T) {
	root := t.TempDir()
	mineDir := filepath.Join(root, "material", "mine")
	prepDir := filepath.Join(root, "material", "preparation")
	if err := os.MkdirAll(mineDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(prepDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mineDir, "profile.md"), []byte("# 我的项目\n我做过机器人数据闭环项目，负责多模态数据治理和模型训练。"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := &captureProvider{}
	coach := NewCoach(provider, nil, knowledge.NewService([]string{mineDir, prepDir}))
	coach.AddTranscript("你做过机器人数据闭环相关项目吗", "interviewer")

	if _, err := coach.GenerateHint(context.Background()); err != nil {
		t.Fatalf("GenerateHint returned error: %v", err)
	}
	if len(provider.messages) != 2 {
		t.Fatalf("expected 2 prompt messages, got %d", len(provider.messages))
	}

	systemPrompt := provider.messages[0].Content
	if !strings.Contains(systemPrompt, "机器人数据闭环项目") {
		t.Fatalf("expected personal material in prompt, got: %s", systemPrompt)
	}
	if !strings.Contains(systemPrompt, "不要拒答") {
		t.Fatalf("expected no-refusal instruction in prompt, got: %s", systemPrompt)
	}
}
