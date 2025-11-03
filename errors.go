package exceltable

import "errors"

var (
	ErrNotStructType = errors.New("exceltable: not struct type")
	ErrUnknownMethod = errors.New("exceltable: unknown method")
	ErrInvalidMethod = errors.New("exceltable: invalid method")
	ErrUnknownTag    = errors.New("exceltable: unknown tag")
)
