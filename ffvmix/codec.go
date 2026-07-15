package ffvmix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func MarshalTemplate(template *Template) ([]byte, error) {
	if template == nil {
		return nil, &TemplateCodecError{Operation: "encode", Err: ErrInvalidTemplate}
	}
	if err := template.Validate(); err != nil {
		return nil, &TemplateCodecError{Operation: "encode", Err: err}
	}
	data, err := json.Marshal(template)
	if err != nil {
		return nil, &TemplateCodecError{Operation: "encode", Err: err}
	}
	return data, nil
}

func UnmarshalTemplate(data []byte) (*Template, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var template Template
	if err := decoder.Decode(&template); err != nil {
		return nil, &TemplateCodecError{Operation: "decode", Err: err}
	}
	if err := ensureTemplateJSONEOF(decoder); err != nil {
		return nil, &TemplateCodecError{Operation: "decode", Err: err}
	}
	if err := template.Validate(); err != nil {
		return nil, &TemplateCodecError{Operation: "decode", Err: err}
	}
	return &template, nil
}

func ensureTemplateJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return nil
}
