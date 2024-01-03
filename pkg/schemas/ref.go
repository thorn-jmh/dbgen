package schemas

import (
	"github.com/thorn-jmh/errorst"
)

var (
	ErrGetRefType           = errorst.NewError("cannot get $ref type")
	ErrUnsupportedRefSchema = errorst.Wrap(ErrGetRefType, "unsupported $ref schema")
)

type RefType string

const (
	RefTypeFile    RefType = "file"
	RefTypeHTTP    RefType = "http"
	RefTypeHTTPS   RefType = "https"
	RefTypeUnknown RefType = "unknown"
)
