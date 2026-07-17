package ffcut

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func TestProjectValidate(t *testing.T) {
	project := validProject()
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectValidateRejectsNil(t *testing.T) {
	var project *Project
	if err := project.Validate(); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("Validate() error = %v, want ErrInvalidProject", err)
	}
}

func TestProjectValidateRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		path string
		want error
		edit func(*Project)
	}{
		{
			name: "unsupported version",
			path: "version",
			want: ErrUnsupportedVersion,
			edit: func(project *Project) { project.Version = 99 },
		},
		{
			name: "duplicate object ID",
			path: "audio[0].id",
			edit: func(project *Project) { project.Audio[0].ID = project.Video.Clips[0].ID },
		},
		{
			name: "invalid canvas",
			path: "canvas.width",
			edit: func(project *Project) { project.Canvas.Width = 0 },
		},
		{
			name: "background payload mismatch",
			path: "canvas.background.color",
			edit: func(project *Project) { project.Canvas.Background.Color = nil },
		},
		{
			name: "invalid source fingerprint",
			path: "video.clips[0].source.fingerprint.sha256",
			edit: func(project *Project) { project.Video.Clips[0].Source.Fingerprint.SHA256 = "not-a-sha256" },
		},
		{
			name: "source path must be absolute and local",
			path: "video.clips[0].source.path",
			edit: func(project *Project) { project.Video.Clips[0].Source.Path = "https://example.com/a.mp4" },
		},
		{
			name: "negative source start",
			path: "video.clips[0].source_range.start",
			want: ErrInvalidDuration,
			edit: func(project *Project) { project.Video.Clips[0].SourceRange.Start = duration(-time.Second) },
		},
		{
			name: "sub-microsecond timeline",
			path: "video.clips[0].timeline_range.duration",
			want: ErrInvalidDuration,
			edit: func(project *Project) { project.Video.Clips[0].TimelineRange.Duration++ },
		},
		{
			name: "transition reference",
			path: "video.transitions[0].to_clip_id",
			edit: func(project *Project) { project.Video.Transitions[0].ToClipID = "other" },
		},
		{
			name: "transition overlap",
			path: "video.transitions[0].range",
			edit: func(project *Project) { project.Video.Transitions[0].Range.Start = duration(3 * time.Second) },
		},
		{
			name: "loop and freeze are mutually exclusive",
			path: "video.clips[0].playback",
			edit: func(project *Project) {
				project.Video.Clips[0].Playback.Loop = true
				project.Video.Clips[0].Playback.FreezeLastFrame = duration(time.Second)
			},
		},
		{
			name: "playback must explain timeline duration",
			path: "video.clips[0].playback",
			edit: func(project *Project) { project.Video.Clips[0].SourceRange.Duration = duration(4 * time.Second) },
		},
		{
			name: "disabled clip audio has no gain",
			path: "video.clips[1].audio.gain",
			edit: func(project *Project) { project.Video.Clips[1].Audio.Gain = 1 },
		},
		{
			name: "non-looping audio cannot outlast source",
			path: "audio[0].timeline_range.duration",
			edit: func(project *Project) { project.Audio[0].SourceRange.Duration = duration(time.Second) },
		},
		{
			name: "audio cannot outlast video",
			path: "audio[0].timeline_range",
			edit: func(project *Project) {
				project.Audio[0].Loop = true
				project.Audio[0].TimelineRange.Duration = duration(10 * time.Second)
			},
		},
		{
			name: "layer payload mismatch",
			path: "layers[0]",
			edit: func(project *Project) { project.Layers[0].Subtitle = project.Layers[2].Subtitle },
		},
		{
			name: "percentage out of range",
			path: "layers[0].image.geometry.width.value",
			edit: func(project *Project) { project.Layers[0].Image.Geometry.Width.Value = 101 },
		},
		{
			name: "animation layer must loop",
			path: "layers[1].media.loop",
			edit: func(project *Project) { project.Layers[1].Media.Loop = false },
		},
		{
			name: "stroke color required",
			path: "layers[2].subtitle.style.stroke_color",
			edit: func(project *Project) { project.Layers[2].Subtitle.Style.StrokeColor = "" },
		},
		{
			name: "subtitle opacity out of range",
			path: "layers[2].subtitle.opacity",
			edit: func(project *Project) { *project.Layers[2].Subtitle.Opacity = 1.1 },
		},
		{
			name: "subtitle rotation must be finite",
			path: "layers[2].subtitle.rotation_degrees",
			edit: func(project *Project) { project.Layers[2].Subtitle.RotationDegrees = math.Inf(1) },
		},
		{
			name: "cue outside layer",
			path: "layers[2].subtitle.cues[1].range",
			edit: func(project *Project) { project.Layers[2].Subtitle.Cues[1].Range.Start = duration(8 * time.Second) },
		},
		{
			name: "subtitle requires a font",
			path: "layers[2].subtitle.style.font_family",
			edit: func(project *Project) { project.Layers[2].Subtitle.Style.FontFamily = "" },
		},
		{
			name: "unsupported subtitle alignment",
			path: "layers[2].subtitle.style.align",
			edit: func(project *Project) { project.Layers[2].Subtitle.Style.Align = "justify" },
		},
		{
			name: "non finite gain",
			path: "audio[0].gain",
			edit: func(project *Project) { project.Audio[0].Gain = math.Inf(1) },
		},
		{
			name: "duplicate selection dimension",
			path: "metadata.selections[1].dimension_id",
			edit: func(project *Project) {
				project.Metadata.Selections[1].DimensionID = project.Metadata.Selections[0].DimensionID
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := cloneProject(t, validProject())
			test.edit(project)
			err := project.Validate()
			if !errors.Is(err, ErrInvalidProject) {
				t.Fatalf("Validate() error = %v, want ErrInvalidProject", err)
			}
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("Validate() error = %v, want %v", err, test.want)
			}
			if !strings.Contains(err.Error(), test.path) {
				t.Fatalf("Validate() error = %v, want path %q", err, test.path)
			}
		})
	}
}

