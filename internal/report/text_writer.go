package report

import (
	"bufio"
	"fmt"
	"io"

	"php-func-parser-go/internal/model"
)

type TextWriter struct{}

func NewTextWriter() *TextWriter { return &TextWriter{} }

func (tw *TextWriter) Write(w io.Writer, reports []model.FileReport) error {
	bw := bufio.NewWriterSize(w, 1<<20)
	for i, r := range reports {
		if _, err := fmt.Fprintln(bw, r.Path); err != nil {
			return err
		}
		for _, fn := range r.Functions {
			if _, err := fmt.Fprintf(bw, "function %s(%s)\n", fn.Name, fn.Params); err != nil {
				return err
			}
		}
		if i != len(reports)-1 {
			if _, err := fmt.Fprintln(bw, ""); err != nil {
				return err
			}
		}
	}
	return bw.Flush()
}
