package parser

import "php-func-parser-go/internal/model"

type FunctionParser interface {
	Parse(content []byte) []model.FunctionDecl
}
