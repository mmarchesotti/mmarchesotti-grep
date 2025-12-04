package backtrack

import (
	"testing"

	"github.com/mmarchesotti/build-your-own-grep/internal/token"
)

func TestRun_Backreferences(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		tokens    []token.Token
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "Simple repetition (cat)\\1",
			line: "catcat",
			// Represents: (cat)\1
			tokens: []token.Token{
				&token.GroupingOpener{},
				&token.Literal{Literal: 'c'},
				&token.Literal{Literal: 'a'},
				&token.Literal{Literal: 't'},
				&token.GroupingCloser{},
				&token.BackReference{CaptureIndex: 1},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "No match repetition (cat)\\1",
			line: "catdog",
			// Represents: (cat)\1 matches "catcat", input is "catdog"
			tokens: []token.Token{
				&token.GroupingOpener{},
				&token.Literal{Literal: 'c'},
				&token.Literal{Literal: 'a'},
				&token.Literal{Literal: 't'},
				&token.GroupingCloser{},
				&token.BackReference{CaptureIndex: 1},
			},
			wantMatch: false,
			wantErr:   false,
		},
		{
			name: "Backref with suffix (a)\\1b",
			line: "aab",
			// Represents: (a)\1b
			tokens: []token.Token{
				&token.GroupingOpener{},
				&token.Literal{Literal: 'a'},
				&token.GroupingCloser{},
				&token.BackReference{CaptureIndex: 1},
				&token.Literal{Literal: 'b'},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "Invalid Group Index",
			line: "hello",
			// Represents: h\5
			tokens: []token.Token{
				&token.Literal{Literal: 'h'},
				&token.BackReference{CaptureIndex: 5}, // Group 5 doesn't exist
			},
			wantMatch: false,
			wantErr:   true,
		},
		{
			name: "Double Backreference (a)(b)\\1\\2",
			line: "abab",
			// Represents: (a)(b)\1\2 -> a b a b
			tokens: []token.Token{
				&token.GroupingOpener{},
				&token.Literal{Literal: 'a'},
				&token.GroupingCloser{},
				&token.GroupingOpener{},
				&token.Literal{Literal: 'b'},
				&token.GroupingCloser{},
				&token.BackReference{CaptureIndex: 1},
				&token.BackReference{CaptureIndex: 2},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "Complex Nested (a(b))\\2\\1",
			line: "abbab",
			// Represents: (a(b))\2\1
			// Group 1: (a(b)) -> matches "ab"
			// Group 2: (b)    -> matches "b"
			// Pattern expects: "ab" + \2("b") + \1("ab") -> "abbab"
			tokens: []token.Token{
				&token.GroupingOpener{}, // Group 1 start
				&token.Literal{Literal: 'a'},
				&token.GroupingOpener{}, // Group 2 start
				&token.Literal{Literal: 'b'},
				&token.GroupingCloser{}, // Group 2 end
				&token.GroupingCloser{}, // Group 1 end
				&token.BackReference{CaptureIndex: 2},
				&token.BackReference{CaptureIndex: 1},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := Run([]byte(tt.line), tt.tokens)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if match != tt.wantMatch {
				t.Errorf("Run() match = %v, want %v", match, tt.wantMatch)
			}
		})
	}
}
