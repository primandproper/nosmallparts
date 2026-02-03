package database

import (
	"errors"
)

var (
	// ErrNilInputProvided indicates nil input was provided in an unacceptable context.
	ErrNilInputProvided = errors.New("nil input provided")

	// ErrInvalidIDProvided indicates a required ID was passed in empty.
	ErrInvalidIDProvided = errors.New("required ID provided is empty")

	// ErrEmptyInputProvided indicates a required input was passed in empty.
	ErrEmptyInputProvided = errors.New("input provided is empty")

	// ErrUserAlreadyExists indicates that a user with that username has already been created.
	ErrUserAlreadyExists = errors.New("user already exists")
)
