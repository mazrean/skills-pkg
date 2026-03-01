package domain

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorConfigNotFound struct {
	Path string
}

func (e *ErrorConfigNotFound) Error() string {
	return fmt.Sprintf("configuration file not found at %s.", e.Path)
}

type ErrorSkillsNotFound struct {
	SkillNames []string
}

func (e *ErrorSkillsNotFound) Error() string {
	quatedNames := make([]string, 0, len(e.SkillNames))
	for _, name := range e.SkillNames {
		quatedNames = append(quatedNames, fmt.Sprintf("'%s'", name))
	}

	return fmt.Sprintf("skills %s not found in configuration.", strings.Join(quatedNames, ", "))
}

type ErrorConfigExists struct {
	Path string
}

func (e *ErrorConfigExists) Error() string {
	return fmt.Sprintf("configuration file already exists at %s", e.Path)
}

type ErrorSkillExists struct {
	SkillName string
}

func (e *ErrorSkillExists) Error() string {
	return fmt.Sprintf("skill '%s' already exists in configuration", e.SkillName)
}

type ErrorInvalidSource struct {
	SourceType string
}

func (e *ErrorInvalidSource) Error() string {
	if e.SourceType == "" {
		return "source type is empty. Supported types: git, go-mod"
	}
	return fmt.Sprintf("source type '%s' is not supported. Supported types: git, go-mod", e.SourceType)
}

type ErrorInvalidSkill struct {
	FieldName string
}

func (e *ErrorInvalidSkill) Error() string {
	return fmt.Sprintf("invalid skill configuration: field '%s' is required", e.FieldName)
}

type ErrorInstallTargetExists struct {
	Target string
}

func (e *ErrorInstallTargetExists) Error() string {
	return fmt.Sprintf("install target '%s' already exists in configuration", e.Target)
}

// Sentinel errors for domain-level error identification.
var (
	// ErrNetworkFailure indicates that a network request failed.
	ErrNetworkFailure = errors.New("network request failed")
)

// IsNetworkError checks if an error is a network-related error.
// It returns true if the error wraps ErrNetworkFailure.
func IsNetworkError(err error) bool {
	return errors.Is(err, ErrNetworkFailure)
}
