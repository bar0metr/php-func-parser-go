package report

import (
	"io"

	"php-func-parser-go/internal/model"
)

type Writer interface {
	Write(w io.Writer, reports []model.FileReport) error
}
