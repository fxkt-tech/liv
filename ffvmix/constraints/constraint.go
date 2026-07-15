package constraints

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/fxkt-tech/liv/ffcut"
)

const (
	ReasonMaxSimilarity     = "max_similarity"
	ReasonMaxVideoAssetUses = "max_video_asset_uses"
	ReasonMaxBGMUses        = "max_bgm_uses"
)

// Decision is the complete result of a pure constraint check.
type Decision struct {
	Accepted bool
	Reason   string
}

// Accept approves a candidate.
func Accept() Decision {
	return Decision{Accepted: true}
}

// Reject rejects a candidate with a stable, machine-readable reason.
func Reject(reason string) Decision {
	return Decision{Reason: reason}
}

// Constraint is a pure candidate predicate. Implementations must not mutate
// external state; Generator alone owns accepted-history updates.
type Constraint interface {
	ID() string
	Fingerprint() string
	Check(CandidateView, HistoryView) (Decision, error)
}

// ConstraintFunc adapts a plain plugin function to the constraint contract.
type ConstraintFunc func(CandidateView, HistoryView) (Decision, error)

type functionConstraint struct {
	id          string
	fingerprint string
	check       ConstraintFunc
}

// NewFunc creates a constraint from a stable identity, configuration
// fingerprint and plugin function.
func NewFunc(id, fingerprint string, check ConstraintFunc) (Constraint, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("constraint id must not be empty")
	}
	if strings.TrimSpace(fingerprint) == "" {
		return nil, fmt.Errorf("constraint fingerprint must not be empty")
	}
	if check == nil {
		return nil, fmt.Errorf("constraint function must not be nil")
	}
	return functionConstraint{id: id, fingerprint: fingerprint, check: check}, nil
}

func (c functionConstraint) ID() string          { return c.id }
func (c functionConstraint) Fingerprint() string { return c.fingerprint }
func (c functionConstraint) Check(candidate CandidateView, history HistoryView) (Decision, error) {
	return c.check(candidate, history)
}

type maxSimilarity struct {
	id          string
	maximum     float64
	fingerprint string
}

// NewMaxSimilarity rejects a candidate when its duration-weighted source
// overlap with any accepted output exceeds maximum.
func NewMaxSimilarity(id string, maximum float64) (Constraint, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("constraint id must not be empty")
	}
	if maximum < 0 || maximum > 1 || math.IsNaN(maximum) || math.IsInf(maximum, 0) {
		return nil, fmt.Errorf("maximum similarity must be between 0 and 1")
	}
	return maxSimilarity{
		id:          id,
		maximum:     maximum,
		fingerprint: builtinFingerprint(ReasonMaxSimilarity, strconv.FormatFloat(maximum, 'g', -1, 64)),
	}, nil
}

func (c maxSimilarity) ID() string          { return c.id }
func (c maxSimilarity) Fingerprint() string { return c.fingerprint }
func (c maxSimilarity) Check(candidate CandidateView, history HistoryView) (Decision, error) {
	for _, accepted := range history.accepted {
		if similarity(candidate.videos, accepted.videos) > c.maximum {
			return Reject(ReasonMaxSimilarity), nil
		}
	}
	return Accept(), nil
}

type maxVideoAssetUses struct {
	id          string
	maximum     int
	fingerprint string
}

// NewMaxVideoAssetUses limits selection occurrences by normalized local video
// path. Multiple source ranges from the same file share one counter.
func NewMaxVideoAssetUses(id string, maximum int) (Constraint, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("constraint id must not be empty")
	}
	if maximum <= 0 {
		return nil, fmt.Errorf("maximum video asset uses must be positive")
	}
	return maxVideoAssetUses{
		id:          id,
		maximum:     maximum,
		fingerprint: builtinFingerprint(ReasonMaxVideoAssetUses, strconv.Itoa(maximum)),
	}, nil
}

func (c maxVideoAssetUses) ID() string          { return c.id }
func (c maxVideoAssetUses) Fingerprint() string { return c.fingerprint }
func (c maxVideoAssetUses) Check(candidate CandidateView, history HistoryView) (Decision, error) {
	uses := make(map[string]int)
	for _, accepted := range history.accepted {
		for _, video := range accepted.videos {
			uses[video.Path]++
		}
	}
	for _, video := range candidate.videos {
		uses[video.Path]++
		if uses[video.Path] > c.maximum {
			return Reject(ReasonMaxVideoAssetUses), nil
		}
	}
	return Accept(), nil
}

type maxBGMUses struct {
	id          string
	maximum     int
	fingerprint string
}

// NewMaxBGMUses limits accepted selections by normalized local BGM path.
func NewMaxBGMUses(id string, maximum int) (Constraint, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("constraint id must not be empty")
	}
	if maximum <= 0 {
		return nil, fmt.Errorf("maximum BGM uses must be positive")
	}
	return maxBGMUses{
		id:          id,
		maximum:     maximum,
		fingerprint: builtinFingerprint(ReasonMaxBGMUses, strconv.Itoa(maximum)),
	}, nil
}

func (c maxBGMUses) ID() string          { return c.id }
func (c maxBGMUses) Fingerprint() string { return c.fingerprint }
func (c maxBGMUses) Check(candidate CandidateView, history HistoryView) (Decision, error) {
	if candidate.bgm == nil {
		return Accept(), nil
	}
	uses := 1
	for _, accepted := range history.accepted {
		if accepted.bgm != nil && accepted.bgm.Path == candidate.bgm.Path {
			uses++
		}
	}
	if uses > c.maximum {
		return Reject(ReasonMaxBGMUses), nil
	}
	return Accept(), nil
}

func similarity(left, right []VideoSelection) float64 {
	rightBySlot := make(map[string]VideoSelection, len(right))
	for _, video := range right {
		rightBySlot[string(video.SlotID)] = video
	}

	var overlap float64
	var comparable float64
	for _, video := range left {
		other, exists := rightBySlot[string(video.SlotID)]
		if !exists {
			comparable += float64(video.SourceRange.Duration)
			continue
		}
		leftDuration := float64(video.SourceRange.Duration)
		rightDuration := float64(other.SourceRange.Duration)
		if rightDuration > leftDuration {
			comparable += rightDuration
		} else {
			comparable += leftDuration
		}
		if video.Path != other.Path {
			continue
		}
		overlap += float64(overlapDuration(video.SourceRange, other.SourceRange))
	}
	if comparable <= 0 {
		return 0
	}
	return overlap / comparable
}

func overlapDuration(left, right ffcut.TimeRange) int64 {
	leftEnd, leftErr := left.End()
	rightEnd, rightErr := right.End()
	if leftErr != nil || rightErr != nil {
		return 0
	}
	start := left.Start
	if right.Start > start {
		start = right.Start
	}
	end := leftEnd
	if rightEnd < end {
		end = rightEnd
	}
	if end <= start {
		return 0
	}
	return int64(end - start)
}

func builtinFingerprint(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte(strconv.Itoa(len(part))))
		_, _ = hash.Write([]byte{':'})
		_, _ = hash.Write([]byte(part))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
