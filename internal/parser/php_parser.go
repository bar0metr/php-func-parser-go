package parser

import (
	"bytes"
	"unicode"

	"php-func-parser-go/internal/model"
)

type PHPFunctionParser struct{}

func NewPHPFunctionParser() *PHPFunctionParser { return &PHPFunctionParser{} }

func (p *PHPFunctionParser) Parse(content []byte) []model.FunctionDecl {
	// Fast single-pass tokenizer focused on named function declarations.
	var out []model.FunctionDecl

	st := scanState{}
	for i := 0; i < len(content); i++ {
		c := content[i]
		if st.step(content, i) {
			continue
		}

		if c != 'f' {
			continue
		}
		if !hasKeywordAt(content, i, "function") {
			continue
		}
		if !wordBoundary(content, i, len("function")) {
			continue
		}

		j := i + len("function")
		j = skipSpace(content, j)
		if j < len(content) && content[j] == '&' {
			j++
			j = skipSpace(content, j)
		}

		name, next := readIdentifier(content, j)
		if name == "" {
			// anonymous function: function(...) { ... }
			continue
		}
		j = skipSpace(content, next)
		if j >= len(content) || content[j] != '(' {
			continue
		}

		params, end := readParenBlockNormalized(content, j)
		if end == -1 {
			continue
		}

		out = append(out, model.FunctionDecl{Name: name, Params: params})
		i = end
	}
	return out
}

type scanState struct {
	inLineComment  bool
	inBlockComment bool
	inSingle       bool
	inDouble       bool
	inBacktick     bool
}

// step returns true if caller should skip further processing for the current index.
func (s *scanState) step(b []byte, i int) bool {
	c := b[i]

	if s.inLineComment {
		if c == '\n' {
			s.inLineComment = false
		}
		return true
	}
	if s.inBlockComment {
		if c == '*' && i+1 < len(b) && b[i+1] == '/' {
			s.inBlockComment = false
		}
		return true
	}
	if s.inSingle {
		if c == '\\' {
			return true
		}
		if c == '\'' {
			s.inSingle = false
		}
		return true
	}
	if s.inDouble {
		if c == '\\' {
			return true
		}
		if c == '"' {
			s.inDouble = false
		}
		return true
	}
	if s.inBacktick {
		if c == '\\' {
			return true
		}
		if c == '`' {
			s.inBacktick = false
		}
		return true
	}

	// Enter states
	if c == '/' && i+1 < len(b) {
		n := b[i+1]
		if n == '/' {
			s.inLineComment = true
			return true
		}
		if n == '*' {
			s.inBlockComment = true
			return true
		}
	}
	if c == '#' {
		s.inLineComment = true
		return true
	}
	if c == '\'' {
		s.inSingle = true
		return true
	}
	if c == '"' {
		s.inDouble = true
		return true
	}
	if c == '`' {
		s.inBacktick = true
		return true
	}
	return false
}

func hasKeywordAt(b []byte, i int, kw string) bool {
	if i+len(kw) > len(b) {
		return false
	}
	return bytes.Equal(b[i:i+len(kw)], []byte(kw))
}

func wordBoundary(b []byte, start int, length int) bool {
	if start > 0 {
		if isIdentChar(rune(b[start-1])) {
			return false
		}
	}
	end := start + length
	if end < len(b) {
		if isIdentChar(rune(b[end])) {
			return false
		}
	}
	return true
}

func skipSpace(b []byte, i int) int {
	for i < len(b) {
		s := b[i]
		if s == ' ' || s == '\t' || s == '\n' || s == '\r' || s == '\f' {
			i++
			continue
		}
		break
	}
	return i
}

func readIdentifier(b []byte, i int) (string, int) {
	if i >= len(b) {
		return "", i
	}
	r := rune(b[i])
	if !(unicode.IsLetter(r) || r == '_') {
		return "", i
	}
	start := i
	i++
	for i < len(b) {
		r = rune(b[i])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			i++
			continue
		}
		break
	}
	return string(b[start:i]), i
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func readParenBlockNormalized(b []byte, i int) (string, int) {
	if i >= len(b) || b[i] != '(' {
		return "", -1
	}
	depth := 0
	inSingle := false
	inDouble := false
	inBacktick := false
	inLineComment := false
	inBlockComment := false
	lastWasSpace := false

	var buf bytes.Buffer
	for j := i; j < len(b); j++ {
		c := b[j]

		if inLineComment {
			if c == '\n' {
				inLineComment = false
				lastWasSpace = true
			}
			continue
		}
		if inBlockComment {
			if c == '*' && j+1 < len(b) && b[j+1] == '/' {
				inBlockComment = false
				j++
				lastWasSpace = true
			}
			continue
		}
		if inSingle {
			buf.WriteByte(c)
			if c == '\\' {
				if j+1 < len(b) {
					j++
					buf.WriteByte(b[j])
				}
				continue
			}
			if c == '\'' {
				inSingle = false
			}
			continue
		}
		if inDouble {
			buf.WriteByte(c)
			if c == '\\' {
				if j+1 < len(b) {
					j++
					buf.WriteByte(b[j])
				}
				continue
			}
			if c == '"' {
				inDouble = false
			}
			continue
		}
		if inBacktick {
			buf.WriteByte(c)
			if c == '\\' {
				if j+1 < len(b) {
					j++
					buf.WriteByte(b[j])
				}
				continue
			}
			if c == '`' {
				inBacktick = false
			}
			continue
		}

		if c == '/' && j+1 < len(b) {
			n := b[j+1]
			if n == '/' {
				inLineComment = true
				j++
				continue
			}
			if n == '*' {
				inBlockComment = true
				j++
				continue
			}
		}
		if c == '#' {
			inLineComment = true
			continue
		}
		if c == '\'' {
			inSingle = true
			buf.WriteByte(c)
			lastWasSpace = false
			continue
		}
		if c == '"' {
			inDouble = true
			buf.WriteByte(c)
			lastWasSpace = false
			continue
		}
		if c == '`' {
			inBacktick = true
			buf.WriteByte(c)
			lastWasSpace = false
			continue
		}

		switch c {
		case '(':
			depth++
			if depth > 1 {
				buf.WriteByte(c)
				lastWasSpace = false
			}
			continue
		case ')':
			depth--
			if depth == 0 {
				return string(bytes.TrimSpace(buf.Bytes())), j
			}
			buf.WriteByte(c)
			lastWasSpace = false
			continue
		}

		if depth == 0 {
			continue
		}

		if isWhitespace(c) {
			if !lastWasSpace {
				buf.WriteByte(' ')
				lastWasSpace = true
			}
			continue
		}
		buf.WriteByte(c)
		lastWasSpace = false
	}
	return "", -1
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}
