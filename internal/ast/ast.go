// Package ast defines the structure of an AST
package ast

import predefinedclass "github.com/mmarchesotti/build-your-own-grep/internal/predefinedclass"

type ASTNode interface {
	isASTNode()
}

type baseASTNode struct{}

func (n *baseASTNode) isASTNode() {}

type CaptureGroupNode struct {
	baseASTNode
	Child      ASTNode
	GroupIndex int
}

type AlternationNode struct {
	baseASTNode
	Left  ASTNode
	Right ASTNode
}

type ConcatenationNode struct {
	baseASTNode
	Left  ASTNode
	Right ASTNode
}

type KleeneClosureNode struct {
	baseASTNode
	Child ASTNode
}

type PositiveClosureNode struct {
	baseASTNode
	Child ASTNode
}

type OptionalNode struct {
	baseASTNode
	Child ASTNode
}

type LiteralNode struct {
	baseASTNode
	Literal rune
}

type CharacterSetNode struct {
	baseASTNode
	IsPositive       bool
	Literals         []rune
	Ranges           [][2]rune
	CharacterClasses []predefinedclass.PredefinedClass
}

type WildcardNode struct {
	baseASTNode
}

type DigitNode struct {
	baseASTNode
}

type AlphaNumericNode struct {
	baseASTNode
}

type StartAnchorNode struct {
	baseASTNode
}

type EndAnchorNode struct {
	baseASTNode
}
