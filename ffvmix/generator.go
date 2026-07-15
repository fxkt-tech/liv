package ffvmix

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fxkt-tech/liv/ffcut"
	constraintpkg "github.com/fxkt-tech/liv/ffvmix/constraints"
)

const defaultSearchBudget uint64 = 1000

var (
	ErrInvalidGenerator = errors.New("ffvmix: invalid generator")
	ErrConcurrentNext   = errors.New("ffvmix: concurrent Next call")
	ErrConstraintCheck  = errors.New("ffvmix: constraint check failed")
	ErrProjectBuild     = errors.New("ffvmix: project construction failed")
)

// GenerationStatus describes why a Next call returned without ambiguity.
type GenerationStatus string

const (
	Yielded        GenerationStatus = "yielded"
	Exhausted      GenerationStatus = "exhausted"
	BudgetExceeded GenerationStatus = "budget_exceeded"
)

const (
	ReasonInfeasibleVideo        = "engine_infeasible_video"
	ReasonIncompatibleTransition = "engine_incompatible_transition"
)

// GenerationStats is a snapshot of generator progress.
type GenerationStats struct {
	Attempts   uint64
	Yielded    uint64
	Rejected   uint64
	Rejections map[string]uint64
}

// GenerationResult is one Next outcome. Project is non-nil only for Yielded.
type GenerationResult struct {
	Status  GenerationStatus
	Project *ffcut.Project
	Stats   GenerationStats
}

type generatorOptions struct {
	seedProvided bool
	seed         uint64
	searchBudget uint64
	constraints  []Constraint
}

// GeneratorOption configures a generator before any enumeration occurs.
type GeneratorOption func(*generatorOptions) error

// WithSeed makes generation order and random trim offsets reproducible.
func WithSeed(seed uint64) GeneratorOption {
	return func(options *generatorOptions) error {
		options.seedProvided = true
		options.seed = seed
		return nil
	}
}

// WithSearchBudget limits tuple scans performed by each Next call.
func WithSearchBudget(maximum uint64) GeneratorOption {
	return func(options *generatorOptions) error {
		if maximum == 0 {
			return fmt.Errorf("%w: search budget must be positive", ErrInvalidGenerator)
		}
		options.searchBudget = maximum
		return nil
	}
}

// WithConstraint appends a custom pure constraint after template-configured
// built-ins.
func WithConstraint(constraint Constraint) GeneratorOption {
	return func(options *generatorOptions) error {
		if constraint == nil {
			return fmt.Errorf("%w: custom constraint must not be nil", ErrInvalidGenerator)
		}
		options.constraints = append(options.constraints, constraint)
		return nil
	}
}

// WithConstraintFunc appends a custom plugin function with stable provenance.
func WithConstraintFunc(id, fingerprint string, check ConstraintFunc) GeneratorOption {
	return func(options *generatorOptions) error {
		constraint, err := NewConstraint(id, fingerprint, check)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidGenerator, err)
		}
		options.constraints = append(options.constraints, constraint)
		return nil
	}
}

// Generator lazily enumerates and greedily filters one immutable compiled
// template. Next calls must be serial.
type Generator struct {
	seed                uint64
	searchBudget        uint64
	templateID          ID
	templateFingerprint string
	canvas              CanvasSpec
	background          CompiledBackground
	slots               []CompiledSlot
	joins               []CompiledJoin
	bgms                []CompiledBGM
	layers              []CompiledLayer
	bgmDimensionID      ID
	dimensions          []orderedDimension
	enumerator          *diagonalEnumerator
	constraints         []Constraint
	history             []constraintpkg.AcceptedSummary
	running             atomic.Bool
	statsMu             sync.RWMutex
	stats               GenerationStats
}

