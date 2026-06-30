package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"Interview_Assistant/pkg/logger"
)

const (
	defaultChunkSize = 900
	chunkOverlap     = 120
)

var latinTokenRe = regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_+#./-]*`)

type Chunk struct {
	ID      string
	Path    string
	Title   string
	Kind    string
	Content string
	tokens  map[string]int
}

type Service struct {
	mu     sync.RWMutex
	roots  []string
	chunks []Chunk
}

func NewService(roots []string) *Service {
	s := &Service{roots: roots}
	if err := s.Reload(); err != nil {
		logger.Printf("加载个性化材料失败: %v\n", err)
	}
	return s
}

func (s *Service) Reload() error {
	var chunks []Chunk
	for _, root := range s.roots {
		if strings.TrimSpace(root) == "" {
			continue
		}
		rootChunks, err := loadRoot(root)
		if err != nil {
			logger.Printf("跳过材料目录 %s: %v\n", root, err)
			continue
		}
		chunks = append(chunks, rootChunks...)
	}

	s.mu.Lock()
	s.chunks = chunks
	s.mu.Unlock()

	logger.Printf("个性化面试材料已加载: %d 个片段\n", len(chunks))
	return nil
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.chunks)
}

func (s *Service) Search(query string, limit int, maxChars int) []Chunk {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 5
	}
	if maxChars <= 0 {
		maxChars = 4200
	}

	s.mu.RLock()
	chunks := append([]Chunk(nil), s.chunks...)
	s.mu.RUnlock()

	type scored struct {
		chunk Chunk
		score float64
	}
	scoredChunks := make([]scored, 0, len(chunks))
	for _, chunk := range chunks {
		score := scoreChunk(queryTokens, chunk)
		if score > 0 {
			scoredChunks = append(scoredChunks, scored{chunk: chunk, score: score})
		}
	}
	sort.Slice(scoredChunks, func(i, j int) bool {
		if scoredChunks[i].score == scoredChunks[j].score {
			return len(scoredChunks[i].chunk.Content) > len(scoredChunks[j].chunk.Content)
		}
		return scoredChunks[i].score > scoredChunks[j].score
	})

	result := make([]Chunk, 0, limit)
	totalChars := 0
	seenFiles := map[string]int{}
	for _, item := range scoredChunks {
		if len(result) >= limit {
			break
		}
		contentLen := utf8.RuneCountInString(item.chunk.Content)
		if totalChars+contentLen > maxChars && len(result) > 0 {
			continue
		}
		if seenFiles[item.chunk.Path] >= 3 {
			continue
		}
		result = append(result, item.chunk)
		totalChars += contentLen
		seenFiles[item.chunk.Path]++
	}
	return result
}

func (s *Service) BuildContext(query string, limit int, maxChars int) string {
	chunks := s.Search(query, limit, maxChars)
	if len(chunks) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, chunk := range chunks {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("[材料%d | %s | %s]\n", i+1, chunk.Kind, chunk.Title))
		sb.WriteString(strings.TrimSpace(chunk.Content))
	}
	return sb.String()
}

func loadRoot(root string) ([]Chunk, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("不是目录")
	}

	var chunks []Chunk
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel := path
		if cwd, err := os.Getwd(); err == nil {
			if r, relErr := filepath.Rel(cwd, path); relErr == nil {
				rel = r
			}
		}
		kind := "preparation"
		if strings.Contains(filepath.ToSlash(path), "/mine/") {
			kind = "mine"
		}
		chunks = append(chunks, splitDocument(rel, kind, string(data))...)
		return nil
	})
	return chunks, err
}

func splitDocument(path string, kind string, content string) []Chunk {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			title = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))
			break
		}
	}

	paragraphs := splitParagraphs(normalized)
	var chunks []Chunk
	var current strings.Builder
	chunkIndex := 0
	section := title
	flush := func() {
		text := strings.TrimSpace(current.String())
		if text == "" {
			return
		}
		chunkIndex++
		chunkTitle := title
		if section != "" && section != title {
			chunkTitle = title + " / " + section
		}
		chunks = append(chunks, Chunk{
			ID:      fmt.Sprintf("%s#%d", path, chunkIndex),
			Path:    path,
			Title:   chunkTitle,
			Kind:    kind,
			Content: text,
			tokens:  tokenize(path + " " + chunkTitle + "\n" + text),
		})
		current.Reset()
	}

	for _, para := range paragraphs {
		trimmed := strings.TrimSpace(para)
		if strings.HasPrefix(trimmed, "#") {
			section = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		}
		if current.Len() > 0 && utf8.RuneCountInString(current.String()+"\n\n"+trimmed) > defaultChunkSize {
			previous := current.String()
			flush()
			if utf8.RuneCountInString(previous) > chunkOverlap {
				current.WriteString(tailRunes(previous, chunkOverlap))
				current.WriteString("\n\n")
			}
		}
		current.WriteString(trimmed)
		current.WriteString("\n\n")
	}
	flush()
	return chunks
}

func splitParagraphs(content string) []string {
	raw := strings.Split(content, "\n\n")
	result := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func tailRunes(text string, n int) string {
	runes := []rune(text)
	if len(runes) <= n {
		return text
	}
	return string(runes[len(runes)-n:])
}

func scoreChunk(queryTokens map[string]int, chunk Chunk) float64 {
	score := 0.0
	for token, qCount := range queryTokens {
		if cCount, ok := chunk.tokens[token]; ok {
			score += float64(qCount*cCount) * tokenWeight(token)
		}
	}
	if chunk.Kind == "mine" {
		score *= 1.35
	}
	return score
}

func tokenWeight(token string) float64 {
	runeCount := utf8.RuneCountInString(token)
	if runeCount >= 4 {
		return 2.0
	}
	if runeCount == 3 {
		return 1.5
	}
	return 1.0
}

func tokenize(text string) map[string]int {
	tokens := map[string]int{}
	lower := strings.ToLower(text)
	for _, token := range latinTokenRe.FindAllString(lower, -1) {
		token = strings.Trim(token, "./-")
		if len(token) >= 2 {
			tokens[token]++
		}
	}

	var hanRun []rune
	flushHan := func() {
		if len(hanRun) == 0 {
			return
		}
		if len(hanRun) == 1 {
			tokens[string(hanRun)]++
		} else {
			for i := 0; i < len(hanRun)-1; i++ {
				tokens[string(hanRun[i:i+2])]++
			}
			if len(hanRun) >= 4 {
				for i := 0; i < len(hanRun)-2; i++ {
					tokens[string(hanRun[i:i+3])]++
				}
			}
		}
		hanRun = hanRun[:0]
	}

	for _, r := range lower {
		if unicode.Is(unicode.Han, r) {
			hanRun = append(hanRun, r)
		} else {
			flushHan()
		}
	}
	flushHan()
	return tokens
}
