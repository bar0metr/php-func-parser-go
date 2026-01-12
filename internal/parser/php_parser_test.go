package parser

import (
	"testing"

	"php-func-parser-go/internal/model"
)

func TestPHPFunctionParser_Parse_Basic(t *testing.T) {
	p := NewPHPFunctionParser()
	src := []byte(`<?php
function tralslaal(aparenejeje) {}
function dhsjwjwjsje(dhshshs,dhsjsh,djsjsjd){}
?>`)

	got := p.Parse(src)
	want := []model.FunctionDecl{
		{Name: "tralslaal", Params: "aparenejeje"},
		{Name: "dhsjwjwjsje", Params: "dhshshs,dhsjsh,djsjsjd"},
	}
	assertDecls(t, got, want)
}

func TestPHPFunctionParser_Parse_Ignores_Comments_And_Strings(t *testing.T) {
	p := NewPHPFunctionParser()
	src := []byte(`<?php
// function fake(a)
# function fake2(b)
/* function fake3(c) */
$S = "function not_a_decl(x)";
$T = 'function not_a_decl2(y)';
function real1($a, $b = "/* ) */") {}
?>`)

	got := p.Parse(src)
	want := []model.FunctionDecl{
		{Name: "real1", Params: "$a, $b = \"/* ) */\""},
	}
	assertDecls(t, got, want)
}

func TestPHPFunctionParser_Parse_Multiline_Params_Normalized(t *testing.T) {
	p := NewPHPFunctionParser()
	src := []byte(`<?php
function f(
	int $a,
	array $b = array(1,2,3),
	callable $c = null
) { }
?>`)

	got := p.Parse(src)
	want := []model.FunctionDecl{
		{Name: "f", Params: "int $a, array $b = array(1,2,3), callable $c = null"},
	}
	assertDecls(t, got, want)
}

func TestPHPFunctionParser_Parse_Skips_Anonymous_Functions(t *testing.T) {
	p := NewPHPFunctionParser()
	src := []byte(`<?php
$fn = function($a) { return $a; };
function named($b) {}
?>`)

	got := p.Parse(src)
	want := []model.FunctionDecl{{Name: "named", Params: "$b"}}
	assertDecls(t, got, want)
}

func assertDecls(t *testing.T, got, want []model.FunctionDecl) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d\n%#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Name != want[i].Name || got[i].Params != want[i].Params {
			t.Fatalf("idx %d: got %#v want %#v", i, got[i], want[i])
		}
	}
}
