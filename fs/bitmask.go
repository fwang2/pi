package fs

type Bits uint32

const (
	FB_NAME Bits = 1 << iota
	FB_SIZE
	FB_TYPE_F // file
	FB_TYPE_D // directory
	FB_TYPE_L // sym link
	FB_TYPE_P // pipe
	FB_TYPE_A // all
	FB_MTIME
	FB_ATIME
	FB_CTIME
)

func Set(b, flag Bits) Bits    { return b | flag }
func Clear(b, flag Bits) Bits  { return b &^ flag }
func Toggle(b, flag Bits) Bits { return b ^ flag }
func Has(b, flag Bits) bool    { return b&flag != 0 }
func Empty(b Bits) bool        { return b == 0 }
