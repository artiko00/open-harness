package fixture

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("already exists")
)

func describeError(err error) string {
	switch err {
	case ErrNotFound:
		return "the requested resource was not found"
	case ErrUnauthorized:
		return "the caller is not authorized for this action"
	case ErrInvalidInput:
		return "the provided input did not pass validation"
	case ErrAlreadyExists:
		return "a resource with the same identifier already exists"
	}
	return "unknown error"
}
