package fs

const (
	GREAT_THAN = ">"
	LESS_THAN  = "<"
	EQUAL      = "=="
)

type FindControl struct {
	Size     int64
	SizeOp   string
	Apparent bool
	Name     string
	Flags    Bits
}
