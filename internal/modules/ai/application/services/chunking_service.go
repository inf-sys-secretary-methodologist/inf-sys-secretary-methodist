// Package services contains application services for the AI module.
package services

import (
	"strings"
	"unicode"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// ChunkingConfig holds configuration for text chunking
type ChunkingConfig struct {
	MaxTokens    int     // Maximum tokens per chunk (default: 512)
	OverlapRatio float64 // Overlap ratio between chunks (default: 0.1)
}

// DefaultChunkingConfig returns the default chunking configuration
func DefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		MaxTokens:    512,
		OverlapRatio: 0.1,
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

// ChunkDocument splits document text into chunks
func (s *ChunkingService) ChunkDocument(documentID int64, text string) []entities.DocumentChunk {
	// Clean and normalize text
	text = normalizeText(text)

	if len(text) == 0 {
		return nil
	}

	// Split into sentences
	sentences := splitIntoSentences(text)

	// Estimate tokens (rough approximation: 4 chars = 1 token)
	estimateTokens := func(s string) int {
		return len(s) / 4
	}

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

// normalizeText cleans and normalizes text for chunking
func normalizeText(text string) string {
	// Replace multiple whitespace with single space
	var builder strings.Builder
	prevSpace := false

	for _, r := range text {
		if unicode.IsSpace(r) {
			if !prevSpace {
				builder.WriteRune(' ')
				prevSpace = true
			}
		} else {
			builder.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(builder.String())
}

// splitIntoSentences splits text into sentences
func splitIntoSentences(text string) []string {
	// Simple sentence splitting by common delimiters
	delimiters := []string{". ", "! ", "? ", ".\n", "!\n", "?\n"}

	sentences := []string{text}

	for _, delim := range delimiters {
		var newSentences []string
		for _, s := range sentences {
			parts := strings.Split(s, delim)
			for i, part := range parts {
				part = strings.TrimSpace(part)
				if len(part) == 0 {
					continue
				}
				// Add back the delimiter (except for the last part)
				if i < len(parts)-1 {
					part += string(delim[0])
				}
				newSentences = append(newSentences, part)
			}
		}
		sentences = newSentences
	}

	return sentences
}

// getOverlapText extracts overlap text from the end of a chunk
func getOverlapText(text string, overlapTokens int) string {
	if overlapTokens <= 0 {
		return ""
	}

	// Estimate characters based on token count
	overlapChars := overlapTokens * 4
	if overlapChars >= len(text) {
		return text
	}

	// Find the last sentence boundary within overlap range
	overlapStart := len(text) - overlapChars
	lastSentence := strings.LastIndex(text[:overlapStart], ". ")

	if lastSentence > 0 && lastSentence < overlapStart {
		return strings.TrimSpace(text[lastSentence+2:])
	}

	// Fall back to word boundary
	lastSpace := strings.LastIndex(text[:overlapStart], " ")
	if lastSpace > 0 {
		return strings.TrimSpace(text[lastSpace+1:])
	}

	return strings.TrimSpace(text[overlapStart:])
}
