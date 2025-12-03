// Package token defines the types of tokens used when tokenizing an input
package token

import "github.com/mmarchesotti/build-your-own-grep/internal/predefinedclass"

type TokenType string

// --- Helper Functions ---

func IsAlternation(t Token) bool {
	_, ok := t.(*Alternation)
	return ok
}

func IsGroupingOpener(t Token) bool {
	_, ok := t.(*GroupingOpener)
	return ok
}

func IsGroupingCloser(t Token) bool {
	_, ok := t.(*GroupingCloser)
	return ok
}

func CanConcatenate(t Token) bool {
	switch t.(type) {
	case *Literal, *CharacterSet, *Wildcard, *Digit, *AlphaNumeric,
		*StartAnchor, *EndAnchor, *GroupingOpener:
		return true
	default:
		return false
	}
}

func IsUnaryOperator(t Token) bool {
	switch t.(type) {
	case *OptionalQuantifier, *KleeneClosure, *PositiveClosure:
		return true
	default:
		return false
	}
}

func IsAtom(t Token) bool {
	switch t.(type) {
	case *Literal, *CharacterSet, *Wildcard, *Digit, *AlphaNumeric:
		return true
	default:
		return false
	}
}

// --- Token Interface and Structs ---

type Token interface {
	getType() string
}

type baseToken struct {
	pType TokenType
}

func (token *baseToken) getType() string {
	return string(token.pType)
}

type (
	Concatenation      struct{ baseToken }
	KleeneClosure      struct{ baseToken }
	PositiveClosure    struct{ baseToken }
	OptionalQuantifier struct{ baseToken }
	Alternation        struct{ baseToken }
	StartAnchor        struct{ baseToken }
	EndAnchor          struct{ baseToken }
	Wildcard           struct{ baseToken }
	GroupingOpener     struct{ baseToken }
	GroupingCloser     struct{ baseToken }
	Digit              struct{ baseToken }
	AlphaNumeric       struct{ baseToken }
	Literal            struct {
		baseToken
		Literal rune
	}
	CharacterSet struct {
		baseToken
		IsPositive       bool
		Literals         []rune
		Ranges           [][2]rune
		CharacterClasses []predefinedclass.PredefinedClass
	}
	BackReference struct {
		baseToken
		CaptureIndex int
	}
)
