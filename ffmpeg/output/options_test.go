package output

import "testing"

func TestVideoProfile(t *testing.T) {
	got := New(VideoProfile("high")).Params()

	if !hasAdjacentParams(got, "-profile:v", "high") {
		t.Fatalf("VideoProfile() params = %v, want -profile:v high", got)
	}
}

func TestAudioProfile(t *testing.T) {
	got := New(AudioProfile("aac_low")).Params()

	if !hasAdjacentParams(got, "-profile:a", "aac_low") {
		t.Fatalf("AudioProfile() params = %v, want -profile:a aac_low", got)
	}
}

func hasAdjacentParams(params []string, key string, value string) bool {
	for i := 0; i+1 < len(params); i++ {
		if params[i] == key && params[i+1] == value {
			return true
		}
	}
	return false
}
