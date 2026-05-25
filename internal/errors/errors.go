package errors

import "errors"

var (
	WordNotFoundError = errors.New("word is not a stop word")
	InvalidWordError  = errors.New("word is empty")
)
