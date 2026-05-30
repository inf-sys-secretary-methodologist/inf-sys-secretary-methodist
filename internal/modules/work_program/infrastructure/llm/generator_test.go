package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/infrastructure/llm"
)

func sampleReq() usecases.DraftRequest {
	return usecases.DraftRequest{
		DisciplineName:     "Базы данных и СУБД",
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		HoursLecture:       32,
		HoursPractice:      48,
		HoursLab:           16,
		HoursSelfStudy:     24,
		ControlForm:        "экзамен",
		Annotation:         "Курс по проектированию реляционных БД",
	}
}

// chatCompletion wraps a model output string in an OpenAI-style chat
// completion envelope.
func chatCompletion(content string) string {
	resp := map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"role": "assistant", "content": content}, "finish_reason": "stop"},
		},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func newTestGenerator(url string) *llm.Generator {
	return llm.NewGenerator(llm.Config{
		BaseURL: url, APIKey: "test-key", Model: "test-model", Timeout: 5 * time.Second,
	})
}

func TestGenerator_GenerateDraft_HappyPath(t *testing.T) {
	const draftJSON = `{
		"goals": ["Сформировать навыки проектирования БД", "Изучить SQL"],
		"competences": [{"code":"ПК-1","type":"pk","description":"Способен проектировать БД"}],
		"topics": [{"kind":"lecture","title":"Реляционная модель","hours":4},{"kind":"practice","title":"Нормализация","hours":6}],
		"references": [{"kind":"main","citation":"Дейт К. Введение в системы баз данных"}],
		"assessments": [{"type":"current","description":"Контрольная работа","max_score":30,"example_questions":["Приведите отношение к 3НФ"]},{"type":"final","description":"Экзамен","max_score":70}]
	}`

	var gotAuth, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, chatCompletion(draftJSON))
	}))
	defer srv.Close()

	res, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.NoError(t, err)

	assert.Equal(t, "Bearer test-key", gotAuth)
	assert.Equal(t, "/chat/completions", gotPath)
	assert.Contains(t, gotBody, "test-model", "request must name the configured model")
	assert.Contains(t, gotBody, "Базы данных и СУБД", "prompt must carry the discipline context")

	require.Len(t, res.Goals, 2)
	assert.Equal(t, "Сформировать навыки проектирования БД", res.Goals[0])
	require.Len(t, res.Competences, 1)
	assert.Equal(t, "ПК-1", res.Competences[0].Code)
	assert.Equal(t, "pk", res.Competences[0].Type)
	require.Len(t, res.Topics, 2)
	assert.Equal(t, "lecture", res.Topics[0].Kind)
	assert.Equal(t, 4, res.Topics[0].Hours)
	assert.Equal(t, "practice", res.Topics[1].Kind)
	require.Len(t, res.References, 1)
	assert.Equal(t, "main", res.References[0].Kind)
	assert.Equal(t, "Дейт К. Введение в системы баз данных", res.References[0].Citation)
	require.Len(t, res.Assessments, 2)
	assert.Equal(t, "current", res.Assessments[0].Type)
	assert.Equal(t, "Контрольная работа", res.Assessments[0].Description)
	assert.Equal(t, 30, res.Assessments[0].MaxScore)
	require.Len(t, res.Assessments[0].ExampleQuestions, 1)
	assert.Equal(t, "Приведите отношение к 3НФ", res.Assessments[0].ExampleQuestions[0])
	assert.Equal(t, "final", res.Assessments[1].Type)
	assert.Equal(t, 70, res.Assessments[1].MaxScore)
}

func TestGenerator_StripsMarkdownFences(t *testing.T) {
	// Models often wrap JSON in ```json fences despite instructions.
	const fenced = "```json\n{\"goals\":[\"Цель\"],\"competences\":[],\"topics\":[],\"references\":[]}\n```"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, chatCompletion(fenced))
	}))
	defer srv.Close()

	res, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.NoError(t, err)
	require.Len(t, res.Goals, 1)
	assert.Equal(t, "Цель", res.Goals[0])
}

func TestGenerator_HTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"message":"boom"}}`)
	}))
	defer srv.Close()

	_, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.Error(t, err)
}

func TestGenerator_APIErrorField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"error":{"message":"invalid model"}}`)
	}))
	defer srv.Close()

	_, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.Error(t, err)
}

func TestGenerator_NonJSONErrorBody(t *testing.T) {
	// A gateway 502 with an HTML body must surface the status, not a
	// confusing "decode response" error from trying to JSON-parse HTML.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "<html><body>502 Bad Gateway</body></html>")
	}))
	defer srv.Close()

	_, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "502", "non-200 non-JSON body must surface the status code")
}

