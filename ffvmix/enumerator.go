package ffvmix

import (
	"container/heap"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"sort"
	"strconv"
	"strings"
)

type weightedOption struct {
	id     ID
	weight float64
	index  int
}

type orderedDimension struct {
	id      ID
	options []weightedOption
}

func orderDimension(seed uint64, dimensionID ID, options []weightedOption) orderedDimension {
	ordered := append([]weightedOption(nil), options...)
	type ranked struct {
		option weightedOption
		key    float64
	}
	rankedOptions := make([]ranked, len(ordered))
	for index, option := range ordered {
		random := deterministicUnit(seed, "dimension", string(dimensionID), string(option.id))
		rankedOptions[index] = ranked{
			option: option,
			key:    -math.Log(random) / option.weight,
		}
	}
	sort.Slice(rankedOptions, func(left, right int) bool {
		if rankedOptions[left].key != rankedOptions[right].key {
			return rankedOptions[left].key < rankedOptions[right].key
		}
		return rankedOptions[left].option.id < rankedOptions[right].option.id
	})
	for index, rankedOption := range rankedOptions {
		ordered[index] = rankedOption.option
	}
	return orderedDimension{id: dimensionID, options: ordered}
}

type tupleItem struct {
	indices  []int
	diagonal int
	tie      uint64
}

type tupleHeap []tupleItem

func (h tupleHeap) Len() int { return len(h) }
func (h tupleHeap) Less(left, right int) bool {
	if h[left].diagonal != h[right].diagonal {
		return h[left].diagonal < h[right].diagonal
	}
	if h[left].tie != h[right].tie {
		return h[left].tie < h[right].tie
	}
	for index := range h[left].indices {
		if h[left].indices[index] != h[right].indices[index] {
			return h[left].indices[index] < h[right].indices[index]
		}
	}
	return false
}
func (h tupleHeap) Swap(left, right int) { h[left], h[right] = h[right], h[left] }
func (h *tupleHeap) Push(value any)      { *h = append(*h, value.(tupleItem)) }
func (h *tupleHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}

type diagonalEnumerator struct {
	seed       uint64
	dimensions []orderedDimension
	queue      tupleHeap
	visited    map[string]struct{}
}

func newDiagonalEnumerator(seed uint64, dimensions []orderedDimension) *diagonalEnumerator {
	enumerator := &diagonalEnumerator{
		seed:       seed,
		dimensions: append([]orderedDimension(nil), dimensions...),
		visited:    make(map[string]struct{}),
	}
	if len(dimensions) == 0 {
		return enumerator
	}
	indices := make([]int, len(dimensions))
	enumerator.push(indices)
	heap.Init(&enumerator.queue)
	return enumerator
}

func (e *diagonalEnumerator) next() ([]int, bool) {
	if e == nil || len(e.queue) == 0 {
		return nil, false
	}
	current := heap.Pop(&e.queue).(tupleItem)
	for dimension := range current.indices {
		next := append([]int(nil), current.indices...)
		next[dimension]++
		if next[dimension] >= len(e.dimensions[dimension].options) {
			continue
		}
		e.push(next)
	}
	return append([]int(nil), current.indices...), true
}

func (e *diagonalEnumerator) exhausted() bool {
	return e == nil || len(e.queue) == 0
}

func (e *diagonalEnumerator) push(indices []int) {
	key := tupleKey(indices)
	if _, exists := e.visited[key]; exists {
		return
	}
	e.visited[key] = struct{}{}
	diagonal := 0
	for _, index := range indices {
		diagonal += index
	}
	heap.Push(&e.queue, tupleItem{
		indices:  append([]int(nil), indices...),
		diagonal: diagonal,
		tie:      deterministicUint64(e.seed, "tuple", key),
	})
}

func tupleKey(indices []int) string {
	var builder strings.Builder
	for index, value := range indices {
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.Itoa(value))
	}
	return builder.String()
}

func deterministicUnit(seed uint64, parts ...string) float64 {
	value := deterministicUint64(seed, parts...)
	const denominator = float64(uint64(1) << 53)
	return (float64(value>>11) + 0.5) / denominator
}

func deterministicUint64(seed uint64, parts ...string) uint64 {
	hash := sha256.New()
	var encodedSeed [8]byte
	binary.BigEndian.PutUint64(encodedSeed[:], seed)
	_, _ = hash.Write(encodedSeed[:])
	for _, part := range parts {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write([]byte(part))
	}
	return binary.BigEndian.Uint64(hash.Sum(nil)[:8])
}
