// Package nfasimulator defines the rules for walking(simulating) an NFA
package nfasimulator

import (
	"fmt"
	"unicode/utf8"

	"github.com/mmarchesotti/build-your-own-grep/internal/nfa"
)

type Capture struct {
	Start int
	End   int
}

type thread struct {
	state     nfa.State
	lineIndex int
	captures  []Capture
}

func (t *thread) key() string {
	return fmt.Sprintf("%p-%d", t.state, t.lineIndex)
}

type task struct {
	isRevert bool
	thread   thread
	undoLog  []undoEntry
}

type undoEntry struct {
	captureIndex int
	isStart      bool
	oldValue     int
}

func Simulate(line []byte, fragment nfa.Fragment, captureCount int) (<-chan []Capture, error) {
	out := make(chan []Capture)

	go func() {
		defer close(out)

		searchIndex := 0
		for searchIndex <= len(line) {

			stream := findMatchAt(fragment.Start, line, searchIndex, captureCount)

			foundAnyAtThisIndex := false
			maxEndIndex := searchIndex

			for match := range stream {
				foundAnyAtThisIndex = true

				if match[0].End > maxEndIndex {
					maxEndIndex = match[0].End
				}

				out <- match
			}

			if foundAnyAtThisIndex {
				if maxEndIndex > searchIndex {
					searchIndex = maxEndIndex
				} else {
					searchIndex++
				}
			} else {
				searchIndex++
			}
		}
	}()

	return out, nil
}

func findMatchAt(startState nfa.State, line []byte, startIndex int, captureCount int) <-chan []Capture {
	out := make(chan []Capture)

	go func() {
		defer close(out)
		stack := []task{}

		initialCaptures := make([]Capture, captureCount)
		for i := range initialCaptures {
			initialCaptures[i] = Capture{Start: -1, End: -1}
		}

		initialThread := thread{
			state:     startState,
			lineIndex: startIndex,
			captures:  initialCaptures,
		}
		stack = append(stack, task{
			isRevert: false,
			thread:   initialThread,
			undoLog:  nil,
		})

		visited := make(map[string]bool)

		for len(stack) > 0 {
			currentTask := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if currentTask.isRevert {
				for _, entry := range currentTask.undoLog {
					if entry.isStart {
						currentTask.thread.captures[entry.captureIndex].Start = entry.oldValue
					} else {
						currentTask.thread.captures[entry.captureIndex].End = entry.oldValue
					}
				}
				continue
			}

			threadKey := currentTask.thread.key()
			if visited[threadKey] {
				continue
			}
			visited[threadKey] = true

			currentState := currentTask.thread.state
			switch st := currentState.(type) {
			case *nfa.AcceptingState:
				resultCopy := copyCaptures(currentTask.thread.captures)
				out <- resultCopy
				continue
			case *nfa.MatcherState:
				if currentTask.thread.lineIndex < len(line) {
					r, size := utf8.DecodeRune(line[currentTask.thread.lineIndex:])
					match, _ := st.Matcher.Match(r)
					if match {
						nextThread := thread{
							state:     st.Out,
							lineIndex: currentTask.thread.lineIndex + size,
							captures:  currentTask.thread.captures,
						}
						stack = append(stack, task{
							isRevert: false,
							thread:   nextThread,
						})
					}
				}
			case *nfa.SplitState:
				thread1 := thread{
					state:     st.Branch1,
					lineIndex: currentTask.thread.lineIndex,
					captures:  currentTask.thread.captures,
				}
				thread2 := thread{
					state:     st.Branch2,
					lineIndex: currentTask.thread.lineIndex,
					captures:  currentTask.thread.captures,
				}
				stack = append(stack, task{
					isRevert: false,
					thread:   thread2,
				})
				stack = append(stack, task{
					isRevert: false,
					thread:   thread1,
				})
			case *nfa.CaptureStartState:
				undo := undoEntry{
					captureIndex: st.GroupIndex,
					isStart:      true,
					oldValue:     currentTask.thread.captures[st.GroupIndex].Start,
				}

				currentTask.thread.captures[st.GroupIndex].Start = currentTask.thread.lineIndex

				nextThread := thread{
					state:     st.Out,
					lineIndex: currentTask.thread.lineIndex,
					captures:  currentTask.thread.captures,
				}

				stack = append(stack, task{
					isRevert: true,
					thread:   currentTask.thread, undoLog: []undoEntry{undo},
				})
				stack = append(stack, task{
					isRevert: false,
					thread:   nextThread,
				})
			case *nfa.CaptureEndState:
				undo := undoEntry{
					captureIndex: st.GroupIndex,
					isStart:      false,
					oldValue:     currentTask.thread.captures[st.GroupIndex].End,
				}
				currentTask.thread.captures[st.GroupIndex].End = currentTask.thread.lineIndex

				nextThread := thread{
					state:     st.Out,
					lineIndex: currentTask.thread.lineIndex,
					captures:  currentTask.thread.captures,
				}

				stack = append(stack, task{
					isRevert: true,
					thread:   currentTask.thread, undoLog: []undoEntry{undo},
				})
				stack = append(stack, task{
					isRevert: false,
					thread:   nextThread,
				})
			case *nfa.StartAnchorState:
				if currentTask.thread.lineIndex == 0 {
					nextThread := thread{
						state:     st.Out,
						lineIndex: currentTask.thread.lineIndex,
						captures:  currentTask.thread.captures,
					}
					stack = append(stack, task{
						isRevert: false,
						thread:   nextThread,
					})
				}
			case *nfa.EndAnchorState:
				if currentTask.thread.lineIndex == len(line) {
					nextThread := thread{
						state:     st.Out,
						lineIndex: currentTask.thread.lineIndex,
						captures:  currentTask.thread.captures,
					}
					stack = append(stack, task{
						isRevert: false,
						thread:   nextThread,
					})
				}
			}
		}
	}()

	return nil
}

func copyCaptures(src []Capture) []Capture {
	dst := make([]Capture, len(src))

	copy(dst, src)

	return dst
}
