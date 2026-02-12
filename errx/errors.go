package errx

import "github.com/samber/oops"

type Code string

const (
	CodeInternal   Code = "internal"
	CodeUnknown    Code = "unknown"
	CodeNotFound   Code = "not_found"
	CodeConflict   Code = "conflict"
	CodeValidation Code = "validation"
	CodeZero       Code = "0"
)

func (c Code) String() string {
	return string(c)
}

//nolint:gochecknoglobals
var (
	ErrInternal   = oops.Code(string(CodeInternal)).Public("Internal error")
	ErrUnknown    = oops.Code(string(CodeUnknown)).Public("Unknown error")
	ErrNotFound   = oops.Code(string(CodeNotFound)).Public("Resource not found")
	ErrConflict   = oops.Code(string(CodeConflict)).Public("Resource already exists")
	ErrValidation = oops.Code(string(CodeValidation)).Public("Validation error")
)
