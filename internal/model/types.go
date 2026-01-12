package model

type FunctionDecl struct {
	Name   string
	Params string // normalized, without surrounding parentheses
}

type FileReport struct {
	Path      string
	Functions []FunctionDecl
}
