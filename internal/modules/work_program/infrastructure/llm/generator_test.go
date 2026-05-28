package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
		"references": [{"kind":"main","citation":"Дейт К. Введение в системы баз данных"}]
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
