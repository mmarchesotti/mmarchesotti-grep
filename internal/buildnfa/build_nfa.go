// Package buildnfa defines the logic for building an NFA
package buildnfa

import (
	"fmt"

	"github.com/mmarchesotti/build-your-own-grep/internal/ast"
	"github.com/mmarchesotti/build-your-own-grep/internal/matcher"
	"github.com/mmarchesotti/build-your-own-grep/internal/nfa"
	"github.com/mmarchesotti/build-your-own-grep/internal/predefinedclass"
)

func newMatcherFragment(m matcher.Matcher) nfa.Fragment {
	state := nfa.MatcherState{
		Out:     nil,
		Matcher: m,
	}
	return nfa.Fragment{
		Start: &state,
		Out:   []*nfa.State{&state.Out},
	}
}

func processNode(n ast.ASTNode) (nfa.Fragment, error) {
	switch node := n.(type) {
	case *ast.CaptureGroupNode:
		subfragment, err := processNode(node.Child)
		if err != nil {
			return nfa.Fragment{}, err
		}

		startState := &nfa.CaptureStartState{
			GroupIndex: node.GroupIndex,
			Out:        subfragment.Start,
		}
		endState := &nfa.CaptureEndState{
			GroupIndex: node.GroupIndex,
			Out:        nil,
		}

		nfa.SetStates(subfragment.Out, endState)

		return nfa.Fragment{
			Start: startState,
			Out:   []*nfa.State{&endState.Out},
		}, nil
	case *ast.AlternationNode:
		subfragment1, err1 := processNode(node.Left)
		if err1 != nil {
			return nfa.Fragment{}, err1
		}
		subfragment2, err2 := processNode(node.Right)
		if err2 != nil {
			return nfa.Fragment{}, err2
		}
		frag := nfa.Fragment{
			Start: &nfa.SplitState{
				Branch1: subfragment1.Start,
				Branch2: subfragment2.Start,
			},
			Out: append(subfragment1.Out, subfragment2.Out...),
		}
		return frag, nil
	case *ast.ConcatenationNode:
		subfragment1, err1 := processNode(node.Left)
		if err1 != nil {
			return nfa.Fragment{}, err1
		}
		subfragment2, err2 := processNode(node.Right)
		if err2 != nil {
			return nfa.Fragment{}, err2
		}
		nfa.SetStates(subfragment1.Out, subfragment2.Start)
		frag := nfa.Fragment{
			Start: subfragment1.Start,
			Out:   subfragment2.Out,
		}
		return frag, nil
	case *ast.KleeneClosureNode:
		subfragment, err := processNode(node.Child)
		if err != nil {
			return nfa.Fragment{}, err
		}
		split := nfa.SplitState{
			Branch1: subfragment.Start,
			Branch2: nil,
		}
		nfa.SetStates(subfragment.Out, &split)
		frag := nfa.Fragment{
			Start: &split,
			Out:   []*nfa.State{&split.Branch2},
		}
		return frag, nil
	case *ast.PositiveClosureNode:
		subfragment, err := processNode(node.Child)
		if err != nil {
			return nfa.Fragment{}, err
		}
		split := nfa.SplitState{
			Branch1: subfragment.Start,
			Branch2: nil,
		}
		nfa.SetStates(subfragment.Out, &split)
		frag := nfa.Fragment{
			Start: subfragment.Start,
			Out:   []*nfa.State{&split.Branch2},
		}
		return frag, nil
	case *ast.OptionalNode:
		subfragment, err := processNode(node.Child)
		if err != nil {
			return nfa.Fragment{}, err
		}
		split := nfa.SplitState{
			Branch1: subfragment.Start,
			Branch2: nil,
		}
		frag := nfa.Fragment{
			Start: &split,
			Out:   append(subfragment.Out, &split.Branch2),
		}
		return frag, nil
	case *ast.CharacterSetNode:
		var characterClassesMatchers []matcher.PredefinedClassMatcher
		for _, characterClass := range node.CharacterClasses {
			var m matcher.PredefinedClassMatcher
			switch characterClass {
			case predefinedclass.ClassDigit:
				m = &matcher.DigitMatcher{}
			case predefinedclass.ClassAlphanumeric:
				m = &matcher.AlphaNumericMatcher{}
			}
			characterClassesMatchers = append(characterClassesMatchers, m)
		}
		characterSetMatcher := &matcher.CharacterSetMatcher{
			IsPositive:               node.IsPositive,
			Literals:                 node.Literals,
			Ranges:                   node.Ranges,
			CharacterClassesMatchers: characterClassesMatchers,
		}
		return newMatcherFragment(characterSetMatcher), nil
	case *ast.LiteralNode:
		return newMatcherFragment(&matcher.LiteralMatcher{Literal: node.Literal}), nil
	case *ast.WildcardNode:
		return newMatcherFragment(&matcher.WildcardMatcher{}), nil
	case *ast.DigitNode:
		return newMatcherFragment(&matcher.DigitMatcher{}), nil
	case *ast.AlphaNumericNode:
		return newMatcherFragment(&matcher.AlphaNumericMatcher{}), nil
	case *ast.StartAnchorNode:
		s := &nfa.StartAnchorState{
			Out: nil,
		}
		frag := nfa.Fragment{
			Start: s,
			Out:   []*nfa.State{&s.Out},
		}
		return frag, nil
	case *ast.EndAnchorNode:
		s := &nfa.EndAnchorState{
			Out: nil,
		}
		frag := nfa.Fragment{
			Start: s,
			Out:   []*nfa.State{&s.Out},
		}
		return frag, nil
	default:
		return nfa.Fragment{}, fmt.Errorf("unexpected node type %T", node)
	}
}

func Build(tree ast.ASTNode) (nfa.Fragment, error) {
	mainFrag, err := processNode(tree)
	if err != nil {
		return nfa.Fragment{}, err
	}

	startState := &nfa.CaptureStartState{
		GroupIndex: 0,
		Out:        mainFrag.Start,
	}
	endState := &nfa.CaptureEndState{
		GroupIndex: 0,
		Out:        nil,
	}
	nfa.SetStates(mainFrag.Out, endState)

	acceptingState := &nfa.AcceptingState{}
	nfa.SetStates([]*nfa.State{&endState.Out}, acceptingState)

	finalFragment := nfa.Fragment{
		Start: startState,
		Out:   []*nfa.State{},
	}

	return finalFragment, nil
}