// NewGenerator creates a pure in-memory iterator over a CompiledTemplate.
func NewGenerator(compiled *CompiledTemplate, optionFunctions ...GeneratorOption) (*Generator, error) {
	if compiled == nil {
		return nil, fmt.Errorf("%w: compiled template is required", ErrInvalidGenerator)
	}
	options := generatorOptions{searchBudget: defaultSearchBudget}
	for _, option := range optionFunctions {
		if option == nil {
			continue
		}
		if err := option(&options); err != nil {
			return nil, err
		}
	}
	if options.searchBudget == 0 {
		return nil, fmt.Errorf("%w: search budget must be positive", ErrInvalidGenerator)
	}
	if !options.seedProvided {
		seed, err := randomSeed()
		if err != nil {
			return nil, fmt.Errorf("%w: generate seed: %v", ErrInvalidGenerator, err)
		}
		options.seed = seed
	}

	generator := &Generator{
		seed:                options.seed,
		searchBudget:        options.searchBudget,
		templateID:          compiled.ID(),
		templateFingerprint: compiled.Fingerprint(),
		canvas:              compiled.Canvas(),
		background:          compiled.Background(),
		slots:               compiled.Slots(),
		joins:               compiled.Joins(),
		bgms:                compiled.BGMs(),
		layers:              compiled.Layers(),
		stats:               GenerationStats{Rejections: make(map[string]uint64)},
	}
	if generator.templateID == "" || strings.TrimSpace(generator.templateFingerprint) == "" {
		return nil, fmt.Errorf("%w: compiled template identity is incomplete", ErrInvalidGenerator)
	}
	if len(generator.slots) == 0 {
		return nil, fmt.Errorf("%w: compiled template has no slots", ErrInvalidGenerator)
	}
	if len(generator.joins) != len(generator.slots)-1 {
		return nil, fmt.Errorf("%w: compiled join count does not match slots", ErrInvalidGenerator)
	}

	dimensions, err := generator.buildDimensions()
	if err != nil {
		return nil, err
	}
	generator.dimensions = dimensions
	generator.enumerator = newDiagonalEnumerator(generator.seed, dimensions)

	builtins, err := builtInConstraints(compiled.Constraints())
	if err != nil {
		return nil, err
	}
	generator.constraints = append(builtins, options.constraints...)
	if err := validateRuntimeConstraints(generator.constraints); err != nil {
		return nil, err
	}
	return generator, nil
}

// Seed returns the actual seed, including an automatically generated one.
func (g *Generator) Seed() uint64 {
	if g == nil {
		return 0
	}
	return g.seed
}

// Stats returns a copy of current progress counters.
func (g *Generator) Stats() GenerationStats {
	if g == nil {
		return GenerationStats{}
	}
	g.statsMu.RLock()
	defer g.statsMu.RUnlock()
	return cloneGenerationStats(g.stats)
}

// Next scans up to the configured budget and returns the next accepted Project.
func (g *Generator) Next(ctx context.Context) (GenerationResult, error) {
	if g == nil {
		return GenerationResult{}, fmt.Errorf("%w: generator is nil", ErrInvalidGenerator)
	}
	if !g.running.CompareAndSwap(false, true) {
		return GenerationResult{}, ErrConcurrentNext
	}
	defer g.running.Store(false)
	if ctx == nil {
		return GenerationResult{}, fmt.Errorf("%w: context is nil", ErrInvalidGenerator)
	}

	for scanned := uint64(0); scanned < g.searchBudget; scanned++ {
		if err := ctx.Err(); err != nil {
			return GenerationResult{}, err
		}
		tuple, ok := g.enumerator.next()
		if !ok {
			return g.result(Exhausted, nil), nil
		}
		g.recordAttempt()

		selection, reason, err := g.selectCombination(tuple)
		if err != nil {
			return GenerationResult{}, err
		}
		if reason != "" {
			g.recordRejection(reason)
			continue
		}

		candidate, err := g.candidateView(selection)
		if err != nil {
			return GenerationResult{}, err
		}
		history := constraintpkg.NewHistoryView(g.history)
		rejected := false
		for _, constraint := range g.constraints {
			decision, checkErr := constraint.Check(candidate, history)
			if checkErr != nil {
				return GenerationResult{}, fmt.Errorf("%w: %s: %w", ErrConstraintCheck, constraint.ID(), checkErr)
			}
			if decision.Accepted {
				continue
			}
			if strings.TrimSpace(decision.Reason) == "" {
				return GenerationResult{}, fmt.Errorf("%w: %s returned a rejection without a reason", ErrConstraintCheck, constraint.ID())
			}
			g.recordRejection(decision.Reason)
			rejected = true
			break
		}
		if rejected {
			continue
		}

		project, err := g.buildProject(selection, candidate)
		if err != nil {
			return GenerationResult{}, fmt.Errorf("%w: %w", ErrProjectBuild, err)
		}
		if err := project.Validate(); err != nil {
			return GenerationResult{}, fmt.Errorf("%w: validate generated project: %w", ErrProjectBuild, err)
		}
		g.history = append(g.history, candidate.Summary())
		g.recordYield()
		return g.result(Yielded, project), nil
	}
	if g.enumerator.exhausted() {
		return g.result(Exhausted, nil), nil
	}
	return g.result(BudgetExceeded, nil), nil
}

