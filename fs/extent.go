package fs

//
// ioctl(2) with FS_IOC_FIEMAP
// to obtain a map of file extends excluding holes
//

import (
	"math"
	"os"
)

const (
	FIEMAP_MAX_OFFSET            = math.MaxInt64
	FIEMAP_FLAG_SYNC             = 0x00000001 // Sync file before map
	FIEMAP_FLAG_XATTR            = 0x00000002 // Map extended attribute tree
	FIEMAP_FLAGS_COMPAT          = (FIEMAP_FLAG_SYNC | FIEMAP_FLAG_XATTR)
	FIEMAP_EXTENT_LAST           = 0x00000001 // Last extent in file
	FIEMAP_EXTENT_UNKNOWN        = 0x00000002 // Data location unknown
	FIEMAP_EXTENT_DELALLOC       = 0x00000004 // Location pending
	FIEMAP_EXTENT_ENCODED        = 0x00000008 // Data cannot be read when fs is unmounte
	FIEMAP_EXTENT_DATA_ENCRYPTED = 0x00000080 // Data encrypted by fs
	FIEMAP_EXTENT_NOT_ALIGNED    = 0x00000100 // Extent offset may not be block aligned
	FIEMAP_EXTENT_DATA_INLINE    = 0x00000200 // Data mixed with metadata
	FIEMAP_EXTENT_DATA_TAIL      = 0x00000400 // Multiple files in block
	FIEMAP_EXTENT_UNWRITTEN      = 0x00000800 // Space allocated, but not data
	FIEMAP_EXTENT_MERGED         = 0x00001000 // File doesn not natively support extents, results merged
	FIEMAP_EXTENT_SHARED         = 0x00002000 // Space shared with other files
)

type FieMap struct {
	fm_start          uint64 // Logical offset (inclusive) to start mapping
	fm_length         uint64 // Logical length of mapping
	fm_flags          uint32 // FIEMAP_FLAG_* flags for request
	fm_mapped_extents uint32 // Number of extents that were mapped out
	fm_extent_count   uint32 // Size of extents
	fm_reserved       uint32
}

type ExtentInfo struct {
	Ext_logical int64
	Ext_length  int64
	Ext_flags   int
}

type ExtentScan struct {
	fileptr          *os.File
	fd               int
	scan_start       int64
	ext_count        int64 // how many extent returned
	init_scan_ok     bool  // fail flag for scan
	hit_final_extent bool  // true when scan is all good
	exts             []ExtentInfo
}

// func ExtentScanInit(file *osFile) (scan *ExtentScan) {
// 	scan.fd = file.Fd()
// 	scan.fileptr = file

// }

// func ExtentScanRead() {

// }
