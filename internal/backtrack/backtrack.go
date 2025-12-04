// Package backtrack defines the backtracking logic for backreferences
package backtrack

import (
	"fmt"
	"slices"

	"github.com/mmarchesotti/build-your-own-grep/internal/buildnfa"
	"github.com/mmarchesotti/build-your-own-grep/internal/nfasimulator"
	"github.com/mmarchesotti/build-your-own-grep/internal/parser"
	"github.com/mmarchesotti/build-your-own-grep/internal/token"
)

func Run(line []byte, tokens []token.Token) (match bool, err error) {
	return processTokens(line, 0, tokens, []nfasimulator.Capture{})
}

func processTokens(line []byte, lineIndex int, tokens []token.Token, allCapturedGroups []nfasimulator.Capture) (match bool, err error) {
	// 1. Success Condition: We ran out of tokens to match, meaning we succeeded!
	if len(tokens) == 0 {
		return true, nil
	}

	if lineIndex > len(line) {
		return false, nil
	}

	for captureIndex, t := range tokens {
		backReferenceToken, isBackreferenceToken := t.(*token.BackReference)
		if !isBackreferenceToken {
			continue
		}

		prefixTokens := tokens[:captureIndex]

		// FIX: Define as read-only channel (<-chan) to match return type of Simulate
		var matchesChannel <-chan []nfasimulator.Capture

		// 2. Handle Empty Prefix
		// If the backreference is the first token (e.g. inside recursion for \1\2),
		// prefixTokens is empty. We skip parsing and simulate a single empty match.
		if len(prefixTokens) == 0 {
			// Create a bidirectional temp channel to send the value
			tempChannel := make(chan []nfasimulator.Capture, 1)
			tempChannel <- []nfasimulator.Capture{{Start: 0, End: 0}}
			close(tempChannel)
			// Assign to the read-only variable (valid in Go)
			matchesChannel = tempChannel
		} else {
			tree, captureCount, err := parser.Parse(prefixTokens)
			if err != nil {
				return false, err
			}

			fragment, err := buildnfa.Build(tree)
			if err != nil {
				return false, err
			}

			matchesChannel, err = nfasimulator.Simulate(line[lineIndex:], fragment, captureCount)
			if err != nil {
				return false, fmt.Errorf("invalid pattern: %w", err)
			}
		}

		for match := range matchesChannel {
			adjustedMatch := make([]nfasimulator.Capture, len(match))
			for i, cap := range match {
				adjustedMatch[i] = nfasimulator.Capture{
					Start: cap.Start + lineIndex,
					End:   cap.End + lineIndex,
				}
			}

			currentCaptures := allCapturedGroups
			if len(adjustedMatch) > 1 {
				currentCaptures = slices.Concat(allCapturedGroups, adjustedMatch[1:])
			}

			fragmentEndIndex := adjustedMatch[0].End

			backReferenceGroup, err := getReferencedGroup(currentCaptures, *backReferenceToken)
			if err != nil {
				// If the group doesn't exist, this path fails.
				return false, err
			}

			if !matchBackReference(line, fragmentEndIndex, backReferenceGroup) {
				continue
			}

			backRefLength := backReferenceGroup.End - backReferenceGroup.Start
			newLineIndex := fragmentEndIndex + backRefLength

			restOfPatternMatch, err := processTokens(line, newLineIndex, tokens[captureIndex+1:], currentCaptures)
			if err != nil {
				return false, err
			}
			if restOfPatternMatch {
				return true, nil
			}
		}

		// If we processed a backreference token but found no valid path,
		// we must return false.
		return false, nil
	}

	// BASE CASE: Standard tokens (no backreferences)
	tree, captureCount, err := parser.Parse(tokens)
	if err != nil {
		return false, err
	}

	fragment, err := buildnfa.Build(tree)
	if err != nil {
		return false, err
	}

	captures, err := nfasimulator.Simulate(line[lineIndex:], fragment, captureCount)
	if err != nil {
		return false, fmt.Errorf("invalid pattern: %w", err)
	}

	_, hasMatch := <-captures

	return hasMatch, nil
}

func getReferencedGroup(allCapturedGroups []nfasimulator.Capture, backReferenceToken token.BackReference) (nfasimulator.Capture, error) {
	if backReferenceToken.CaptureIndex > len(allCapturedGroups) {
		return nfasimulator.Capture{}, fmt.Errorf("reference to non-existing group '%d'", backReferenceToken.CaptureIndex)
	}
	if backReferenceToken.CaptureIndex < 1 {
		return nfasimulator.Capture{}, fmt.Errorf("invalid group index '%d'", backReferenceToken.CaptureIndex)
	}
	capturedGroupIndex := backReferenceToken.CaptureIndex - 1
	return allCapturedGroups[capturedGroupIndex], nil
}

func matchBackReference(line []byte, lineIndex int, capture nfasimulator.Capture) bool {
	length := capture.End - capture.Start
	if lineIndex+length > len(line) {
		return false
	}
	for i := range length {
		if line[lineIndex+i] != line[capture.Start+i] {
			return false
		}
	}
	return true
}