type selectedCombination struct {
	videos      []CompiledVideo
	transitions []CompiledTransition
	bgm         *CompiledBGM
	fingerprint string
}

func (g *Generator) buildDimensions() ([]orderedDimension, error) {
	dimensions := make([]orderedDimension, 0, len(g.slots)+len(g.joins)+1)
	for _, slot := range g.slots {
		options := make([]weightedOption, len(slot.Videos))
		for index, video := range slot.Videos {
			if !finitePositive(video.Weight) {
				return nil, fmt.Errorf("%w: slot %s video %s has invalid weight", ErrInvalidGenerator, slot.ID, video.ID)
			}
			options[index] = weightedOption{id: video.ID, weight: video.Weight, index: index}
		}
		if len(options) == 0 {
			return nil, fmt.Errorf("%w: slot %s has no video options", ErrInvalidGenerator, slot.ID)
		}
		dimensions = append(dimensions, orderDimension(g.seed, slot.ID, options))
	}
	for _, join := range g.joins {
		options := make([]weightedOption, len(join.Transitions))
		for index, transition := range join.Transitions {
			if !finitePositive(transition.Weight) {
				return nil, fmt.Errorf("%w: join %s transition %s has invalid weight", ErrInvalidGenerator, join.ID, transition.ID)
			}
			options[index] = weightedOption{id: transition.ID, weight: transition.Weight, index: index}
		}
		if len(options) == 0 {
			return nil, fmt.Errorf("%w: join %s has no transition options", ErrInvalidGenerator, join.ID)
		}
		dimensions = append(dimensions, orderDimension(g.seed, join.ID, options))
	}
	if len(g.bgms) > 0 {
		g.bgmDimensionID = uniqueBGMDimensionID(g.templateID, g.slots, g.joins)
		options := make([]weightedOption, len(g.bgms))
		for index, bgm := range g.bgms {
			if !finitePositive(bgm.Weight) {
				return nil, fmt.Errorf("%w: BGM %s has invalid weight", ErrInvalidGenerator, bgm.ID)
			}
			options[index] = weightedOption{id: bgm.ID, weight: bgm.Weight, index: index}
		}
		dimensions = append(dimensions, orderDimension(g.seed, g.bgmDimensionID, options))
	}
	return dimensions, nil
}

func (g *Generator) selectCombination(tuple []int) (selectedCombination, string, error) {
	selection := selectedCombination{
		videos:      make([]CompiledVideo, len(g.slots)),
		transitions: make([]CompiledTransition, len(g.joins)),
	}
	if len(tuple) != len(g.dimensions) {
		return selection, "", fmt.Errorf("%w: enumerator returned %d dimensions, want %d", ErrInvalidGenerator, len(tuple), len(g.dimensions))
	}
	for index, slot := range g.slots {
		option := g.dimensions[index].options[tuple[index]]
		video := slot.Videos[option.index]
		selection.videos[index] = video
		if !video.Plan.Feasible {
			return selection, ReasonInfeasibleVideo, nil
		}
	}
	joinOffset := len(g.slots)
	for index, join := range g.joins {
		option := g.dimensions[joinOffset+index].options[tuple[joinOffset+index]]
		transition := join.Transitions[option.index]
		selection.transitions[index] = transition
		if !join.IsCompatible(transition.ID, selection.videos[index].ID, selection.videos[index+1].ID) {
			return selection, ReasonIncompatibleTransition, nil
		}
	}
	if len(g.bgms) > 0 {
		option := g.dimensions[len(g.dimensions)-1].options[tuple[len(tuple)-1]]
		bgm := g.bgms[option.index]
		selection.bgm = &bgm
	}
	selection.fingerprint = combinationFingerprint(g, selection)
	return selection, "", nil
}

