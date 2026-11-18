package fusion

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

// Timeline is a renderer-facing, local-media composition. It deliberately
// contains no upload, job, or product semantics.
type Timeline struct {
	Width      int32            `json:"width"`
	Height     int32            `json:"height"`
	FPS        float64          `json:"fps"`
	Background string           `json:"background"`
	DurationMs int64            `json:"duration_ms"`
	LUTFile    string           `json:"lut_file,omitempty"`
	OmitAudio  bool             `json:"omit_audio,omitempty"`
	Tracks     []*TimelineTrack `json:"tracks"`
}

type TimelineTrackKind string

const (
	TimelineTrackVisual TimelineTrackKind = "visual"
	TimelineTrackAudio  TimelineTrackKind = "audio"
)

// TimelineTrack.Order is ascending: a higher visual order is composited on
// top of a lower order. IDs provide a deterministic tie-breaker.
type TimelineTrack struct {
	ID     string            `json:"id"`
	Kind   TimelineTrackKind `json:"kind"`
	Order  int               `json:"order"`
	Hidden bool              `json:"hidden,omitempty"`
	Muted  bool              `json:"muted,omitempty"`
	Clips  []*TimelineClip   `json:"clips"`
}

type TimelineClipKind string

const (
	TimelineClipVideo     TimelineClipKind = "video"
	TimelineClipImage     TimelineClipKind = "image"
	TimelineClipAnimation TimelineClipKind = "animation"
	TimelineClipAudio     TimelineClipKind = "audio"
)

type TimelineRange struct {
	StartMs    int64 `json:"start_ms"`
	DurationMs int64 `json:"duration_ms"`
}

type TimelineCrop struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type TimelineFit string

const (
	TimelineFitCover   TimelineFit = "cover"
	TimelineFitContain TimelineFit = "contain"
	TimelineFitStretch TimelineFit = "stretch"
)

type TimelineTransform struct {
	X               int32         `json:"x"`
	Y               int32         `json:"y"`
	Width           int32         `json:"width"`
	Height          int32         `json:"height"`
	RotationDegrees float64       `json:"rotation_degrees"`
	Opacity         float64       `json:"opacity"`
	Fit             TimelineFit   `json:"fit"`
	Crop            *TimelineCrop `json:"crop,omitempty"`
}

type TimelineAudioConfig struct {
	Enabled bool    `json:"enabled"`
	Volume  float64 `json:"volume"`
}

// TimelineClip.Order is an explicit stacking order inside a visual track.
// SourceRange is required for video/audio and ignored for still/animation
// images. HasAudio avoids assuming every video container has an audio stream.
type TimelineClip struct {
	ID          string              `json:"id"`
	Kind        TimelineClipKind    `json:"kind"`
	Order       int                 `json:"order"`
	Path        string              `json:"path"`
	Range       TimelineRange       `json:"range"`
	SourceRange TimelineRange       `json:"source_range"`
	Transform   TimelineTransform   `json:"transform"`
	Audio       TimelineAudioConfig `json:"audio"`
	HasAudio    bool                `json:"has_audio,omitempty"`
	Loop        bool                `json:"loop,omitempty"`
}

type TimelineValidationIssue struct {
	Path    string
	Message string
}

func (i TimelineValidationIssue) Error() string {
	return i.Path + ": " + i.Message
}

type TimelineValidationError struct {
	Issues []TimelineValidationIssue
}

func (e *TimelineValidationError) Error() string {
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, issue.Error())
	}
	return "invalid fusion timeline: " + strings.Join(parts, "; ")
}

func (e *TimelineValidationError) Is(target error) bool {
	return target == ErrInvalidTimeline
}

var ErrInvalidTimeline = errors.New("invalid fusion timeline")

