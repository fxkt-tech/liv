package ffvmix

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidTemplate = errors.New("invalid ffvmix template")
	ErrCompile         = errors.New("ffvmix template compilation failed")
)

type IssueCode string

const (
	IssueInvalidValue         IssueCode = "invalid_value"
	IssueInvalidID            IssueCode = "invalid_id"
	IssueInvalidReference     IssueCode = "invalid_reference"
	IssueInvalidUnion         IssueCode = "invalid_union"
	IssuePathResolution       IssueCode = "path_resolution"
	IssueFileStat             IssueCode = "file_stat"
	IssueFingerprint          IssueCode = "fingerprint"
	IssueProbe                IssueCode = "probe"
	IssueMissingVideo         IssueCode = "missing_video"
	IssueMissingAudio         IssueCode = "missing_audio"
	IssueSourceRange          IssueCode = "source_range"
	IssueSubtitleParse        IssueCode = "subtitle_parse"
	IssueNoFeasibleSource     IssueCode = "no_feasible_source"
	IssueNoFeasibleTransition IssueCode = "no_feasible_transition"
	IssueCanceled             IssueCode = "canceled"
)

type Issue struct {
	Code      IssueCode
	Path      string
	LocalPath string
	Message   string
	Cause     error
}

func (i Issue) Error() string {
	parts := make([]string, 0, 4)
	if i.Path != "" {
		parts = append(parts, i.Path)
	}
	if i.LocalPath != "" {
		parts = append(parts, fmt.Sprintf("local path %q", i.LocalPath))
	}
	if i.Message != "" {
		parts = append(parts, i.Message)
	}
	if i.Cause != nil && !errors.Is(i.Cause, ErrInvalidTemplate) {
		parts = append(parts, i.Cause.Error())
	}
	if len(parts) == 0 {
		return string(i.Code)
	}
	return strings.Join(parts, ": ")
}

func (i Issue) Unwrap() error {
	return i.Cause
}

type CompileError struct {
	Issues []Issue
}

func (e *CompileError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrCompile.Error()
	}
	parts := make([]string, len(e.Issues))
	for index, issue := range e.Issues {
		parts[index] = issue.Error()
	}
	return fmt.Sprintf("%s: %s", ErrCompile, strings.Join(parts, "; "))
}

func (e *CompileError) Is(target error) bool {
	if target == ErrCompile {
		return true
	}
	if e == nil {
		return false
	}
	for _, issue := range e.Issues {
		if errors.Is(issue.Cause, target) {
			return true
		}
	}
	return false
}

func (e *CompileError) Unwrap() []error {
	if e == nil {
		return nil
	}
	causes := make([]error, 0, len(e.Issues))
	for _, issue := range e.Issues {
		if issue.Cause != nil {
			causes = append(causes, issue.Cause)
		}
	}
	return causes
}

type TemplateCodecError struct {
	Operation string
	Err       error
}

func (e *TemplateCodecError) Error() string {
	if e == nil {
		return ErrInvalidTemplate.Error()
	}
	return fmt.Sprintf("%s ffvmix template: %v", e.Operation, e.Err)
}

func (e *TemplateCodecError) Is(target error) bool {
	return target == ErrInvalidTemplate || errors.Is(e.Err, target)
}

func (e *TemplateCodecError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
