// Package predefinedclass defines the types of predefined character classes
package predefinedclass

type PredefinedClass int

const (
	ClassDigit PredefinedClass = iota
	ClassAlphanumeric
	ClassWhitespace
)
