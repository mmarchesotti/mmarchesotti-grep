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
	for captureIndex, t := range tokens {
		backReferenceToken, isBackreferenceToken := t.(*token.BackReference)
		if !isBackreferenceToken {
			continue
		}

		tree, captureCount, err := parser.Parse(tokens)
		if err != nil {
			return false, err
		}

		fragment, err := buildnfa.Build(tree)
		if err != nil {
			return false, err
		}

		matchesChannel, err := nfasimulator.Simulate(line, fragment, captureCount)
		if err != nil {
			return false, fmt.Errorf("invalid pattern: %w", err)
		}

		for match := range matchesChannel {
			localLineIndex := lineIndex
			if len(match) > 1 {
				allCapturedGroups = slices.Concat(allCapturedGroups, match[1:])
			}
			allCapturedGroups[0].End = match[0].End
			backReferenceGroup, err := getReferencedGroup(allCapturedGroups, *backReferenceToken)
			if err != nil {
				return false, err
			}
			backReferenceMatch := matchBackReference(line, localLineIndex, backReferenceGroup)
			if !backReferenceMatch {
				continue
			}
			localLineIndex += backReferenceGroup.End - backReferenceGroup.Start
			restOfPatternMatch, err := processTokens(line, localLineIndex, tokens[captureIndex+1:], allCapturedGroups)
			if err != nil {
				return false, err
			}
			if restOfPatternMatch {
				return true, nil
			}
		}
		return false, nil
	}

	tree, captureCount, err := parser.Parse(tokens)
	if err != nil {
		return false, err
	}

	fragment, err := buildnfa.Build(tree)
	if err != nil {
		return false, err
	}

	captures, err := nfasimulator.Simulate(line, fragment, captureCount)
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
	capturedGroupIndex := backReferenceToken.CaptureIndex - 1
	return allCapturedGroups[capturedGroupIndex], nil
}

func matchBackReference(line []byte, lineIndex int, capture nfasimulator.Capture) bool {
	for captureIndex := 0; captureIndex < capture.End-capture.Start; captureIndex++ {
		if lineIndex+captureIndex > len(line) || line[capture.Start+captureIndex] != line[lineIndex+captureIndex] {
			return false
		}
	}
	return true
}
