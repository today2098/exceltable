package exceltable

import "errors"

// Sentinel errors.
var (
	ErrNotStructType    = errors.New("exceltable: not struct type")
	ErrUnknownPredicate = errors.New("exceltable: unknown predicate method")
	ErrInvalidPredicate = errors.New("exceltable: invalid predicate method")
)
