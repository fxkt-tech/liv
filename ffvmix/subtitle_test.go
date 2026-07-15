package ffvmix

import (
	"reflect"
	"testing"
	"time"
)

func TestSRTAndASSNormalizeToEquivalentCues(t *testing.T) {
	srt := []byte("1\r\n00:00:01,000 --> 00:00:02,500\r\nHello\r\nworld\r\n")
	ass := []byte("[Script Info]\nTitle: test\n\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\nDialogue: 0,0:00:01.00,0:00:02.50,Default,,0,0,0,,{\\b1}Hello\\Nworld\n")

	srtCues, err := parseSRT(srt, "subtitle")
	if err != nil {
		t.Fatalf("parseSRT() error = %v", err)
	}
	assCues, err := parseASS(ass, "subtitle")
	if err != nil {
		t.Fatalf("parseASS() error = %v", err)
	}
	if !reflect.DeepEqual(srtCues, assCues) {
		t.Fatalf("normalized cues differ\n SRT: %#v\n ASS: %#v", srtCues, assCues)
	}
	if srtCues[0].Range.Start.Std() != time.Second || srtCues[0].Range.Duration.Std() != 1500*time.Millisecond {
		t.Fatalf("cue range = %#v", srtCues[0].Range)
	}
}

func TestSubtitleParsersRejectMalformedInput(t *testing.T) {
	if _, err := parseSRT([]byte("not a subtitle"), "layer"); err == nil {
		t.Fatal("parseSRT() error = nil, want error")
	}
	if _, err := parseASS([]byte("[Events]\nDialogue: malformed"), "layer"); err == nil {
		t.Fatal("parseASS() error = nil, want error")
	}
}
