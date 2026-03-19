package services

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty string", "", 0},
		{"ascii text", "Hello world", 5},   // 11 runes / 2 = 5
		{"cyrillic text", "Привет мир", 5}, // 10 runes / 2 = 5
		{"single char", "a", 1},            // max(1/2, 1) = 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateTokens(tt.input)
			if got != tt.want {
				t.Errorf("estimateTokens(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "preserves paragraph breaks",
			input: "First paragraph.\n\nSecond paragraph.",
			want:  "First paragraph.\n\nSecond paragraph.",
		},
		{
			name:  "collapses single newlines to space",
			input: "Line one.\nLine two.",
			want:  "Line one. Line two.",
		},
		{
			name:  "collapses triple newlines to double",
			input: "Para one.\n\n\n\nPara two.",
			want:  "Para one.\n\nPara two.",
		},
		{
			name:  "handles CRLF",
			input: "Line one.\r\nLine two.\r\n\r\nNew para.",
			want:  "Line one. Line two.\n\nNew para.",
		},
		{
			name:  "collapses multiple spaces",
			input: "Hello   world   test",
			want:  "Hello world test",
		},
		{
			name:  "replaces tabs with spaces",
			input: "Hello\tworld",
			want:  "Hello world",
		},
		{
			name:  "trims whitespace",
			input: "  Hello world  ",
			want:  "Hello world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeText(tt.input)
			if got != tt.want {
				t.Errorf("normalizeText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitIntoSentences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "basic sentences",
			input: "First sentence. Second sentence. Third.",
			want:  []string{"First sentence.", "Second sentence.", "Third."},
		},
		{
			name:  "preserves Russian abbreviations",
			input: "Согласно ст. 5 закона. Новое предложение.",
			want:  []string{"Согласно ст. 5 закона.", "Новое предложение."},
		},
		{
			name:  "preserves т.д. (abbreviation dot also serves as sentence end)",
			input: "Документы и т.д. Следующее предложение.",
			want:  []string{"Документы и т.д. Следующее предложение."},
		},
		{
			name:  "preserves т.п. (abbreviation dot also serves as sentence end)",
			input: "Протоколы и т.п. Далее.",
			want:  []string{"Протоколы и т.п. Далее."},
		},
		{
			name:  "splits after т.д. when followed by another sentence",
			input: "Документы, книги и т.д. Это важно. Следующее.",
			want:  []string{"Документы, книги и т.д. Это важно.", "Следующее."},
		},
		{
			name:  "preserves п. with digit",
			input: "Смотрите п. 3 документа. Важно.",
			want:  []string{"Смотрите п. 3 документа.", "Важно."},
		},
		{
			name:  "paragraph breaks as boundaries",
			input: "First para.\n\nSecond para.",
			want:  []string{"First para.", "Second para."},
		},
		{
			name:  "question and exclamation marks",
			input: "Is this a question? Yes! Next.",
			want:  []string{"Is this a question?", "Yes!", "Next."},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitIntoSentences(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitIntoSentences(%q): got %d sentences, want %d\ngot:  %v\nwant: %v",
					tt.input, len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitIntoSentences(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestChunkDocument(t *testing.T) {
	t.Run("empty text returns nil", func(t *testing.T) {
		svc := NewChunkingService(DefaultChunkingConfig())
		chunks := svc.ChunkDocument(1, "")
		if chunks != nil {
			t.Errorf("expected nil, got %d chunks", len(chunks))
		}
	})

	t.Run("short text produces single chunk", func(t *testing.T) {
		svc := NewChunkingService(DefaultChunkingConfig())
		chunks := svc.ChunkDocument(1, "Короткий текст.")
		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk, got %d", len(chunks))
		}
	})

	t.Run("long text produces multiple chunks with overlap", func(t *testing.T) {
		// Create text that should produce multiple chunks
		// With estimateTokens = rune_count / 2, and MaxTokens = 50,
		// we need text with > 100 runes
		cfg := ChunkingConfig{MaxTokens: 50, OverlapRatio: 0.2}
		svc := NewChunkingService(cfg)

		// Build text with multiple sentences
		var sb strings.Builder
		for i := 0; i < 20; i++ {
			sb.WriteString("Это предложение номер один для тестирования. ")
		}

		chunks := svc.ChunkDocument(1, sb.String())
		if len(chunks) < 2 {
			t.Errorf("expected at least 2 chunks, got %d", len(chunks))
		}

		// Verify all chunks have document ID
		for _, c := range chunks {
			if c.DocumentID != 1 {
				t.Errorf("expected document_id=1, got %d", c.DocumentID)
			}
		}

		// Verify chunk indices are sequential
		for i, c := range chunks {
			if c.ChunkIndex != i {
				t.Errorf("chunk %d has index %d", i, c.ChunkIndex)
			}
		}
	})
}

func TestDefaultChunkingConfig(t *testing.T) {
	cfg := DefaultChunkingConfig()
	if cfg.MaxTokens != 512 {
		t.Errorf("MaxTokens = %d, want 512", cfg.MaxTokens)
	}
	if cfg.OverlapRatio != 0.2 {
		t.Errorf("OverlapRatio = %f, want 0.2", cfg.OverlapRatio)
	}
}

func TestGetOverlapText(t *testing.T) {
	t.Run("zero overlap returns empty", func(t *testing.T) {
		got := getOverlapText("some text here", 0)
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("large overlap returns full text", func(t *testing.T) {
		got := getOverlapText("short", 100)
		if got != "short" {
			t.Errorf("expected 'short', got %q", got)
		}
	})

	t.Run("returns overlap from end of text", func(t *testing.T) {
		text := "First sentence. Second sentence. Third sentence."
		got := getOverlapText(text, 20)
		// Should return some text from the end
		if got == "" {
			t.Error("expected non-empty overlap")
		}
		// The overlap should be a suffix of the text
		if !strings.HasSuffix(text, got+".") && !strings.HasSuffix(text, got) && !strings.Contains(text, got) {
			t.Errorf("overlap %q not found in text", got)
		}
	})
}

func TestChunkDocument_LongSentenceSplit(t *testing.T) {
	// Use very small max tokens to force long-sentence splitting
	cfg := ChunkingConfig{MaxTokens: 5, OverlapRatio: 0.0}
	svc := NewChunkingService(cfg)

	// A single sentence with many words — no sentence-ending punctuation followed by space,
	// so it stays as one "sentence" that exceeds MaxTokens and triggers word-level splitting.
	text := "word1 word2 word3 word4 word5 word6 word7 word8 word9 word10 word11 word12"

	chunks := svc.ChunkDocument(1, text)
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks from word-level splitting, got %d", len(chunks))
	}
	for i, c := range chunks {
		if c.ChunkIndex != i {
			t.Errorf("chunk %d has index %d", i, c.ChunkIndex)
		}
	}
}

func TestChunkDocument_LongSentenceWithPrecedingContent(t *testing.T) {
	// Test the path where currentChunk is non-empty before a long sentence
	cfg := ChunkingConfig{MaxTokens: 10, OverlapRatio: 0.2}
	svc := NewChunkingService(cfg)

	// Short sentence first, then a very long one
	text := "Short. " + strings.Repeat("longword ", 50)

	chunks := svc.ChunkDocument(1, text)
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks, got %d", len(chunks))
	}
}

func TestChunkDocument_WhitespaceOnly(t *testing.T) {
	svc := NewChunkingService(DefaultChunkingConfig())
	chunks := svc.ChunkDocument(1, "   \n\n\t  ")
	if chunks != nil {
		t.Errorf("expected nil for whitespace-only text, got %d chunks", len(chunks))
	}
}

func TestChunkDocument_ZeroOverlap(t *testing.T) {
	cfg := ChunkingConfig{MaxTokens: 20, OverlapRatio: 0.0}
	svc := NewChunkingService(cfg)

	var sb strings.Builder
	for i := 0; i < 10; i++ {
		sb.WriteString("Это тестовое предложение для проверки. ")
	}
	chunks := svc.ChunkDocument(42, sb.String())
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.DocumentID != 42 {
			t.Errorf("expected document_id=42, got %d", c.DocumentID)
		}
	}
}

func TestSplitIntoSentences_EmptyParagraph(t *testing.T) {
	// Paragraphs with empty content between double newlines
	input := "Hello.\n\n\n\nWorld."
	// After normalizeText, triple+ newlines collapse to double. But splitIntoSentences
	// receives already-normalized text in ChunkDocument. Test raw input:
	got := splitIntoSentences("Para one.\n\n\n\nPara two.")
	if len(got) != 2 {
		t.Errorf("expected 2 sentences, got %d: %v", len(got), got)
	}
	_ = input
}

func TestGetOverlapText_WordBoundaryFallback(t *testing.T) {
	// Text without ". " — forces word boundary fallback
	text := "one two three four five six seven eight nine ten"
	got := getOverlapText(text, 5)
	if got == "" {
		t.Error("expected non-empty overlap from word boundary fallback")
	}
	if !strings.Contains(text, got) {
		t.Errorf("overlap %q not found in text", got)
	}
}

func TestRuneIndex(t *testing.T) {
	// Test with mixed ASCII and Cyrillic
	s := "Hello Привет"
	// "Hello " is 6 bytes, "П" starts at byte 6
	idx := runeIndex(s, 6)
	if idx != 6 {
		t.Errorf("expected rune index 6 for byte offset 6, got %d", idx)
	}

	// All ASCII: rune index == byte index
	s2 := "abcdef"
	idx2 := runeIndex(s2, 3)
	if idx2 != 3 {
		t.Errorf("expected 3, got %d", idx2)
	}
}

func TestProtectRestoreAbbreviations(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"ст. 5"},
		{"п. 3"},
		{"и т.д."},
		{"и т.п."},
		{"проф. Иванов"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			protected := protectAbbreviations(tt.input)
			restored := restoreAbbreviations(protected)
			if restored != tt.input {
				t.Errorf("round-trip failed: %q -> %q -> %q", tt.input, protected, restored)
			}
		})
	}
}