func (t *Timeline) Validate() error {
	issues := make([]TimelineValidationIssue, 0)
	add := func(path, message string) {
		issues = append(issues, TimelineValidationIssue{Path: path, Message: message})
	}
	if t == nil {
		add("timeline", "is required")
		return &TimelineValidationError{Issues: issues}
	}
	if t.Width <= 0 {
		add("width", "must be positive")
	}
	if t.Height <= 0 {
		add("height", "must be positive")
	}
	if !finiteTimelineNumber(t.FPS) || t.FPS <= 0 {
		add("fps", "must be finite and positive")
	}
	if strings.TrimSpace(t.Background) == "" {
		add("background", "must not be empty")
	}
	if t.DurationMs <= 0 {
		add("duration_ms", "must be positive")
	}
	if t.LUTFile != "" && !filepath.IsAbs(t.LUTFile) {
		add("lut_file", "must be an absolute path")
	}

	trackIDs := make(map[string]string)
	clipIDs := make(map[string]string)
	for trackIndex, track := range t.Tracks {
		trackPath := fmt.Sprintf("tracks[%d]", trackIndex)
		if track == nil {
			add(trackPath, "must not be nil")
			continue
		}
		validateTimelineID(&issues, trackIDs, trackPath+".id", track.ID)
		switch track.Kind {
		case TimelineTrackVisual, TimelineTrackAudio:
		default:
			add(trackPath+".kind", fmt.Sprintf("unsupported value %q", track.Kind))
		}
		for clipIndex, clip := range track.Clips {
			clipPath := fmt.Sprintf("%s.clips[%d]", trackPath, clipIndex)
			if clip == nil {
				add(clipPath, "must not be nil")
				continue
			}
			validateTimelineID(&issues, clipIDs, clipPath+".id", clip.ID)
			if !filepath.IsAbs(clip.Path) {
				add(clipPath+".path", "must be an absolute path")
			}
			validateTimelineRange(&issues, clipPath+".range", clip.Range, false)
			if clip.Range.StartMs >= 0 && clip.Range.DurationMs > 0 && t.DurationMs > 0 && clip.Range.StartMs > t.DurationMs-clip.Range.DurationMs {
				add(clipPath+".range", "must not extend beyond timeline duration")
			}
			switch clip.Kind {
			case TimelineClipVideo:
				if track.Kind != TimelineTrackVisual {
					add(clipPath+".kind", "video requires a visual track")
				}
				validateTimelineRange(&issues, clipPath+".source_range", clip.SourceRange, false)
				validateTimelineTransform(&issues, clipPath+".transform", clip.Transform)
				if clip.Loop {
					add(clipPath+".loop", "is not supported for video clips")
				}
			case TimelineClipImage:
				if track.Kind != TimelineTrackVisual {
					add(clipPath+".kind", "image requires a visual track")
				}
				if clip.Loop {
					add(clipPath+".loop", "must be false for a still image")
				}
				validateTimelineTransform(&issues, clipPath+".transform", clip.Transform)
			case TimelineClipAnimation:
				if track.Kind != TimelineTrackVisual {
					add(clipPath+".kind", "animation requires a visual track")
				}
				validateTimelineTransform(&issues, clipPath+".transform", clip.Transform)
				if !clip.Loop {
					add(clipPath+".loop", "must be true for animation clips")
				}
			case TimelineClipAudio:
				if track.Kind != TimelineTrackAudio {
					add(clipPath+".kind", "audio requires an audio track")
				}
				validateTimelineRange(&issues, clipPath+".source_range", clip.SourceRange, false)
				if clip.Loop {
					add(clipPath+".loop", "is not supported for audio clips")
				}
			default:
				add(clipPath+".kind", fmt.Sprintf("unsupported value %q", clip.Kind))
			}
			if clip.Kind == TimelineClipAudio && clip.HasAudio {
				add(clipPath+".has_audio", "is implicit for audio clips and must be false")
			}
			if clip.Kind != TimelineClipVideo && clip.Kind != TimelineClipAudio && clip.Audio.Enabled {
				add(clipPath+".audio.enabled", "is only valid for video or audio clips")
			}
			if clip.Kind == TimelineClipVideo && clip.Audio.Enabled && !clip.HasAudio {
				add(clipPath+".audio.enabled", "requires has_audio")
			}
			if !finiteTimelineNumber(clip.Audio.Volume) || clip.Audio.Volume < 0 {
				add(clipPath+".audio.volume", "must be finite and non-negative")
			}
			if !clip.Audio.Enabled && clip.Audio.Volume != 0 {
				add(clipPath+".audio.volume", "must be zero when audio is disabled")
			}
		}
	}
	if len(issues) > 0 {
		return &TimelineValidationError{Issues: issues}
	}
	return nil
}

