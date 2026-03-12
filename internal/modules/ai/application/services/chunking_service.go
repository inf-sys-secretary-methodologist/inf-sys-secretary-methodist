// Package services contains application services for the AI module.
package services

import (
	"regexp"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// ChunkingConfig holds configuration for text chunking
type ChunkingConfig struct {
	MaxTokens    int     // Maximum tokens per chunk (default: 512)
	OverlapRatio float64 // Overlap ratio between chunks (default: 0.2)
}

// DefaultChunkingConfig returns the default chunking configuration
func DefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		MaxTokens:    512,
		OverlapRatio: 0.2,
	}
}

// ChunkingService handles text chunking for document indexing
type ChunkingService struct {
	config ChunkingConfig
}

// NewChunkingService creates a new chunking service
func NewChunkingService(config ChunkingConfig) *ChunkingService {
	return &ChunkingService{config: config}
}

// estimateTokens estimates the number of tokens in a string.
// Uses rune count / 2 which is more accurate for multilingual text
// (Cyrillic characters are 2 bytes in UTF-8 but ~0.5 tokens each).
func estimateTokens(s string) int {
	n := len([]rune(s))
	if n == 0 {
		return 0
	}
	return max(n/2, 1)
}

// ChunkDocument splits document text into chunks
func (s *ChunkingService) ChunkDocument(documentID int64, text string) []entities.DocumentChunk {
	// Clean and normalize text
	text = normalizeText(text)

	if len(text) == 0 {
		return nil
	}

	// Split into sentences
	sentences := splitIntoSentences(text)

	chunks := make([]entities.DocumentChunk, 0)
	var currentChunk strings.Builder
	currentTokens := 0
	chunkIndex := 0

	overlapTokens := int(float64(s.config.MaxTokens) * s.config.OverlapRatio)
	overlapText := ""

	for _, sentence := range sentences {
		sentenceTokens := estimateTokens(sentence)

		// If this sentence alone exceeds max tokens, split it
		if sentenceTokens > s.config.MaxTokens {
			// Save current chunk if not empty
			if currentChunk.Len() > 0 {
				chunkText := currentChunk.String()
				tokens := estimateTokens(chunkText)
				chunks = append(chunks, *entities.NewDocumentChunk(documentID, chunkIndex, chunkText, tokens))
				chunkIndex++
				overlapText = getOverlapText(chunkText, overlapTokens)
				currentChunk.Reset()
				currentTokens = 0
			}

			// Split long sentence into multiple chunks
			words := strings.Fields(sentence)
			for _, word := range words {
				wordTokens := estimateTokens(word + " ")
				if currentTokens+wordTokens > s.config.MaxTokens {
					if currentChunk.Len() > 0 {
						chunkText := currentChunk.String()
						tokens := estimateTokens(chunkText)
						chunks = append(chunks, *entities.NewDocumentChunk(documentID, chunkIndex, chunkText, tokens))
						chunkIndex++
						overlapText = getOverlapText(chunkText, overlapTokens)
						currentChunk.Reset()
						currentChunk.WriteString(overlapText)
						currentTokens = estimateTokens(overlapText)
					}
				}
				currentChunk.WriteString(word)
				currentChunk.WriteString(" ")
				currentTokens += wordTokens
			}
			continue
		}

		// Check if adding this sentence would exceed max tokens
		if currentTokens+sentenceTokens > s.config.MaxTokens {
			// Save current chunk
			chunkText := currentChunk.String()
			tokens := estimateTokens(chunkText)
			chunks = append(chunks, *entities.NewDocumentChunk(documentID, chunkIndex, chunkText, tokens))
			chunkIndex++

			// Start new chunk with overlap
			overlapText = getOverlapText(chunkText, overlapTokens)
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
			currentTokens = estimateTokens(overlapText)
		}

		currentChunk.WriteString(sentence)
		currentChunk.WriteString(" ")
		currentTokens += sentenceTokens
	}

	// Save remaining chunk
	if currentChunk.Len() > 0 {
		chunkText := strings.TrimSpace(currentChunk.String())
		if len(chunkText) > 0 {
			tokens := estimateTokens(chunkText)
			chunks = append(chunks, *entities.NewDocumentChunk(documentID, chunkIndex, chunkText, tokens))
		}
	}

	return chunks
}

// normalizeText cleans and normalizes text for chunking.
// Preserves paragraph breaks (\n\n) while collapsing single newlines to spaces.
func normalizeText(text string) string {
	// 1. Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// 2. Replace tabs with spaces
	text = strings.ReplaceAll(text, "\t", " ")

	// 3. Collapse 3+ newlines into exactly \n\n (paragraph break)
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	// 4. Process text: preserve \n\n, replace single \n with space
	var builder strings.Builder
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		if runes[i] == '\n' && i+1 < len(runes) && runes[i+1] == '\n' {
			builder.WriteString("\n\n")
			i += 2
			// Skip any spaces after paragraph break
			for i < len(runes) && runes[i] == ' ' {
				i++
			}
		} else if runes[i] == '\n' {
			builder.WriteRune(' ')
			i++
		} else {
			builder.WriteRune(runes[i])
			i++
		}
	}
	text = builder.String()

	// 5. Collapse multiple spaces into one
	text = multiSpaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// Pre-compiled regexps for hot paths
