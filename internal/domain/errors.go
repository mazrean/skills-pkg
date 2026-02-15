package domain

import "errors"

// Sentinel errors for domain-level error identification.
// These errors provide a standard way to identify and report error conditions
// across the application, supporting requirements 12.2 and 12.3.
var (
	// ErrConfigNotFound indicates that the configuration file was not found.
	ErrConfigNotFound = errors.New("configuration file not found")

	// ErrConfigExists indicates that a configuration file already exists.
	ErrConfigExists = errors.New("configuration file already exists")

	// ErrSkillNotFound indicates that the requested skill was not found.
	ErrSkillNotFound = errors.New("skill not found")

	// ErrHashMismatch indicates that hash verification failed.
	ErrHashMismatch = errors.New("hash mismatch detected")

	// ErrInvalidSource indicates that an invalid source type was specified.
	ErrInvalidSource = errors.New("invalid source type")

	// ErrNetworkFailure indicates that a network request failed.
	ErrNetworkFailure = errors.New("network request failed")

	// ErrSkillExists indicates that a skill with the same name already exists.
	ErrSkillExists = errors.New("skill already exists")

	// ErrInvalidSkill indicates that a skill has invalid field values.
	ErrInvalidSkill = errors.New("invalid skill configuration")
)
