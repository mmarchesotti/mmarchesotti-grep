// Package lexer defines the tokenizing logic
package lexer

import (
	"fmt"
	"strings"

	"github.com/mmarchesotti/build-your-own-grep/internal/predefinedclass"
	"github.com/mmarchesotti/build-your-own-grep/internal/token"
)

func Tokenize(inputPattern string) ([]token.Token, error) {
	tokens := make([]token.Token, 0, len(inputPattern))

	for inputIndex := 0; inputIndex < len(inputPattern); inputIndex++ {
		currentCharacter := inputPattern[inputIndex]
		var newToken token.Token

		switch currentCharacter {
		case '\\':
			if inputIndex+1 >= len(inputPattern) {
				return nil, fmt.Errorf("dangling backslash")
			}
			nextCharacter := inputPattern[inputIndex+1]
			switch nextCharacter {
			case 'd':
				newToken = &token.Digit{}
			case 'w':
				newToken = &token.AlphaNumeric{}
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				newToken = &token.BackReference{CaptureIndex: int(nextCharacter - '0')}
			default:
				newToken = &token.Literal{Literal: rune(nextCharacter)}
			}
			inputIndex += 1
		case '[':
			distanceToClosing := strings.Index(inputPattern[inputIndex:], "]")
			if distanceToClosing == -1 {
				return nil, fmt.Errorf("unmatched character set opener [")
			}

			var setLiterals []rune
			var characterClasses []predefinedclass.PredefinedClass
			setCharacters := inputPattern[inputIndex+1 : inputIndex+distanceToClosing]
			inputIndex += distanceToClosing

			startingSetIndex := 0
			negated := len(setCharacters) > 0 && setCharacters[0] == '^'
			if negated {
				startingSetIndex = 1
			}

			for setIndex := startingSetIndex; setIndex < len(setCharacters); setIndex++ {
				currentGroupCharacter := setCharacters[setIndex]

				if currentGroupCharacter == '\\' {
					if setIndex+1 >= len(setCharacters) {
						return nil, fmt.Errorf("dangling backslash inside character set")
					}
					nextCharacter := setCharacters[setIndex+1]
					switch nextCharacter {
					case 'd':
						characterClasses = append(characterClasses, predefinedclass.ClassDigit)
					case 'w':
						characterClasses = append(characterClasses, predefinedclass.ClassAlphanumeric)
					default:
						setLiterals = append(setLiterals, rune(nextCharacter))
					}
					setIndex += 1
				} else {
					setLiterals = append(setLiterals, rune(setCharacters[setIndex]))
				}
			}

			newToken = &token.CharacterSet{
				IsPositive:       !negated,
				CharacterClasses: characterClasses,
				Literals:         setLiterals,
			}

		case '^':
			newToken = &token.StartAnchor{}
		case '$':
			newToken = &token.EndAnchor{}
		case '*':
			newToken = &token.KleeneClosure{}
		case '+':
			newToken = &token.PositiveClosure{}
		case '?':
			newToken = &token.OptionalQuantifier{}
		case '.':
			newToken = &token.Wildcard{}
		case '|':
			newToken = &token.Alternation{}
		case '(':
			newToken = &token.GroupingOpener{}
		case ')':
			newToken = &token.GroupingCloser{}
		default:
			newToken = &token.Literal{
				Literal: rune(inputPattern[inputIndex]),
			}
		}
		tokens = append(tokens, newToken)
	}

	return tokens, nil
}
