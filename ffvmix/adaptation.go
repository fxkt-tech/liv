package ffvmix

import (
	"fmt"
	"math"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

type AdaptationKind string

const (
	AdaptationNatural    AdaptationKind = "natural"
	AdaptationSpeedUp    AdaptationKind = "speed_up"
	AdaptationSlowDown   AdaptationKind = "slow_down"
	AdaptationTrim       AdaptationKind = "trim"
	AdaptationLoop       AdaptationKind = "loop"
	AdaptationFreeze     AdaptationKind = "freeze"
	AdaptationInfeasible AdaptationKind = "infeasible"
)

type EffectiveSlotPolicy struct {
	Fit             ffcut.FitMode
	AudioGain       float64
	Overflow        OverflowPolicy
	Underflow       UnderflowPolicy
	Trim            TrimMode
	MinPlaybackRate float64
	MaxPlaybackRate float64
}

type AdaptationPlan struct {
	Kind              AdaptationKind
	Feasible          bool
	Reason            string
	AvailableRange    ffcut.TimeRange
	SourceDuration    ffcut.Duration
	TimelineDuration  ffcut.Duration
	Rate              float64
	Loop              bool
	FreezeLastFrame   ffcut.Duration
	TrimMode          TrimMode
	FixedTrimOffset   ffcut.Duration
	MaximumTrimOffset ffcut.Duration
}

func effectiveSlotPolicy(defaults SlotDefaults, overrides SlotOverrides) EffectiveSlotPolicy {
	policy := EffectiveSlotPolicy(defaults)
	if overrides.Fit != nil {
		policy.Fit = *overrides.Fit
	}
	if overrides.AudioGain != nil {
		policy.AudioGain = *overrides.AudioGain
	}
	if overrides.Overflow != nil {
		policy.Overflow = *overrides.Overflow
	}
	if overrides.Underflow != nil {
		policy.Underflow = *overrides.Underflow
	}
	if overrides.Trim != nil {
		policy.Trim = *overrides.Trim
	}
	if overrides.MinPlaybackRate != nil {
		policy.MinPlaybackRate = *overrides.MinPlaybackRate
	}
	if overrides.MaxPlaybackRate != nil {
		policy.MaxPlaybackRate = *overrides.MaxPlaybackRate
	}
	return policy
}

func planAdaptation(available ffcut.TimeRange, target *ffcut.Duration, policy EffectiveSlotPolicy) AdaptationPlan {
	plan := AdaptationPlan{
		Kind:             AdaptationNatural,
		Feasible:         true,
		AvailableRange:   available,
		SourceDuration:   available.Duration,
		TimelineDuration: available.Duration,
		Rate:             1,
	}
	if target == nil || available.Duration == *target {
		return plan
	}

	if available.Duration > *target {
		switch policy.Overflow {
		case OverflowSpeedUp:
			rate := float64(available.Duration) / float64(*target)
			if !finite(rate) || rate > policy.MaxPlaybackRate {
				return infeasiblePlan(available, *target, fmt.Sprintf("speed_up rate %.6g exceeds maximum %.6g", rate, policy.MaxPlaybackRate))
			}
			plan.Kind = AdaptationSpeedUp
			plan.TimelineDuration = *target
			plan.Rate = rate
			return plan
		case OverflowTrim:
			maximumOffset := available.Duration - *target
			plan.Kind = AdaptationTrim
			plan.SourceDuration = *target
			plan.TimelineDuration = *target
			plan.TrimMode = policy.Trim
			plan.MaximumTrimOffset = maximumOffset
			switch policy.Trim {
			case TrimStart, TrimRandom:
				plan.FixedTrimOffset = 0
			case TrimCenter:
				plan.FixedTrimOffset = ffcut.Duration((maximumOffset.Std() / 2).Truncate(time.Microsecond))
			case TrimEnd:
				plan.FixedTrimOffset = maximumOffset
			}
			return plan
		case OverflowReject:
			return infeasiblePlan(available, *target, "source is longer than fixed slot and overflow policy is reject")
		}
	}

	switch policy.Underflow {
	case UnderflowSlowDown:
		rate := float64(available.Duration) / float64(*target)
		if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < policy.MinPlaybackRate {
			return infeasiblePlan(available, *target, fmt.Sprintf("slow_down rate %.6g is below minimum %.6g", rate, policy.MinPlaybackRate))
		}
		plan.Kind = AdaptationSlowDown
		plan.TimelineDuration = *target
		plan.Rate = rate
		return plan
	case UnderflowLoop:
		plan.Kind = AdaptationLoop
		plan.TimelineDuration = *target
		plan.Loop = true
		return plan
	case UnderflowFreeze:
		plan.Kind = AdaptationFreeze
		plan.TimelineDuration = *target
		plan.FreezeLastFrame = *target - available.Duration
		return plan
	case UnderflowReject:
		return infeasiblePlan(available, *target, "source is shorter than fixed slot and underflow policy is reject")
	}
	return infeasiblePlan(available, *target, "unsupported duration policy")
}

func infeasiblePlan(available ffcut.TimeRange, target ffcut.Duration, reason string) AdaptationPlan {
	return AdaptationPlan{
		Kind:             AdaptationInfeasible,
		Feasible:         false,
		Reason:           reason,
		AvailableRange:   available,
		SourceDuration:   available.Duration,
		TimelineDuration: target,
		Rate:             1,
	}
}
