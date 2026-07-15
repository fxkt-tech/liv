package constraints

import (
	"errors"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

func TestMaxSimilarityUsesSymmetricDurationWeightedOverlap(t *testing.T) {
	accepted := candidate(
		"accepted",
		[]VideoSelection{{SlotID: "slot", Path: "/video.mp4", SourceRange: constraintRange(0, 10*time.Second)}},
		nil,
	).Summary()
	history := NewHistoryView([]AcceptedSummary{accepted})
	current := candidate(
		"current",
		[]VideoSelection{{SlotID: "slot", Path: "/video.mp4", SourceRange: constraintRange(5*time.Second, 10*time.Second)}},
		nil,
	)

	constraint, err := NewMaxSimilarity("similarity", 0.5)
	if err != nil {
		t.Fatalf("NewMaxSimilarity() error = %v", err)
	}
	decision, err := constraint.Check(current, history)
	if err != nil || !decision.Accepted {
		t.Fatalf("boundary decision = %+v, err = %v; want accepted", decision, err)
	}
	constraint, err = NewMaxSimilarity("similarity", 0.499)
	if err != nil {
		t.Fatalf("NewMaxSimilarity() error = %v", err)
	}
	decision, err = constraint.Check(current, history)
	if err != nil || decision.Accepted || decision.Reason != ReasonMaxSimilarity {
		t.Fatalf("over-limit decision = %+v, err = %v", decision, err)
	}
}

func TestMaxSimilarityIgnoresDifferentPathsAndBGM(t *testing.T) {
	acceptedBGM := BGMSelection{BGMID: "bgm-a", Path: "/same.mp3"}
	currentBGM := BGMSelection{BGMID: "bgm-b", Path: "/same.mp3"}
	history := NewHistoryView([]AcceptedSummary{candidate(
		"accepted",
		[]VideoSelection{{SlotID: "slot", Path: "/a.mp4", SourceRange: constraintRange(0, 5*time.Second)}},
		&acceptedBGM,
	).Summary()})
	current := candidate(
		"current",
		[]VideoSelection{{SlotID: "slot", Path: "/b.mp4", SourceRange: constraintRange(0, 5*time.Second)}},
		&currentBGM,
	)

	constraint, err := NewMaxSimilarity("similarity", 0)
	if err != nil {
		t.Fatalf("NewMaxSimilarity() error = %v", err)
	}
	decision, err := constraint.Check(current, history)
	if err != nil || !decision.Accepted {
		t.Fatalf("decision = %+v, err = %v; different video paths must have zero similarity", decision, err)
	}
}

func TestMaxVideoAssetUsesCountsRangesFromSamePath(t *testing.T) {
	history := NewHistoryView([]AcceptedSummary{candidate(
		"accepted",
		[]VideoSelection{{SlotID: "slot-a", Path: "/same.mp4", SourceRange: constraintRange(0, time.Second)}},
		nil,
	).Summary()})
	current := candidate(
		"current",
		[]VideoSelection{{SlotID: "slot-b", Path: "/same.mp4", SourceRange: constraintRange(time.Second, time.Second)}},
		nil,
	)

	constraint, err := NewMaxVideoAssetUses("uses", 2)
	if err != nil {
		t.Fatalf("NewMaxVideoAssetUses() error = %v", err)
	}
	decision, err := constraint.Check(current, history)
	if err != nil || !decision.Accepted {
		t.Fatalf("boundary decision = %+v, err = %v; want accepted", decision, err)
	}
	constraint, err = NewMaxVideoAssetUses("uses", 1)
	if err != nil {
		t.Fatalf("NewMaxVideoAssetUses() error = %v", err)
	}
	decision, err = constraint.Check(current, history)
	if err != nil || decision.Accepted || decision.Reason != ReasonMaxVideoAssetUses {
		t.Fatalf("over-limit decision = %+v, err = %v", decision, err)
	}
}

func TestMaxBGMUsesIsIndependent(t *testing.T) {
	bgm := BGMSelection{BGMID: "bgm", Path: "/bgm.mp3"}
	history := NewHistoryView([]AcceptedSummary{candidate("accepted", nil, &bgm).Summary()})
	current := candidate("current", nil, &bgm)

	constraint, err := NewMaxBGMUses("uses", 2)
	if err != nil {
		t.Fatalf("NewMaxBGMUses() error = %v", err)
	}
	decision, err := constraint.Check(current, history)
	if err != nil || !decision.Accepted {
		t.Fatalf("boundary decision = %+v, err = %v; want accepted", decision, err)
	}
	constraint, err = NewMaxBGMUses("uses", 1)
	if err != nil {
		t.Fatalf("NewMaxBGMUses() error = %v", err)
	}
	decision, err = constraint.Check(current, history)
	if err != nil || decision.Accepted || decision.Reason != ReasonMaxBGMUses {
		t.Fatalf("over-limit decision = %+v, err = %v", decision, err)
	}
}

func TestViewsAndFunctionConstraintAreImmutable(t *testing.T) {
	videos := []VideoSelection{{SlotID: "slot", Path: "/a.mp4", SourceRange: constraintRange(0, time.Second)}}
	transitions := []TransitionSelection{{JoinID: "join", TransitionID: "fade"}}
	view := NewCandidateView("candidate", videos, transitions, nil)
	videos[0].Path = "/mutated.mp4"
	transitions[0].TransitionID = "mutated"
	returned := view.Videos()
	returned[0].Path = "/also-mutated.mp4"
	if got := view.Videos()[0].Path; got != "/a.mp4" {
		t.Fatalf("immutable path = %q", got)
	}
	returnedTransitions := view.Summary().Transitions()
	returnedTransitions[0].TransitionID = "also-mutated"
	if got := view.Summary().Transitions()[0].TransitionID; got != "fade" {
		t.Fatalf("immutable transition = %q", got)
	}

	sentinel := errors.New("sentinel")
	constraint, err := NewFunc("custom", "fingerprint", func(CandidateView, HistoryView) (Decision, error) {
		return Decision{}, sentinel
	})
	if err != nil {
		t.Fatalf("NewFunc() error = %v", err)
	}
	_, err = constraint.Check(view, NewHistoryView(nil))
	if !errors.Is(err, sentinel) {
		t.Fatalf("Check() error = %v, want sentinel", err)
	}
}

func TestBuiltInConstructorsRejectInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{name: "empty similarity ID", call: func() error { _, err := NewMaxSimilarity("", 0.5); return err }},
		{name: "invalid similarity", call: func() error { _, err := NewMaxSimilarity("id", 2); return err }},
		{name: "invalid video uses", call: func() error { _, err := NewMaxVideoAssetUses("id", 0); return err }},
		{name: "invalid BGM uses", call: func() error { _, err := NewMaxBGMUses("id", 0); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err == nil {
				t.Fatal("constructor error = nil")
			}
		})
	}
}

func candidate(fingerprint string, videos []VideoSelection, bgm *BGMSelection) CandidateView {
	return NewCandidateView(fingerprint, videos, nil, bgm)
}

func constraintRange(start, duration time.Duration) ffcut.TimeRange {
	return ffcut.TimeRange{Start: constraintDuration(start), Duration: constraintDuration(duration)}
}

func constraintDuration(value time.Duration) ffcut.Duration {
	result, err := ffcut.NewDuration(value)
	if err != nil {
		panic(err)
	}
	return result
}
