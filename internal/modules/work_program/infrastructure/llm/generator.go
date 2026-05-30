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
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
)

//go:embed prompts/work_program_draft.txt
var systemPrompt string

//go:embed prompts/work_program_revision.txt
var revisionPrompt string

const defaultTimeout = 60 * time.Second

// maxDiffPayloadBytes caps the LLM-originated structured diff blob stored
// on a revision. The change_summary (human text) is the authoritative
// field; an oversized diff_payload is dropped to nil rather than persisted,
// so a hostile/verbose model cannot bloat the aggregate (carry-forward
// from the 11a review).
const maxDiffPayloadBytes = 64 << 10 // 64 KiB

// maxRespBytes caps how much of the upstream response we read into
// memory, so a hostile or buggy provider cannot OOM the process.
const maxRespBytes = 1 << 20 // 1 MiB

// maxRevisionOrderTextRunes caps how much of the order's extracted document
// text (slice 7) is woven into the revision prompt. A full ministry order
// can run to many pages; sending it whole would blow the model's context
// window and the per-call cost. The manual OrderSummary still carries the
// human digest, so a truncated body is acceptable — the cut is marked.
const maxRevisionOrderTextRunes = 8000

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

// compile-time checks that the adapter satisfies the application ports.
var (
	_ usecases.DraftGenerator         = (*Generator)(nil)
	_ usecases.RevisionDraftGenerator = (*Generator)(nil)
)

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
	Assessments []struct {
		Type             string   `json:"type"`
		Description      string   `json:"description"`
		MaxScore         int      `json:"max_score"`
		ExampleQuestions []string `json:"example_questions"`
	} `json:"assessments"`
}

// GenerateDraft POSTs the discipline context to {BaseURL}/chat/completions
// and maps the model's JSON output into a DraftResult. The use case
// re-validates every row through the domain constructors, so this method
// does not enforce domain invariants — it only transports + parses.
func (g *Generator) GenerateDraft(ctx context.Context, req usecases.DraftRequest) (usecases.DraftResult, error) {
	content, err := g.complete(ctx, systemPrompt, buildUserPrompt(req))
	if err != nil {
		return usecases.DraftResult{}, err
	}
	var wire draftWire
	if err := json.Unmarshal([]byte(content), &wire); err != nil {
		return usecases.DraftResult{}, fmt.Errorf("work_program/llm: decode generated draft: %w", err)
	}
	return toDraftResult(wire), nil
}

// GenerateRevision asks the model for one лист актуализации (revision)
// proposal grounded on a recorded order + the affected РПД's identity, and
// maps the model's JSON into a RevisionProposal. The use case re-validates
// change_type through the domain constructor, so this method only
// transports + parses; it does guard the optional diff_payload size.
func (g *Generator) GenerateRevision(ctx context.Context, req usecases.RevisionDraftRequest) (usecases.RevisionProposal, error) {
	content, err := g.complete(ctx, revisionPrompt, buildRevisionUserPrompt(req))
	if err != nil {
		return usecases.RevisionProposal{}, err
	}
	var wire revisionWire
	if err := json.Unmarshal([]byte(content), &wire); err != nil {
		return usecases.RevisionProposal{}, fmt.Errorf("work_program/llm: decode generated revision: %w", err)
	}
	return toRevisionProposal(wire), nil
}

// complete POSTs a system+user prompt pair to {BaseURL}/chat/completions
// and returns the model's fence-stripped message content. Shared by both
// GenerateDraft and GenerateRevision.
func (g *Generator) complete(ctx context.Context, system, user string) (string, error) {
	reqBody := chatRequest{
		Model: g.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:    g.cfg.Temperature,
		MaxTokens:      g.cfg.MaxTokens,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("work_program/llm: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("work_program/llm: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.cfg.APIKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("work_program/llm: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxRespBytes))
	if err != nil {
		return "", fmt.Errorf("work_program/llm: read response: %w", err)
	}

	// Check the status before decoding: a non-200 body is often non-JSON
	// (e.g. an HTML gateway 502) and would otherwise fail json.Unmarshal
	// with a misleading "decode response" error instead of the status.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("work_program/llm: unexpected status %d: %s",
			resp.StatusCode, truncate(string(respBody), 256))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("work_program/llm: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("work_program/llm: api error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("work_program/llm: response had no choices")
	}
	return stripCodeFence(parsed.Choices[0].Message.Content), nil
}

// revisionWire is the JSON contract the model must return for a revision
// proposal; it mirrors prompts/work_program_revision.txt. diff_payload is
// optional raw JSON (omitted → nil).
type revisionWire struct {
	ChangeType    string          `json:"change_type"`
	ChangeSummary string          `json:"change_summary"`
	DiffPayload   json.RawMessage `json:"diff_payload"`
}

func buildRevisionUserPrompt(req usecases.RevisionDraftRequest) string {
	var b strings.Builder
	b.WriteString("Приказ Минобрнауки:\n")
	fmt.Fprintf(&b, "Номер: %s\n", req.OrderNumber)
	fmt.Fprintf(&b, "Заголовок: %s\n", req.OrderTitle)
	if strings.TrimSpace(req.OrderSummary) != "" {
		fmt.Fprintf(&b, "Краткое содержание: %s\n", req.OrderSummary)
	}
	fmt.Fprintf(&b, "Год публикации: %d\n", req.PublishedYear)
	if txt := strings.TrimSpace(req.OrderText); txt != "" {
		fmt.Fprintf(&b, "Текст приказа (из приложенного документа):\n%s\n", truncateRunes(txt, maxRevisionOrderTextRunes))
	}
	b.WriteString("\n")
	b.WriteString("Рабочая программа дисциплины (РПД):\n")
	fmt.Fprintf(&b, "Название: %s\n", req.WorkProgramTitle)
	fmt.Fprintf(&b, "Специальность (код): %s\n", req.SpecialtyCode)
	fmt.Fprintf(&b, "Год начала применения: %d\n\n", req.ApplicableFromYear)
	b.WriteString("Предложи одну актуализацию этой РПД во исполнение приказа в формате JSON по описанным правилам.")
	return b.String()
}

// truncateRunes returns s capped to max runes, appending a Russian
// truncation marker when the cut happens. Rune-based (not byte-based) so a
// multibyte order text is never sliced mid-character.
func truncateRunes(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit]) + "…(текст приказа усечён)"
}

// toRevisionProposal maps the wire shape into the application DTO,
// dropping an oversized diff_payload to nil (the change_summary is the
// authoritative human field; the structured diff is best-effort).
func toRevisionProposal(w revisionWire) usecases.RevisionProposal {
	p := usecases.RevisionProposal{
		ChangeType:    w.ChangeType,
		ChangeSummary: w.ChangeSummary,
	}
	if len(w.DiffPayload) > 0 && len(w.DiffPayload) <= maxDiffPayloadBytes {
		p.DiffPayload = []byte(w.DiffPayload)
	}
	return p
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
	// Back off to a rune boundary so the truncated error string never
	// splits a multibyte rune mid-character.
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
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
	for _, a := range w.Assessments {
		res.Assessments = append(res.Assessments, usecases.AssessmentDraft{
			Type:             a.Type,
			Description:      a.Description,
			MaxScore:         a.MaxScore,
			ExampleQuestions: a.ExampleQuestions,
		})
	}
	return res
}
