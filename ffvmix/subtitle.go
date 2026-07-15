package ffvmix

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

var (
	srtBlocks      = regexp.MustCompile(`\n[\t ]*\n+`)
	assOverrideTag = regexp.MustCompile(`\{[^}]*\}`)
)

func parseSubtitleFile(kind SubtitleInputKind, path string, layerID ID) ([]NormalizedCue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch kind {
	case SubtitleInputSRT:
		return parseSRT(data, layerID)
	case SubtitleInputASS:
		return parseASS(data, layerID)
	default:
		return nil, fmt.Errorf("unsupported subtitle input kind %q", kind)
	}
}

func parseSRT(data []byte, layerID ID) ([]NormalizedCue, error) {
	text := strings.TrimPrefix(string(data), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("SRT file contains no cues")
	}

	blocks := srtBlocks.Split(text, -1)
	cues := make([]NormalizedCue, 0, len(blocks))
	for blockIndex, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		timingIndex := 0
		if len(lines) > 0 && !strings.Contains(lines[0], "-->") {
			timingIndex = 1
		}
		if timingIndex >= len(lines) || !strings.Contains(lines[timingIndex], "-->") {
			return nil, fmt.Errorf("SRT block %d has no timing line", blockIndex+1)
		}
		parts := strings.SplitN(lines[timingIndex], "-->", 2)
		start, err := parseSRTTimestamp(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("SRT block %d start: %w", blockIndex+1, err)
		}
		endText := strings.TrimSpace(parts[1])
		if fields := strings.Fields(endText); len(fields) > 0 {
			endText = fields[0]
		}
		end, err := parseSRTTimestamp(endText)
		if err != nil {
			return nil, fmt.Errorf("SRT block %d end: %w", blockIndex+1, err)
		}
		if end <= start {
			return nil, fmt.Errorf("SRT block %d end must be after start", blockIndex+1)
		}
		cueText := strings.TrimSpace(strings.Join(lines[timingIndex+1:], "\n"))
		if cueText == "" {
			return nil, fmt.Errorf("SRT block %d text is empty", blockIndex+1)
		}
		startDuration, duration, err := subtitleRange(start, end)
		if err != nil {
			return nil, fmt.Errorf("SRT block %d: %w", blockIndex+1, err)
		}
		cues = append(cues, NormalizedCue{
			ID:    subtitleCueID(layerID, len(cues)),
			Range: ffcut.TimeRange{Start: startDuration, Duration: duration},
			Text:  cueText,
		})
	}
	return cues, nil
}

func parseSRTTimestamp(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	hours, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || hours < 0 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	minutes, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	secondParts := strings.FieldsFunc(parts[2], func(r rune) bool { return r == ',' || r == '.' })
	if len(secondParts) != 2 || len(secondParts[1]) != 3 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	seconds, err := strconv.ParseInt(secondParts[0], 10, 64)
	if err != nil || seconds < 0 || seconds >= 60 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	milliseconds, err := strconv.ParseInt(secondParts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second + time.Duration(milliseconds)*time.Millisecond, nil
}

func parseASS(data []byte, layerID ID) ([]NormalizedCue, error) {
	scanner := bufio.NewScanner(bytes.NewReader(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))))
	inEvents := false
	var format []string
	startIndex, endIndex, textIndex := -1, -1, -1
	cues := make([]NormalizedCue, 0)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inEvents = strings.EqualFold(line, "[Events]")
			continue
		}
		if !inEvents || line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		name, payload, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		switch {
		case strings.EqualFold(strings.TrimSpace(name), "Format"):
			format = splitAndTrim(payload)
			startIndex = fieldIndex(format, "Start")
			endIndex = fieldIndex(format, "End")
			textIndex = fieldIndex(format, "Text")
			if startIndex < 0 || endIndex < 0 || textIndex < 0 {
				return nil, fmt.Errorf("ASS Events Format must contain Start, End and Text")
			}
		case strings.EqualFold(strings.TrimSpace(name), "Dialogue"):
			if len(format) == 0 {
				return nil, fmt.Errorf("ASS line %d appears before Events Format", lineNumber)
			}
			fields := strings.SplitN(payload, ",", len(format))
			if len(fields) != len(format) {
				return nil, fmt.Errorf("ASS line %d has %d fields, want %d", lineNumber, len(fields), len(format))
			}
			start, err := parseASSTimestamp(strings.TrimSpace(fields[startIndex]))
			if err != nil {
				return nil, fmt.Errorf("ASS line %d start: %w", lineNumber, err)
			}
			end, err := parseASSTimestamp(strings.TrimSpace(fields[endIndex]))
			if err != nil {
				return nil, fmt.Errorf("ASS line %d end: %w", lineNumber, err)
			}
			if end <= start {
				return nil, fmt.Errorf("ASS line %d end must be after start", lineNumber)
			}
			cueText := normalizeASSText(fields[textIndex])
			if cueText == "" {
				return nil, fmt.Errorf("ASS line %d text is empty", lineNumber)
			}
			startDuration, duration, err := subtitleRange(start, end)
			if err != nil {
				return nil, fmt.Errorf("ASS line %d: %w", lineNumber, err)
			}
			cues = append(cues, NormalizedCue{
				ID:    subtitleCueID(layerID, len(cues)),
				Range: ffcut.TimeRange{Start: startDuration, Duration: duration},
				Text:  cueText,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(cues) == 0 {
		return nil, fmt.Errorf("ASS file contains no dialogue cues")
	}
	return cues, nil
}

func parseASSTimestamp(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	hours, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || hours < 0 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	minutes, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	secondParts := strings.Split(parts[2], ".")
	if len(secondParts) != 2 || len(secondParts[1]) < 1 || len(secondParts[1]) > 3 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	seconds, err := strconv.ParseInt(secondParts[0], 10, 64)
	if err != nil || seconds < 0 || seconds >= 60 {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	fraction, err := strconv.ParseInt(secondParts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp %q", value)
	}
	scale := time.Second
	for range len(secondParts[1]) {
		scale /= 10
	}
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second + time.Duration(fraction)*scale, nil
}

func subtitleRange(start, end time.Duration) (ffcut.Duration, ffcut.Duration, error) {
	protocolStart, err := ffcut.NewDuration(start)
	if err != nil {
		return 0, 0, err
	}
	protocolDuration, err := ffcut.NewDuration(end - start)
	if err != nil {
		return 0, 0, err
	}
	return protocolStart, protocolDuration, nil
}

func subtitleCueID(layerID ID, index int) ID {
	return ID(fmt.Sprintf("%s-cue-%06d", layerID, index+1))
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts
}

func fieldIndex(fields []string, name string) int {
	for index, field := range fields {
		if strings.EqualFold(field, name) {
			return index
		}
	}
	return -1
}

func normalizeASSText(value string) string {
	value = assOverrideTag.ReplaceAllString(value, "")
	value = strings.ReplaceAll(value, `\N`, "\n")
	value = strings.ReplaceAll(value, `\n`, "\n")
	value = strings.ReplaceAll(value, `\h`, " ")
	return strings.TrimSpace(value)
}
