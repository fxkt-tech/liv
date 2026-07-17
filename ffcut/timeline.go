package ffcut

type Sequence struct {
	Clips       []VideoClip  `json:"clips"`
	Transitions []Transition `json:"transitions,omitempty"`
}

type VideoClip struct {
	ID            ID          `json:"id"`
	Source        LocalSource `json:"source"`
	SourceRange   TimeRange   `json:"source_range"`
	TimelineRange TimeRange   `json:"timeline_range"`
	Playback      Playback    `json:"playback"`
	Fit           FitMode     `json:"fit"`
	Audio         ClipAudio   `json:"audio"`
}

type Playback struct {
	// Rate maps source time to timeline time. For non-looping clips,
	// source duration / rate + freeze duration must equal timeline duration.
	Rate            float64  `json:"rate"`
	Loop            bool     `json:"loop,omitempty"`
	FreezeLastFrame Duration `json:"freeze_last_frame,omitempty"`
}

type ClipAudio struct {
	Enabled bool    `json:"enabled"`
	Gain    float64 `json:"gain"`
}

type TransitionKind string

const (
	TransitionKindCut  TransitionKind = "cut"
	TransitionKindFade TransitionKind = "fade"
)

type Transition struct {
	ID             ID             `json:"id"`
	Kind           TransitionKind `json:"kind"`
	FromClipID     ID             `json:"from_clip_id"`
	ToClipID       ID             `json:"to_clip_id"`
	Range          TimeRange      `json:"range"`
	AudioCrossfade bool           `json:"audio_crossfade"`
}

type AudioTrackKind string

const AudioTrackKindBGM AudioTrackKind = "bgm"

const AudioTrackKindVoice AudioTrackKind = "voice"

type AudioTrack struct {
	ID            ID             `json:"id"`
	Kind          AudioTrackKind `json:"kind"`
	Source        LocalSource    `json:"source"`
	SourceRange   TimeRange      `json:"source_range"`
	TimelineRange TimeRange      `json:"timeline_range"`
	Loop          bool           `json:"loop,omitempty"`
	Gain          float64        `json:"gain"`
	FadeIn        Duration       `json:"fade_in,omitempty"`
	FadeOut       Duration       `json:"fade_out,omitempty"`
}
