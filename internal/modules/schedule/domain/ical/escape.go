package ical

import (
	"strings"
	"unicode/utf8"
)

// textEscaper escapes the characters that RFC 5545 section 3.3.11 reserves in
// TEXT values. Backslash is escaped first (it is the escape character), and a
// bare CR is dropped so a CRLF collapses to a single "\n".
var textEscaper = strings.NewReplacer(
	`\`, `\\`,
	";", `\;`,
	",", `\,`,
	"\n", `\n`,
	"\r", "",
)

// escapeText escapes a TEXT property value per RFC 5545.
func escapeText(s string) string {
	return textEscaper.Replace(s)
}

// foldLimit is the RFC 5545 maximum content-line length in octets.
const foldLimit = 75

// foldLine folds a single content line so that no line exceeds foldLimit
// octets (RFC 5545 section 3.1). Breaks fall only on UTF-8 rune boundaries so
// multi-byte characters are never split, and continuation lines are prefixed
// with a single space (which counts toward the octet limit). The input must
// not already contain CRLF.
func foldLine(line string) string {
	if len(line) <= foldLimit {
		return line
	}

	var b strings.Builder
	start := 0
	budget := foldLimit
	for start < len(line) {
		end := start
		for end < len(line) {
			_, size := utf8.DecodeRuneInString(line[end:])
			if end-start+size > budget {
				break
			}
			end += size
		}
		if start > 0 {
			b.WriteString("\r\n ")
		}
		b.WriteString(line[start:end])
		start = end
		budget = foldLimit - 1 // leading space on continuation lines counts too
	}
	return b.String()
}
