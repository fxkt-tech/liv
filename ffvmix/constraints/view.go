// Package constraints defines the pure, read-only FFVMix constraint boundary
// and the built-in history constraints.
package constraints

import "github.com/fxkt-tech/liv/ffcut"

// VideoSelection describes one selected source window in one template slot.
type VideoSelection struct {
	SlotID           ffcut.ID
	VideoID          ffcut.ID
	Path             string
	AssetFingerprint string
	SourceRange      ffcut.TimeRange
	TimelineDuration ffcut.Duration
}

// TransitionSelection describes one selected transition in one template join.
type TransitionSelection struct {
	JoinID       ffcut.ID
	TransitionID ffcut.ID
	Kind         ffcut.TransitionKind
	Duration     ffcut.Duration
}

// BGMSelection describes the optional selected background-music asset.
type BGMSelection struct {
	DimensionID      ffcut.ID
	BGMID            ffcut.ID
	Path             string
	AssetFingerprint string
}

// CandidateView is an immutable projection of a combination before its full
// ffcut.Project is allocated.
type CandidateView struct {
	fingerprint string
	videos      []VideoSelection
	transitions []TransitionSelection
	bgm         *BGMSelection
}

// NewCandidateView constructs an owned, immutable candidate projection.
// It is primarily intended for the ffvmix generator.
func NewCandidateView(
	fingerprint string,
	videos []VideoSelection,
	transitions []TransitionSelection,
	bgm *BGMSelection,
) CandidateView {
	view := CandidateView{
		fingerprint: fingerprint,
		videos:      append([]VideoSelection(nil), videos...),
		transitions: append([]TransitionSelection(nil), transitions...),
	}
	if bgm != nil {
		selected := *bgm
		view.bgm = &selected
	}
	return view
}

// Fingerprint identifies the exact slot, transition and BGM combination.
func (c CandidateView) Fingerprint() string {
	return c.fingerprint
}

// Videos returns the selected videos in template slot order.
func (c CandidateView) Videos() []VideoSelection {
	return append([]VideoSelection(nil), c.videos...)
}

// Transitions returns the selected transitions in template join order.
func (c CandidateView) Transitions() []TransitionSelection {
	return append([]TransitionSelection(nil), c.transitions...)
}

// BGM returns a copy of the selected BGM, or nil when the template has no BGM
// pool.
func (c CandidateView) BGM() *BGMSelection {
	if c.bgm == nil {
		return nil
	}
	selected := *c.bgm
	return &selected
}

// AcceptedSummary is the immutable history projection retained by Generator.
type AcceptedSummary struct {
	fingerprint string
	videos      []VideoSelection
	transitions []TransitionSelection
	bgm         *BGMSelection
}

// Summary produces the history projection for an accepted candidate.
func (c CandidateView) Summary() AcceptedSummary {
	summary := AcceptedSummary{
		fingerprint: c.fingerprint,
		videos:      append([]VideoSelection(nil), c.videos...),
		transitions: append([]TransitionSelection(nil), c.transitions...),
	}
	if c.bgm != nil {
		selected := *c.bgm
		summary.bgm = &selected
	}
	return summary
}

// Fingerprint identifies the accepted combination.
func (s AcceptedSummary) Fingerprint() string {
	return s.fingerprint
}

// Videos returns the accepted videos in template slot order.
func (s AcceptedSummary) Videos() []VideoSelection {
	return append([]VideoSelection(nil), s.videos...)
}

// Transitions returns the accepted transitions in template join order.
func (s AcceptedSummary) Transitions() []TransitionSelection {
	return append([]TransitionSelection(nil), s.transitions...)
}

// BGM returns a copy of the accepted BGM selection.
func (s AcceptedSummary) BGM() *BGMSelection {
	if s.bgm == nil {
		return nil
	}
	selected := *s.bgm
	return &selected
}

// HistoryView is an immutable view over accepted summaries.
type HistoryView struct {
	accepted []AcceptedSummary
}

// NewHistoryView constructs a read-only history projection. AcceptedSummary
// already owns its nested state, so only the outer slice needs copying.
func NewHistoryView(accepted []AcceptedSummary) HistoryView {
	return HistoryView{accepted: append([]AcceptedSummary(nil), accepted...)}
}

// Len returns the number of accepted outputs.
func (h HistoryView) Len() int {
	return len(h.accepted)
}

// Accepted returns a copy of the accepted summary slice. Each summary exposes
// nested collections only through copy-returning getters.
func (h HistoryView) Accepted() []AcceptedSummary {
	return append([]AcceptedSummary(nil), h.accepted...)
}
