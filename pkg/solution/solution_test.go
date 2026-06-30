package solution

import (
	"context"
	"testing"

	"Interview_Assistant/pkg/config"
	"Interview_Assistant/pkg/llm"
)

type captureStreamProvider struct {
	messages []llm.Message
}

func (p *captureStreamProvider) GenerateContentStream(ctx context.Context, messages []llm.Message, onChunk llm.StreamCallback) (llm.Message, error) {
	p.messages = append([]llm.Message(nil), messages...)
	if onChunk != nil {
		onChunk(llm.StreamChunk{Type: llm.ChunkContent, Content: "ok"})
	}
	return llm.NewAssistantMessage("ok"), nil
}

func (p *captureStreamProvider) GenerateContent(ctx context.Context, model string, messages []llm.Message) (llm.Message, error) {
	return llm.NewAssistantMessage("ok"), nil
}

func (p *captureStreamProvider) GetModels(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (p *captureStreamProvider) TestChat(ctx context.Context) error {
	return nil
}

func TestSolveSendsOnlyUserMessageWithoutSystemPrompt(t *testing.T) {
	provider := &captureStreamProvider{}
	solver := NewSolver(provider)

	ok := solver.Solve(context.Background(), Request{
		Config: config.Config{
			APIKey:   "test-key",
			Model:    "gpt-4o",
			DomainId: "dev-cpp-exam",
			Prompt:   "must not be injected",
		},
		UserText:    "请只解释思路，不要写代码",
		Screenshots: []string{"data:image/png;base64,abc"},
		IsFollowUp:  true,
	}, Callbacks{})
	if !ok {
		t.Fatal("Solve returned false")
	}
	if len(provider.messages) != 1 {
		t.Fatalf("expected exactly one message, got %d", len(provider.messages))
	}
	msg := provider.messages[0]
	if msg.Role != llm.RoleUser {
		t.Fatalf("expected user message, got %s", msg.Role)
	}
	if len(msg.Parts) != 2 {
		t.Fatalf("expected text + image parts, got %d", len(msg.Parts))
	}
	if msg.Parts[0].Type != llm.ContentText || msg.Parts[0].Text != "请只解释思路，不要写代码" {
		t.Fatalf("unexpected text part: %+v", msg.Parts[0])
	}
	if msg.Parts[1].Type != llm.ContentImage {
		t.Fatalf("expected image part, got %+v", msg.Parts[1])
	}
}

func TestBuildUserMessageTextModeDeduplicatesOCR(t *testing.T) {
	msg, err := buildUserMessage(Request{
		UserText:    "题目文字",
		OCRText:     "题目文字",
		Screenshots: []string{"data:image/png;base64,abc"},
		UseTextMode: true,
	})
	if err != nil {
		t.Fatalf("buildUserMessage returned error: %v", err)
	}
	if msg.Content != "题目文字" {
		t.Fatalf("expected deduplicated content, got %q", msg.Content)
	}
}
