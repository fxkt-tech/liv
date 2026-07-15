package ffvmix

import constraintpkg "github.com/fxkt-tech/liv/ffvmix/constraints"

type CandidateView = constraintpkg.CandidateView
type HistoryView = constraintpkg.HistoryView
type Decision = constraintpkg.Decision
type Constraint = constraintpkg.Constraint
type ConstraintFunc = constraintpkg.ConstraintFunc
type VideoSelectionView = constraintpkg.VideoSelection
type TransitionSelectionView = constraintpkg.TransitionSelection
type BGMSelectionView = constraintpkg.BGMSelection

const (
	ReasonMaxSimilarity     = constraintpkg.ReasonMaxSimilarity
	ReasonMaxVideoAssetUses = constraintpkg.ReasonMaxVideoAssetUses
	ReasonMaxBGMUses        = constraintpkg.ReasonMaxBGMUses
)

// Accept approves a candidate from a custom constraint function.
func Accept() Decision {
	return constraintpkg.Accept()
}

// Reject rejects a candidate with a stable, machine-readable reason.
func Reject(reason string) Decision {
	return constraintpkg.Reject(reason)
}

// NewConstraint adapts a plain plugin function to the constraint interface.
func NewConstraint(id, fingerprint string, check ConstraintFunc) (Constraint, error) {
	return constraintpkg.NewFunc(id, fingerprint, check)
}
