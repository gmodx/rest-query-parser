package restqueryparser

import "errors"

var (
	ErrBadFormat  = errors.New("bad format")
	ErrNotInScope = errors.New("not in scope")
	ErrEmptyValue = errors.New("empty value")
)