func (g *Generator) candidateView(selection selectedCombination) (CandidateView, error) {
	videos := make([]constraintpkg.VideoSelection, len(selection.videos))
	for index, video := range selection.videos {
		sourceRange, err := g.sourceRange(g.slots[index], video)
		if err != nil {
			return CandidateView{}, err
		}
		videos[index] = constraintpkg.VideoSelection{
			SlotID:           g.slots[index].ID,
			VideoID:          video.ID,
			Path:             video.Asset.Path,
			AssetFingerprint: video.Asset.FingerprintString(),
			SourceRange:      sourceRange,
			TimelineDuration: video.Plan.TimelineDuration,
		}
	}
	transitions := make([]constraintpkg.TransitionSelection, len(selection.transitions))
	for index, transition := range selection.transitions {
		transitions[index] = constraintpkg.TransitionSelection{
			JoinID:       g.joins[index].ID,
			TransitionID: transition.ID,
			Kind:         transition.Kind,
			Duration:     transition.Duration,
		}
	}
	var bgm *constraintpkg.BGMSelection
	if selection.bgm != nil {
		bgm = &constraintpkg.BGMSelection{
			DimensionID:      g.bgmDimensionID,
			BGMID:            selection.bgm.ID,
			Path:             selection.bgm.Asset.Path,
			AssetFingerprint: selection.bgm.Asset.FingerprintString(),
		}
	}
	return constraintpkg.NewCandidateView(selection.fingerprint, videos, transitions, bgm), nil
}

func (g *Generator) sourceRange(slot CompiledSlot, video CompiledVideo) (ffcut.TimeRange, error) {
	rangeValue := ffcut.TimeRange{
		Start:    video.Plan.AvailableRange.Start,
		Duration: video.Plan.SourceDuration,
	}
	if video.Plan.Kind != AdaptationTrim || video.Plan.TrimMode != TrimRandom {
		start, err := addProtocolDuration(video.Plan.AvailableRange.Start, video.Plan.FixedTrimOffset)
		if err != nil {
			return ffcut.TimeRange{}, err
		}
		rangeValue.Start = start
		return rangeValue, nil
	}
	maximumMicros, err := video.Plan.MaximumTrimOffset.Microseconds()
	if err != nil {
		return ffcut.TimeRange{}, fmt.Errorf("invalid random trim maximum for slot %s video %s: %w", slot.ID, video.ID, err)
	}
	if maximumMicros < 0 {
		return ffcut.TimeRange{}, fmt.Errorf("invalid negative random trim maximum for slot %s video %s", slot.ID, video.ID)
	}
	random := deterministicUint64(g.seed, "trim", g.templateFingerprint, string(slot.ID), string(video.ID))
	offsetMicros := random % (uint64(maximumMicros) + 1)
	offset, err := ffcut.DurationFromMicroseconds(int64(offsetMicros))
	if err != nil {
		return ffcut.TimeRange{}, err
	}
	start, err := addProtocolDuration(video.Plan.AvailableRange.Start, offset)
	if err != nil {
		return ffcut.TimeRange{}, err
	}
	rangeValue.Start = start
	return rangeValue, nil
}