func TestBackgroundVariantsValidate(t *testing.T) {
	tests := []Background{
		{
			Kind: BackgroundKindImage,
			Image: &ImageBackground{
				Source: testSource("background-source", "/media/background.png", 5000, 5),
				Fit:    FitModeCover,
			},
		},
		{
			Kind: BackgroundKindBlur,
			Blur: &BlurBackground{Sigma: 12},
		},
	}

	for _, background := range tests {
		t.Run(string(background.Kind), func(t *testing.T) {
			project := validProject()
			project.Canvas.Background = background
			if err := project.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestPlaybackVariantsValidate(t *testing.T) {
	tests := []struct {
		name     string
		source   time.Duration
		playback Playback
	}{
		{
			name:     "speed up",
			source:   10 * time.Second,
			playback: Playback{Rate: 2},
		},
		{
			name:     "loop",
			source:   2 * time.Second,
			playback: Playback{Rate: 1, Loop: true},
		},
		{
			name:     "freeze",
			source:   4 * time.Second,
			playback: Playback{Rate: 1, FreezeLastFrame: duration(time.Second)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := validProject()
			project.Video.Clips[0].SourceRange.Duration = duration(test.source)
			project.Video.Clips[0].Playback = test.playback
			if err := project.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestProjectValidateAggregatesIssues(t *testing.T) {
	project := validProject()
	project.Canvas.Width = 0
	project.Canvas.Height = 0
	project.Metadata.TemplateFingerprint = ""

	err := project.Validate()
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Validate() error = %T, want *ValidationError", err)
	}
	if len(validationErr.Issues) < 3 {
		t.Fatalf("len(Issues) = %d, want at least 3", len(validationErr.Issues))
	}
}

func TestVoiceAudioTrackValidates(t *testing.T) {
	project := validProject()
	project.Audio[0].Kind = AudioTrackKindVoice
	project.Metadata.Selections[3].Kind = SelectionKindVoice

	if err := project.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestCutRequiresExactBoundary(t *testing.T) {
	project := validProject()
	project.Video.Clips[1].TimelineRange.Start = duration(5 * time.Second)
	project.Video.Transitions[0] = Transition{
		ID:         "transition-a-b",
		Kind:       TransitionKindCut,
		FromClipID: "clip-a",
		ToClipID:   "clip-b",
		Range:      testRange(5*time.Second, 0),
	}

	if err := project.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
