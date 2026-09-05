package tandoor

import "errors"

var (
	ErrTandoorUnauthorized  = errors.New("tandoor request unauthorized")
	ErrTandoorNotFound      = errors.New("tandoor resource not found")
	ErrTandoorRequestFailed = errors.New("tandoor request failed")
)
