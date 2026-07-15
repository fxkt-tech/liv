package ffcut

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestMarshalRoundTrip(t *testing.T) {
	want := validProject()
	data, err := Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch\n got: %#v\nwant: %#v", got, want)
	}
	if got.Layers[0].ID != "watermark-layer" || got.Layers[1].ID != "subtitle-layer" {
		t.Fatalf("layer order changed: %v", []ID{got.Layers[0].ID, got.Layers[1].ID})
	}
}

func TestMarshalProjectV2Schema(t *testing.T) {
	project := validProject()
	project.Video.Clips = project.Video.Clips[:1]
	project.Video.Transitions = nil
	project.Audio = nil
	project.Layers = nil
	project.Metadata.Selections = project.Metadata.Selections[:1]
	project.Metadata.Constraints = nil

	data, err := Marshal(project)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	want := `{"version":2,"id":"project-1","canvas":{"width":1920,"height":1080,"frame_rate":{"numerator":30,"denominator":1},"background":{"kind":"color","color":{"color":"#000000"}}},"video":{"clips":[{"id":"clip-a","source":{"id":"source-a","path":"/media/a.mp4","fingerprint":{"size":1000,"modified_unix_nano":1}},"source_range":{"start":0,"duration":5000000},"timeline_range":{"start":0,"duration":5000000},"playback":{"rate":1},"fit":"cover","audio":{"enabled":true,"gain":1}}]},"metadata":{"template_fingerprint":"template-sha256","seed":42,"combination_fingerprint":"combination-sha256","selections":[{"kind":"video","dimension_id":"slot-a","option_id":"candidate-a","asset_fingerprint":"asset-a"}]}}`
	if string(data) != want {
		t.Fatalf("Marshal() schema changed\n got: %s\nwant: %s", data, want)
	}
}

func TestMarshalDoesNotMutateProject(t *testing.T) {
	project := validProject()
	before, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("json.Marshal(before) error = %v", err)
	}

	if _, err := Marshal(project); err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	after, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("json.Marshal(after) error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("Marshal() mutated project\nbefore: %s\n after: %s", before, after)
	}
}

func TestMarshalRejectsNilAndInvalidProjects(t *testing.T) {
	if _, err := Marshal(nil); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("Marshal(nil) error = %v, want ErrInvalidProject", err)
	}

	project := validProject()
	project.ID = ""
	if _, err := Marshal(project); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("Marshal(invalid) error = %v, want ErrInvalidProject", err)
	}
}

func TestUnmarshalRejectsUnknownFieldsAndTrailingValues(t *testing.T) {
	data, err := Marshal(validProject())
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	withUnknown := strings.TrimSuffix(string(data), "}") + `,"unknown":true}`
	if _, err := Unmarshal([]byte(withUnknown)); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("Unmarshal(unknown) error = %v, want ErrInvalidProject", err)
	}
	if _, err := Unmarshal(append(data, []byte(` {}`)...)); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("Unmarshal(trailing) error = %v, want ErrInvalidProject", err)
	}
}

func TestUnmarshalPreservesVersionCause(t *testing.T) {
	data, err := Marshal(validProject())
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	data = []byte(strings.Replace(string(data), `"version":2`, `"version":3`, 1))

	_, err = Unmarshal(data)
	if !errors.Is(err, ErrInvalidProject) || !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("Unmarshal() error = %v, want invalid project and unsupported version", err)
	}
}
