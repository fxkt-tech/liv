package ffvmix

import (
	"math"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

func TestPlanAdaptationPolicies(t *testing.T) {
	tests := []struct {
		name       string
		source     time.Duration
		target     time.Duration
		configure  func(*EffectiveSlotPolicy)
		wantKind   AdaptationKind
		wantRate   float64
		wantLoop   bool
		wantFreeze time.Duration
		feasible   bool
	}{
		{
			name:   "natural",
			source: 5 * time.Second, target: 5 * time.Second,
			wantKind: AdaptationNatural, wantRate: 1, feasible: true,
		},
		{
			name:   "speed up",
			source: 10 * time.Second, target: 5 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Overflow = OverflowSpeedUp },
			wantKind:  AdaptationSpeedUp, wantRate: 2, feasible: true,
		},
		{
			name:   "trim",
			source: 10 * time.Second, target: 5 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Overflow = OverflowTrim },
			wantKind:  AdaptationTrim, wantRate: 1, feasible: true,
		},
		{
			name:   "overflow reject",
			source: 10 * time.Second, target: 5 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Overflow = OverflowReject },
			wantKind:  AdaptationInfeasible, wantRate: 1, feasible: false,
		},
		{
			name:   "slow down",
			source: 2 * time.Second, target: 4 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Underflow = UnderflowSlowDown },
			wantKind:  AdaptationSlowDown, wantRate: 0.5, feasible: true,
		},
		{
			name:   "loop",
			source: 2 * time.Second, target: 4 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Underflow = UnderflowLoop },
			wantKind:  AdaptationLoop, wantRate: 1, wantLoop: true, feasible: true,
		},
		{
			name:   "freeze",
			source: 2 * time.Second, target: 4 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Underflow = UnderflowFreeze },
			wantKind:  AdaptationFreeze, wantRate: 1, wantFreeze: 2 * time.Second, feasible: true,
		},
		{
			name:   "underflow reject",
			source: 2 * time.Second, target: 4 * time.Second,
			configure: func(policy *EffectiveSlotPolicy) { policy.Underflow = UnderflowReject },
			wantKind:  AdaptationInfeasible, wantRate: 1, feasible: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaults := DefaultSlotDefaults()
			policy := effectiveSlotPolicy(defaults, SlotOverrides{})
			if test.configure != nil {
				test.configure(&policy)
			}
			source := templateDuration(t, test.source)
			target := templateDuration(t, test.target)
			plan := planAdaptation(ffcut.TimeRange{Duration: source}, &target, policy)
			if plan.Kind != test.wantKind || plan.Feasible != test.feasible {
				t.Fatalf("plan = %#v, want kind %q feasible %v", plan, test.wantKind, test.feasible)
			}
			if math.Abs(plan.Rate-test.wantRate) > 1e-9 {
				t.Fatalf("rate = %v, want %v", plan.Rate, test.wantRate)
			}
			if plan.Loop != test.wantLoop {
				t.Fatalf("loop = %v, want %v", plan.Loop, test.wantLoop)
			}
			if plan.FreezeLastFrame.Std() != test.wantFreeze {
				t.Fatalf("freeze = %s, want %s", plan.FreezeLastFrame.Std(), test.wantFreeze)
			}
		})
	}
}

func TestPlanAdaptationDefersRandomTrimOffset(t *testing.T) {
	policy := effectiveSlotPolicy(DefaultSlotDefaults(), SlotOverrides{})
	policy.Overflow = OverflowTrim
	policy.Trim = TrimRandom
	source := templateDuration(t, 10*time.Second)
	target := templateDuration(t, 4*time.Second)
	plan := planAdaptation(ffcut.TimeRange{Start: templateDuration(t, time.Second), Duration: source}, &target, policy)
	if plan.Kind != AdaptationTrim || plan.TrimMode != TrimRandom {
		t.Fatalf("plan = %#v, want random trim", plan)
	}
	if plan.FixedTrimOffset != 0 || plan.MaximumTrimOffset.Std() != 6*time.Second {
		t.Fatalf("trim offsets = %s/%s, want 0/6s", plan.FixedTrimOffset.Std(), plan.MaximumTrimOffset.Std())
	}
}

func TestPlanAdaptationKeepsCenterTrimAtMicrosecondPrecision(t *testing.T) {
	policy := effectiveSlotPolicy(DefaultSlotDefaults(), SlotOverrides{})
	policy.Overflow = OverflowTrim
	policy.Trim = TrimCenter
	source := templateDuration(t, 5*time.Second+time.Microsecond)
	target := templateDuration(t, 5*time.Second)
	plan := planAdaptation(ffcut.TimeRange{Duration: source}, &target, policy)
	if _, err := plan.FixedTrimOffset.Microseconds(); err != nil {
		t.Fatalf("center trim offset = %s: %v", plan.FixedTrimOffset.Std(), err)
	}
}

func TestPlanAdaptationEnforcesRateLimits(t *testing.T) {
	policy := effectiveSlotPolicy(DefaultSlotDefaults(), SlotOverrides{})
	policy.Overflow = OverflowSpeedUp
	policy.MaxPlaybackRate = 1.5
	source := templateDuration(t, 10*time.Second)
	target := templateDuration(t, 5*time.Second)
	plan := planAdaptation(ffcut.TimeRange{Duration: source}, &target, policy)
	if plan.Feasible || plan.Kind != AdaptationInfeasible {
		t.Fatalf("plan = %#v, want infeasible", plan)
	}
}
