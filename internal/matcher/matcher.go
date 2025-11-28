// Package matcher defines the rules for matching different types of runes
package matcher

import (
	"fmt"
	"slices"
)

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isAlpha(r rune) bool {
	return isUpper(r) || isLower(r)
}

func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || isDigit(r) || r == '_'
}

func match(r rune, rng [2]rune) (bool, error) {
	if rng[1] > rng[0] {
		return false, fmt.Errorf("range values reversed")
	}
	return r >= rng[0] && r <= rng[1], nil
}

type Matcher interface {
	Match(r rune) (bool, error)
}

type PredefinedClassMatcher interface {
	Matcher
	isPredefinedClass()
}

type LiteralMatcher struct {
	Literal rune
}

func (l *LiteralMatcher) Match(r rune) (bool, error) {
	return r == l.Literal, nil
}

type CharacterSetMatcher struct {
	IsPositive               bool
	Literals                 []rune
	Ranges                   [][2]rune
	CharacterClassesMatchers []PredefinedClassMatcher
}

func (p *CharacterSetMatcher) Match(r rune) (bool, error) {
	if slices.Contains(p.Literals, r) {
		return p.IsPositive, nil
	}
	for _, rng := range p.Ranges {
		m, err := match(r, rng)
		if err != nil {
			return false, err
		}
		if m {
			return p.IsPositive, nil
		}
	}
	for _, characterClass := range p.CharacterClassesMatchers {
		m, err := characterClass.Match(r)
		if err != nil {
			return false, err
		}
		if m {
			return p.IsPositive, nil
		}
	}
	return !p.IsPositive, nil
}

type WildcardMatcher struct{}

func (w *WildcardMatcher) Match(r rune) (bool, error) {
	return r != '\n', nil
}

type DigitMatcher struct{}

func (d *DigitMatcher) Match(r rune) (bool, error) {
	return isDigit(r), nil
}

func (d *DigitMatcher) isPredefinedClass() {}

type AlphaNumericMatcher struct{}

func (a *AlphaNumericMatcher) Match(r rune) (bool, error) {
	return isAlphaNumeric(r), nil
}

func (a *AlphaNumericMatcher) isPredefinedClass() {}