func builtInConstraints(specs []ConstraintSpec) ([]Constraint, error) {
	constraints := make([]Constraint, 0, len(specs))
	for _, spec := range specs {
		var constraint Constraint
		var err error
		switch spec.Kind {
		case ConstraintMaxSimilarity:
			if spec.MaxSimilarity == nil {
				return nil, fmt.Errorf("%w: constraint %s has no max-similarity payload", ErrInvalidGenerator, spec.ID)
			}
			constraint, err = constraintpkg.NewMaxSimilarity(string(spec.ID), spec.MaxSimilarity.Maximum)
		case ConstraintMaxVideoAssetUses:
			if spec.MaxVideoAssetUses == nil {
				return nil, fmt.Errorf("%w: constraint %s has no video-use payload", ErrInvalidGenerator, spec.ID)
			}
			constraint, err = constraintpkg.NewMaxVideoAssetUses(string(spec.ID), spec.MaxVideoAssetUses.Maximum)
		case ConstraintMaxBGMUses:
			if spec.MaxBGMUses == nil {
				return nil, fmt.Errorf("%w: constraint %s has no BGM-use payload", ErrInvalidGenerator, spec.ID)
			}
			constraint, err = constraintpkg.NewMaxBGMUses(string(spec.ID), spec.MaxBGMUses.Maximum)
		default:
			return nil, fmt.Errorf("%w: unsupported constraint kind %q", ErrInvalidGenerator, spec.Kind)
		}
		if err != nil {
			return nil, fmt.Errorf("%w: constraint %s: %v", ErrInvalidGenerator, spec.ID, err)
		}
		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

func validateRuntimeConstraints(values []Constraint) error {
	ids := make(map[string]struct{}, len(values))
	for index, constraint := range values {
		if constraint == nil || isNilConstraint(constraint) {
			return fmt.Errorf("%w: constraint %d is nil", ErrInvalidGenerator, index)
		}
		id := strings.TrimSpace(constraint.ID())
		if id == "" {
			return fmt.Errorf("%w: constraint %d has an empty ID", ErrInvalidGenerator, index)
		}
		if _, exists := ids[id]; exists {
			return fmt.Errorf("%w: duplicate constraint ID %q", ErrInvalidGenerator, id)
		}
		ids[id] = struct{}{}
		if strings.TrimSpace(constraint.Fingerprint()) == "" {
			return fmt.Errorf("%w: constraint %q has an empty fingerprint", ErrInvalidGenerator, id)
		}
	}
	return nil
}

func isNilConstraint(constraint Constraint) bool {
	value := reflect.ValueOf(constraint)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func uniqueBGMDimensionID(templateID ID, slots []CompiledSlot, joins []CompiledJoin) ID {
	used := make(map[ID]struct{}, len(slots)+len(joins))
	for _, slot := range slots {
		used[slot.ID] = struct{}{}
	}
	for _, join := range joins {
		used[join.ID] = struct{}{}
	}
	candidate := ID(string(templateID) + ":bgm")
	for {
		if _, exists := used[candidate]; !exists {
			return candidate
		}
		candidate += ":pool"
	}
}

func combinationFingerprint(g *Generator, selection selectedCombination) string {
	parts := make([]string, 0, 2+len(selection.videos)*2+len(selection.transitions)*2+2)
	parts = append(parts, g.templateFingerprint)
	for index, video := range selection.videos {
		parts = append(parts, string(g.slots[index].ID), string(video.ID))
	}
	for index, transition := range selection.transitions {
		parts = append(parts, string(g.joins[index].ID), string(transition.ID))
	}
	if selection.bgm != nil {
		parts = append(parts, string(g.bgmDimensionID), string(selection.bgm.ID))
	}
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte(strconv.Itoa(len(part))))
		_, _ = hash.Write([]byte{':'})
		_, _ = hash.Write([]byte(part))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func randomSeed() (uint64, error) {
	var buffer [8]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buffer[:]), nil
}

func finitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func addProtocolDuration(left, right ffcut.Duration) (ffcut.Duration, error) {
	return (ffcut.TimeRange{Start: left, Duration: right}).End()
}

func (g *Generator) recordAttempt() {
	g.statsMu.Lock()
	g.stats.Attempts++
	g.statsMu.Unlock()
}

func (g *Generator) recordRejection(reason string) {
	g.statsMu.Lock()
	g.stats.Rejected++
	g.stats.Rejections[reason]++
	g.statsMu.Unlock()
}

func (g *Generator) recordYield() {
	g.statsMu.Lock()
	g.stats.Yielded++
	g.statsMu.Unlock()
}

func (g *Generator) result(status GenerationStatus, project *ffcut.Project) GenerationResult {
	return GenerationResult{Status: status, Project: project, Stats: g.Stats()}
}

func cloneGenerationStats(value GenerationStats) GenerationStats {
	cloned := value
	cloned.Rejections = make(map[string]uint64, len(value.Rejections))
	for reason, count := range value.Rejections {
		cloned.Rejections[reason] = count
	}
	return cloned
}
