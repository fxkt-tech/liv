package ffcut

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func Marshal(project *Project) ([]byte, error) {
	if project == nil {
		return nil, &CodecError{Operation: "encode", Err: fmt.Errorf("%w: project is required", ErrInvalidProject)}
	}
	if err := project.Validate(); err != nil {
		return nil, &CodecError{Operation: "encode", Err: err}
	}
	data, err := json.Marshal(project)
	if err != nil {
		return nil, &CodecError{Operation: "encode", Err: err}
	}
	return data, nil
}

func Unmarshal(data []byte) (*Project, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var project Project
	if err := decoder.Decode(&project); err != nil {
		return nil, &CodecError{Operation: "decode", Err: err}
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, &CodecError{Operation: "decode", Err: err}
	}
	if err := project.Validate(); err != nil {
		return nil, &CodecError{Operation: "decode", Err: err}
	}
	return &project, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return nil
}
