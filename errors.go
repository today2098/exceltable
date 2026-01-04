package exceltable

import "errors"

// Sentinel errors.
var (
	ErrNotStructType       = errors.New("exceltable: not struct type")
	ErrNotFuncType         = errors.New("exceltable: not function type")
	ErrUnknownPredicateKey = errors.New("exceltable: unknown predicate key")
	ErrInvalidPredicate    = errors.New("exceltable: invalid predicate")
)
