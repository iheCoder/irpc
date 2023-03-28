package codec

import "errors"

var (
	ErrRequestBodyNotPtr = errors.New("codec: req body not ptr")
)
