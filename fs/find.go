package fs

import (
	"time"
)

const (
	GREAT_THAN = ">"
	LESS_THAN  = "<"
	EQUAL      = "=="
)

type FindControl struct {
	Size       int64
	SizeOp     string
	Apparent   bool
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Flags      Bits
	DeleteFlag bool
}
