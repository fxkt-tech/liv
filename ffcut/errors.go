package ffcut

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidProject     = errors.New("invalid ffcut project")
	ErrUnsupportedVersion = errors.New("unsupported ffcut project version")
	ErrInvalidDuration    = errors.New("invalid ffcut duration")
)

type ValidationIssue struct {
	Path    string
	Message string
	Cause   error
}

func (i ValidationIssue) Error() string {
	switch {
	case i.Path == "" && i.Cause != nil:
		return fmt.Sprintf("%s: %v", i.Message, i.Cause)
	case i.Path == "":
		return i.Message
	case i.Cause != nil:
		return fmt.Sprintf("%s: %s: %v", i.Path, i.Message, i.Cause)
	default:
		return fmt.Sprintf("%s: %s", i.Path, i.Message)
	}
}

func (i ValidationIssue) Unwrap() error {
	return i.Cause
}

type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrInvalidProject.Error()
	}
	parts := make([]string, len(e.Issues))
	for index, issue := range e.Issues {
		parts[index] = issue.Error()
	}
	return fmt.Sprintf("%s: %s", ErrInvalidProject, strings.Join(parts, "; "))
}

func (e *ValidationError) Is(target error) bool {
	if target == ErrInvalidProject {
		return true
	}
	for _, issue := range e.Issues {
		if errors.Is(issue.Cause, target) {
			return true
		}
	}
	return false
}

func (e *ValidationError) Unwrap() []error {
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

type CodecError struct {
	Operation string
	Err       error
}

func (e *CodecError) Error() string {
	if e == nil {
		return ErrInvalidProject.Error()
	}
	return fmt.Sprintf("%s ffcut project: %v", e.Operation, e.Err)
}

func (e *CodecError) Is(target error) bool {
	return target == ErrInvalidProject || errors.Is(e.Err, target)
}

func (e *CodecError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
