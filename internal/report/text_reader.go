package report

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"php-func-parser-go/internal/model"
)

// ReadText parses the text report format produced by TextWriter.
//
// Blocks are separated by blank lines. Each block is:
//
//	<file path>
//	function <name>(<params>)
//	function <name>(<params>)
//
// For compatibility, lines without the leading "function " prefix are also accepted.
func ReadText(r io.Reader) ([]model.FileReport, error) {
	s := bufio.NewScanner(r)
	// Support long lines (very long signatures).
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 4*1024*1024)

	var (
		lineNo  int
		current *model.FileReport
		out     []model.FileReport
	)

	flush := func() {
		if current == nil {
			return
		}
		out = append(out, *current)
		current = nil
	}

	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" {
			// End of block.
			flush()
			continue
		}

		if current == nil {
			current = &model.FileReport{Path: line}
			continue
		}

		decl, err := parseFunctionLine(line)
		if err != nil {
			return nil, fmt.Errorf("parse report line %d: %w", lineNo, err)
		}
		current.Functions = append(current.Functions, decl)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	flush()
	return out, nil
}

func parseFunctionLine(line string) (model.FunctionDecl, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "function ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "function "))
	}
	open := strings.IndexByte(line, '(')
	close := strings.LastIndexByte(line, ')')
	if open <= 0 || close < open {
		return model.FunctionDecl{}, fmt.Errorf("invalid function line: %q", line)
	}
	name := strings.TrimSpace(line[:open])
	params := strings.TrimSpace(line[open+1 : close])
	if name == "" {
		return model.FunctionDecl{}, fmt.Errorf("empty function name in line: %q", line)
	}
	return model.FunctionDecl{Name: name, Params: params}, nil
}
