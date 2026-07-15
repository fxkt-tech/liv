package ffvmix

import "testing"

func TestDiagonalEnumeratorCoversProductExactlyOnce(t *testing.T) {
	dimensions := []orderedDimension{
		orderDimension(7, "a", weightedOptions(2)),
		orderDimension(7, "b", weightedOptions(3)),
		orderDimension(7, "c", weightedOptions(4)),
	}
	enumerator := newDiagonalEnumerator(7, dimensions)
	seen := make(map[string]struct{})
	for {
		tuple, ok := enumerator.next()
		if !ok {
			break
		}
		key := tupleKey(tuple)
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate tuple %q", key)
		}
		seen[key] = struct{}{}
	}
	if got, want := len(seen), 2*3*4; got != want {
		t.Fatalf("tuple count = %d, want %d", got, want)
	}
}

func TestWeightedOrderBiasesEarlierWithoutRemovingOptions(t *testing.T) {
	heavyFirst := 0
	for seed := uint64(1); seed <= 500; seed++ {
		dimension := orderDimension(seed, "slot", []weightedOption{
			{id: "heavy", weight: 9, index: 0},
			{id: "light", weight: 1, index: 1},
		})
		if len(dimension.options) != 2 {
			t.Fatalf("option count = %d", len(dimension.options))
		}
		if dimension.options[0].id == "heavy" {
			heavyFirst++
		}
	}
	if heavyFirst < 400 {
		t.Fatalf("heavy option appeared first %d/500 times, want a clear weight bias", heavyFirst)
	}
}

func weightedOptions(count int) []weightedOption {
	options := make([]weightedOption, count)
	for index := range options {
		options[index] = weightedOption{id: ID(string(rune('a' + index))), weight: 1, index: index}
	}
	return options
}
