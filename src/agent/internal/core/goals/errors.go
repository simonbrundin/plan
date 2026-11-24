package goals

import "errors"

var (
	ErrEmptyTitle = errors.New("goal title cannot be empty")
	ErrEmptyID    = errors.New("goal id cannot be empty")
	ErrNotFound   = errors.New("goal not found")
	ErrDuplicate  = errors.New("goal with this id already exists")
)
