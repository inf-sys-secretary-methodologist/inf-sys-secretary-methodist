package llm

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
)

//go:embed prompts/work_program_draft.txt
var systemPrompt string

const defaultTimeout = 60 * time.Second

// Config configures the OpenAI-compatible draft generator. BaseURL /
// APIKey / Model are provider-agnostic (OpenRouter by default, but any
// OpenAI-compatible endpoint works by changing these three).
type Config struct {
	BaseURL     string
	APIKey      string
	Model       string
	Timeout     time.Duration
	Temperature float64
	MaxTokens   int
}

// Generator calls an OpenAI-compatible chat-completions endpoint to
// draft the content of a WorkProgram (РПД). It implements the
// application-layer usecases.DraftGenerator port and depends on nothing
// from the ai/ module (standalone client, configured by base_url/key/
// model so OpenRouter / OpenAI / Groq are all served by configuration).
type Generator struct {
	cfg    Config
	client *http.Client
}

// compile-time check that the adapter satisfies the application port.
var _ usecases.DraftGenerator = (*Generator)(nil)

// NewGenerator wires the generator. A non-positive timeout falls back to
// defaultTimeout so a mis-configured zero does not mean "no timeout".
func NewGenerator(cfg Config) *Generator {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Generator{cfg: cfg, client: &http.Client{Timeout: timeout}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// draftWire is the JSON contract the model must return; it mirrors the
// schema in prompts/work_program_draft.txt and maps 1:1 to the
// application DTOs.
type draftWire struct {
	Goals       []string `json:"goals"`
	Competences []struct {
		Code        string `json:"code"`
		Type        string `json:"type"`
		Description string `json:"description"`
	} `json:"competences"`
	Topics []struct {
		Kind  string `json:"kind"`
		Title string `json:"title"`
		Hours int    `json:"hours"`
	} `json:"topics"`
	References []struct {
		Kind     string `json:"kind"`
		Citation string `json:"citation"`
	} `json:"references"`
}

// GenerateDraft POSTs the discipline context to {BaseURL}/chat/completions
// and maps the model's JSON output into a DraftResult. The use case
// re-validates every row through the domain constructors, so this method
// does not enforce domain invariants — it only transports + parses.
func (g *Generator) GenerateDraft(ctx context.Context, req usecases.DraftRequest) (usecases.DraftResult, error) {
	reqBody := chatRequest{
		Model: g.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildUserPrompt(req)},
		},
		Temperature:    g.cfg.Temperature,
		MaxTokens:      g.cfg.MaxTokens,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.cfg.APIKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: read response: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: decode response: %w", err)
	}
	if parsed.Error != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: api error: %s", parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: unexpected status %d: %s",
			resp.StatusCode, truncate(string(respBody), 256))
	}
	if len(parsed.Choices) == 0 {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: response had no choices")
	}

	content := stripCodeFence(parsed.Choices[0].Message.Content)
	var wire draftWire
	if err := json.Unmarshal([]byte(content), &wire); err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: decode generated draft: %w", err)
	}
	return toDraftResult(wire), nil
}

func buildUserPrompt(req usecases.DraftRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Дисциплина: %s\n", req.DisciplineName)
	fmt.Fprintf(&b, "Специальность (код): %s\n", req.SpecialtyCode)
	fmt.Fprintf(&b, "Год начала применения: %d\n", req.ApplicableFromYear)
	fmt.Fprintf(&b, "Форма контроля: %s\n", req.ControlForm)
	fmt.Fprintf(&b, "Часовой бюджет — лекции: %d, практики: %d, лабораторные: %d, самостоятельная работа: %d\n",
		req.HoursLecture, req.HoursPractice, req.HoursLab, req.HoursSelfStudy)
	if strings.TrimSpace(req.Annotation) != "" {
		fmt.Fprintf(&b, "Аннотация: %s\n", req.Annotation)
	}
	b.WriteString("Составь черновик содержательной части РПД в формате JSON по описанным правилам.")
	return b.String()
}

// stripCodeFence removes a leading ```json (or ```) fence and the
// trailing ``` that models often add despite the JSON-only instruction.
func stripCodeFence(s string) string {
	t := strings.TrimSpace(s)
	if !strings.HasPrefix(t, "```") {
		return t
	}
	t = strings.TrimPrefix(t, "```")
	if i := strings.IndexByte(t, '\n'); i >= 0 {
		t = t[i+1:] // drop the rest of the opening fence line (e.g. "json")
	}
	t = strings.TrimSpace(t)
	t = strings.TrimSuffix(t, "```")
	return strings.TrimSpace(t)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func toDraftResult(w draftWire) usecases.DraftResult {
	res := usecases.DraftResult{Goals: w.Goals}
	for _, c := range w.Competences {
		res.Competences = append(res.Competences, usecases.CompetenceDraft{
			Code: c.Code, Type: c.Type, Description: c.Description,
		})
	}
	for _, t := range w.Topics {
		res.Topics = append(res.Topics, usecases.TopicDraft{
			Kind: t.Kind, Title: t.Title, Hours: t.Hours,
		})
	}
	for _, r := range w.References {
		res.References = append(res.References, usecases.ReferenceDraft{
			Kind: r.Kind, Citation: r.Citation,
		})
	}
	return res
}
