package parser

import (
	"reflect"
	"testing"

	"github.com/mmarchesotti/build-your-own-grep/internal/ast"
	"github.com/mmarchesotti/build-your-own-grep/internal/lexer"
)

// --- Test Helper Functions ---
func lit(char rune) ast.ASTNode { return &ast.LiteralNode{Literal: char} }

func alt(left, right ast.ASTNode) ast.ASTNode {
	return &ast.AlternationNode{Left: left, Right: right}
}

func concat(left, right ast.ASTNode) ast.ASTNode {
	return &ast.ConcatenationNode{Left: left, Right: right}
}
func star(child ast.ASTNode) ast.ASTNode { return &ast.KleeneClosureNode{Child: child} }
func plus(child ast.ASTNode) ast.ASTNode { return &ast.PositiveClosureNode{Child: child} }
func opt(child ast.ASTNode) ast.ASTNode  { return &ast.OptionalNode{Child: child} }
func cs(pos bool, lits []rune) ast.ASTNode {
	return &ast.CharacterSetNode{IsPositive: pos, Literals: lits}
}

func capg(index int, child ast.ASTNode) ast.ASTNode {
	return &ast.CaptureGroupNode{GroupIndex: index, Child: child}
}

// --- Main Test Function ---

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      ast.ASTNode
		expectedCount int
	}{
		{
			name:          "single literal",
			input:         "a",
			expected:      lit('a'),
			expectedCount: 1,
		},
		{
			name:          "simple concatenation",
			input:         "ab",
			expected:      concat(lit('a'), lit('b')),
			expectedCount: 1,
		},
		{
			name:          "long concatenation",
			input:         "abc",
			expected:      concat(concat(lit('a'), lit('b')), lit('c')),
			expectedCount: 1,
		},
		{
			name:          "simple alternation",
			input:         "a|b",
			expected:      alt(lit('a'), lit('b')),
			expectedCount: 1,
		},
		{
			name:          "alternation and concatenation precedence",
			input:         "ab|c",
			expected:      alt(concat(lit('a'), lit('b')), lit('c')),
			expectedCount: 1,
		},
		{
			name:  "parentheses for scope",
			input: "a(b|c)",
			expected: concat(
				lit('a'),
				capg(1, alt(lit('b'), lit('c'))),
			),
			expectedCount: 2,
		},
		{
			name:          "simple kleene star",
			input:         "a*",
			expected:      star(lit('a')),
			expectedCount: 1,
		},
		{
			name:  "kleene star on a group",
			input: "(ab)*",
			expected: star(
				capg(1, concat(lit('a'), lit('b'))),
			),
			expectedCount: 2,
		},
		{
			name:          "all quantifiers",
			input:         "a*b+c?",
			expected:      concat(concat(star(lit('a')), plus(lit('b'))), opt(lit('c'))),
			expectedCount: 1,
		},
		{
			name:          "character set",
			input:         "[abc]",
			expected:      cs(true, []rune{'a', 'b', 'c'}),
			expectedCount: 1,
		},
		{
			name:  "complex expression",
			input: "a(b|c)*d",
			expected: concat(
				concat(
					lit('a'),
					star(capg(1, alt(lit('b'), lit('c')))),
				),
				lit('d'),
			),
			expectedCount: 2,
		},
		{
			name:  "nested capture groups",
			input: "a(b(c))d",
			expected: concat(
				concat(
					lit('a'),
					capg(1, concat(lit('b'), capg(2, lit('c')))),
				),
				lit('d'),
			),
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := lexer.Tokenize(tt.input)
			if err != nil {
				t.Fatalf("Tokenize() returned an unexpected error: %v", err)
			}

			actual, actualCount, err := Parse(tokens)
			if err != nil {
				t.Fatalf("Parse() returned an unexpected error: %v", err)
			}

			if actualCount != tt.expectedCount {
				t.Errorf("Parse() for input '%s' returned wrong capture count", tt.input)
				t.Errorf("got:  %d", actualCount)
				t.Errorf("want: %d", tt.expectedCount)
			}

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Parse() for input '%s' failed", tt.input)
				t.Errorf("got:  %#v", actual)
				t.Errorf("want: %#v", tt.expected)
			}
		})
	}
}