func TestGenerator_MalformedContentJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, chatCompletion("this is plainly not json"))
	}))
	defer srv.Close()

	_, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.Error(t, err)
}

func TestGenerator_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[]}`)
	}))
	defer srv.Close()

	_, err := newTestGenerator(srv.URL).GenerateDraft(context.Background(), sampleReq())
	require.Error(t, err)
}

func sampleRevisionReq() usecases.RevisionDraftRequest {
	return usecases.RevisionDraftRequest{
		OrderNumber:        "1234",
		OrderTitle:         "Об утверждении ФГОС ВО",
		OrderSummary:       "Изменены требования к часам по дисциплине",
		PublishedYear:      2026,
		WorkProgramTitle:   "Базы данных и СУБД",
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
	}
}

// Slice 7: the extracted text of the order's attached document is woven into
// the revision prompt (the request body sent to the model) so the LLM grounds
// its proposal on the real приказ. Empty text is omitted (no dangling
// section); oversized text is truncated to bound prompt tokens/cost (the
// manual OrderSummary still carries the digest). Asserted through the exported
// GenerateRevision surface by capturing the outgoing request body.
func TestGenerator_GenerateRevision_OrderDocumentText(t *testing.T) {
	const orderTextLabel = "Текст приказа"

	capture := func(t *testing.T, orderText string) string {
		t.Helper()
		var gotBody string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, chatCompletion(`{"change_type":"other","change_summary":"x"}`))
		}))
		defer srv.Close()
		req := sampleRevisionReq()
		req.OrderText = orderText
		_, err := newTestGenerator(srv.URL).GenerateRevision(context.Background(), req)
		require.NoError(t, err)
		return gotBody
	}

	t.Run("present text is included for the LLM", func(t *testing.T) {
		body := capture(t, "Пункт 1. Установить объём часов 18.")
		assert.Contains(t, body, orderTextLabel)
		assert.Contains(t, body, "Установить объём часов 18.")
	})

	t.Run("empty text omits the section", func(t *testing.T) {
		body := capture(t, "")
		assert.NotContains(t, body, orderTextLabel)
	})

	t.Run("oversized text is truncated to bound the prompt", func(t *testing.T) {
		body := capture(t, strings.Repeat("я", 50000))
		assert.Contains(t, body, orderTextLabel)
		assert.Contains(t, body, "усечён", "a truncation marker signals the cut")
		assert.Less(t, strings.Count(body, "я"), 20000,
			"the document text must be capped, not sent whole")
	})
}

func TestGenerator_GenerateRevision_HappyPath(t *testing.T) {
	const revJSON = `{
		"change_type": "hours",
		"change_summary": "Часы лекций сокращены с 32 до 18 в соответствии с приказом",
		"diff_payload": {"hours_lecture": {"before": 32, "after": 18}}
	}`

	var gotAuth, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, chatCompletion(revJSON))
	}))
	defer srv.Close()

	res, err := newTestGenerator(srv.URL).GenerateRevision(context.Background(), sampleRevisionReq())
	require.NoError(t, err)

	assert.Equal(t, "Bearer test-key", gotAuth)
	assert.Equal(t, "/chat/completions", gotPath)
	assert.Contains(t, gotBody, "test-model")
	assert.Contains(t, gotBody, "1234", "prompt must carry the order number")
	assert.Contains(t, gotBody, "Базы данных и СУБД", "prompt must carry the affected РПД title")

	assert.Equal(t, "hours", res.ChangeType)
	assert.Equal(t, "Часы лекций сокращены с 32 до 18 в соответствии с приказом", res.ChangeSummary)
	assert.JSONEq(t, `{"hours_lecture":{"before":32,"after":18}}`, string(res.DiffPayload))
}

func TestGenerator_GenerateRevision_StripsFencesAndOmittedDiff(t *testing.T) {
	// No diff_payload, wrapped in a markdown fence the model often adds.
	const revJSON = "```json\n{\"change_type\":\"literature\",\"change_summary\":\"Обновлён список литературы\"}\n```"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, chatCompletion(revJSON))
	}))
	defer srv.Close()

	res, err := newTestGenerator(srv.URL).GenerateRevision(context.Background(), sampleRevisionReq())
	require.NoError(t, err)
	assert.Equal(t, "literature", res.ChangeType)
	assert.Equal(t, "Обновлён список литературы", res.ChangeSummary)
	assert.Nil(t, res.DiffPayload, "omitted diff_payload maps to nil, not empty JSON")
}

func TestGenerator_GenerateRevision_HTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "<html>502</html>")
	}))
	defer srv.Close()

	_, err := newTestGenerator(srv.URL).GenerateRevision(context.Background(), sampleRevisionReq())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "502")
}
