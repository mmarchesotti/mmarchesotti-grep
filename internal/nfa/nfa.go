// Package nfa defines the structure of an NFA
package nfa

import "github.com/mmarchesotti/build-your-own-grep/internal/matcher"

type Fragment struct {
	Start State
	Out   []*State
}

func SetStates(out []*State, start State) {
	for _, o := range out {
		*o = start
	}
}

type State interface {
	isState()
}

type BaseState struct{}

func (s *BaseState) isState() {}

type SplitState struct {
	BaseState
	Branch1 State
	Branch2 State
}

type MatcherState struct {
	BaseState
	Out     State
	Matcher matcher.Matcher
}

type CaptureStartState struct {
	BaseState
	Out        State
	GroupIndex int
}

type CaptureEndState struct {
	BaseState
	Out        State
	GroupIndex int
}

type StartAnchorState struct {
	BaseState
	Out State
}

type EndAnchorState struct {
	BaseState
	Out State
}

type AcceptingState struct {
	BaseState
}
