package services

import "errors"

var (
	// ErrInvalidStory indicates that the story provided is nil or invalid.
	ErrInvalidStory = errors.New("story is nil or invalid")

	// ErrInvalidInput indicates that required input is missing or invalid.
	ErrInvalidInput = errors.New("required input is missing or invalid")

	// ErrContextCancelled indicates that the context was cancelled during operation.
	ErrContextCancelled = errors.New("context was cancelled")
)