var (
	multiSpaceRe = regexp.MustCompile(`  +`)
	sentenceRe   = regexp.MustCompile(`([.!?])\s+`)
)

// abbreviation placeholders for sentence splitting
const placeholder = "\x00"

// abbreviationPatterns maps Russian abbreviation patterns to their replacements.
// We use \x00 as a placeholder for dots inside abbreviations.
// Note: Go regexp doesn't support lookbehinds, so we use simpler approaches.
var abbreviationPatterns = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	// "и т.д." / "и т.п." — most specific, match first
	{regexp.MustCompile(`(?i)и\s+т\.\s*д\.`), "и т" + placeholder + "д" + placeholder},
	{regexp.MustCompile(`(?i)и\s+т\.\s*п\.`), "и т" + placeholder + "п" + placeholder},
	// Multi-char abbreviations
	{regexp.MustCompile(`(?i)т\.д\.`), "т" + placeholder + "д" + placeholder},
	{regexp.MustCompile(`(?i)т\.п\.`), "т" + placeholder + "п" + placeholder},
	{regexp.MustCompile(`(?i)т\.е\.`), "т" + placeholder + "е" + placeholder},
	{regexp.MustCompile(`(?i)проф\.`), "проф" + placeholder},
	{regexp.MustCompile(`(?i)доц\.`), "доц" + placeholder},
	{regexp.MustCompile(`(?i)акад\.`), "акад" + placeholder},
	{regexp.MustCompile(`(?i)др\.`), "др" + placeholder},
	{regexp.MustCompile(`(?i)ул\.`), "ул" + placeholder},
	{regexp.MustCompile(`(?i)пр\.`), "пр" + placeholder},
	// Abbreviations with context guards (followed by digit or uppercase)
	{regexp.MustCompile(`(?i)ст\.\s*(\d)`), "ст" + placeholder + " ${1}"},
	{regexp.MustCompile(`(?i)п\.\s*(\d)`), "п" + placeholder + " ${1}"},
	{regexp.MustCompile(`(?i)г\.\s*([А-ЯA-Z\d])`), "г" + placeholder + " ${1}"},
	{regexp.MustCompile(`(?i)д\.\s*(\d)`), "д" + placeholder + " ${1}"},
}

// protectAbbreviations replaces dots in known abbreviations with placeholders
func protectAbbreviations(text string) string {
	for _, ap := range abbreviationPatterns {
		text = ap.pattern.ReplaceAllString(text, ap.replacement)
	}
	return text
}

// restoreAbbreviations restores dots from placeholders
func restoreAbbreviations(text string) string {
	return strings.ReplaceAll(text, placeholder, ".")
}

// splitIntoSentences splits text into sentences, respecting paragraph boundaries
// and Russian abbreviations.
func splitIntoSentences(text string) []string {
	if len(text) == 0 {
		return nil
	}

	// First split by paragraph breaks — these are highest priority boundaries
	paragraphs := strings.Split(text, "\n\n")

	var allSentences []string
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// Protect abbreviations
		protected := protectAbbreviations(paragraph)

		// Split by sentence-ending punctuation followed by whitespace
		parts := sentenceRe.Split(protected, -1)
		delimiters := sentenceRe.FindAllStringSubmatch(protected, -1)

		var sentences []string
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			// Re-attach the delimiter to the sentence (except for the last part)
			if i < len(delimiters) {
				part += delimiters[i][1] // add back the punctuation mark
			}
			// Restore abbreviation dots
			part = restoreAbbreviations(part)
			sentences = append(sentences, part)
		}

		allSentences = append(allSentences, sentences...)
	}

	return allSentences
}

// getOverlapText extracts overlap text from the end of a chunk
func getOverlapText(text string, overlapTokens int) string {
	if overlapTokens <= 0 {
		return ""
	}

	// Estimate characters based on token count (rune count = tokens * 2)
	overlapChars := overlapTokens * 2
	runes := []rune(text)
	if overlapChars >= len(runes) {
		return text
	}

	// Find the last sentence boundary within overlap range
	overlapStart := len(runes) - overlapChars
	prefix := string(runes[:overlapStart])
	lastSentence := strings.LastIndex(prefix, ". ")

	if lastSentence > 0 && lastSentence < len(prefix) {
		return strings.TrimSpace(string([]rune(text)[runeIndex(text, lastSentence+2):]))
	}

	// Fall back to word boundary
	lastSpace := strings.LastIndex(prefix, " ")
	if lastSpace > 0 {
		return strings.TrimSpace(string([]rune(text)[runeIndex(text, lastSpace+1):]))
	}

	return strings.TrimSpace(string(runes[overlapStart:]))
}

// runeIndex converts a byte offset to a rune offset
func runeIndex(s string, byteOffset int) int {
	return len([]rune(s[:byteOffset]))
}
