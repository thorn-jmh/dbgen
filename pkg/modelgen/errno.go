package modelgen

import "github.com/thorn-jmh/errorst"

var (
	ErrInvalidStructure = errorst.NewError("definition tree not allowed")
	ErrWrongSyntax      = errorst.NewError("Syntax Error!")
)