func validateTimelineID(issues *[]TimelineValidationIssue, seen map[string]string, path, id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		*issues = append(*issues, TimelineValidationIssue{Path: path, Message: "must not be empty"})
		return
	}
	if previous, ok := seen[id]; ok {
		*issues = append(*issues, TimelineValidationIssue{Path: path, Message: "duplicates " + previous})
		return
	}
	seen[id] = path
}

func validateTimelineRange(issues *[]TimelineValidationIssue, path string, value TimelineRange, allowEmpty bool) {
	if value.StartMs < 0 {
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".start_ms", Message: "must be non-negative"})
	}
	if value.DurationMs < 0 || (!allowEmpty && value.DurationMs == 0) {
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".duration_ms", Message: "must be positive"})
	}
	if value.StartMs > math.MaxInt64-value.DurationMs {
		*issues = append(*issues, TimelineValidationIssue{Path: path, Message: "end overflows int64"})
	}
}

func validateTimelineTransform(issues *[]TimelineValidationIssue, path string, value TimelineTransform) {
	if value.Width <= 0 {
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".width", Message: "must be positive"})
	}
	if value.Height <= 0 {
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".height", Message: "must be positive"})
	}
	if !finiteTimelineNumber(value.Opacity) || value.Opacity < 0 || value.Opacity > 1 {
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".opacity", Message: "must be between zero and one"})
	}
	if !finiteTimelineNumber(value.RotationDegrees) {
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".rotation_degrees", Message: "must be finite"})
	}
	switch value.Fit {
	case TimelineFitCover, TimelineFitContain, TimelineFitStretch:
	default:
		*issues = append(*issues, TimelineValidationIssue{Path: path + ".fit", Message: fmt.Sprintf("unsupported value %q", value.Fit)})
	}
	if value.Crop != nil {
		crop := value.Crop
		if !finiteTimelineNumber(crop.X) || !finiteTimelineNumber(crop.Y) || !finiteTimelineNumber(crop.Width) || !finiteTimelineNumber(crop.Height) ||
			crop.X < 0 || crop.Y < 0 || crop.Width <= 0 || crop.Height <= 0 || crop.X+crop.Width > 1 || crop.Y+crop.Height > 1 {
			*issues = append(*issues, TimelineValidationIssue{Path: path + ".crop", Message: "must be a positive normalized rectangle inside the source"})
		}
	}
}

func finiteTimelineNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

type orderedTimelineClip struct {
	track *TimelineTrack
	clip  *TimelineClip
}

func (t *Timeline) orderedClips(kind TimelineTrackKind) []orderedTimelineClip {
	tracks := make([]*TimelineTrack, 0, len(t.Tracks))
	for _, track := range t.Tracks {
		if track != nil && !track.Hidden && track.Kind == kind {
			tracks = append(tracks, track)
		}
	}
	sort.SliceStable(tracks, func(i, j int) bool {
		if tracks[i].Order != tracks[j].Order {
			return tracks[i].Order < tracks[j].Order
		}
		return tracks[i].ID < tracks[j].ID
	})
	ordered := make([]orderedTimelineClip, 0)
	for _, track := range tracks {
		clips := append([]*TimelineClip(nil), track.Clips...)
		sort.SliceStable(clips, func(i, j int) bool {
			if clips[i].Order != clips[j].Order {
				return clips[i].Order < clips[j].Order
			}
			if clips[i].Range.StartMs != clips[j].Range.StartMs {
				return clips[i].Range.StartMs < clips[j].Range.StartMs
			}
			return clips[i].ID < clips[j].ID
		})
		for _, clip := range clips {
			ordered = append(ordered, orderedTimelineClip{track: track, clip: clip})
		}
	}
	return ordered
}
