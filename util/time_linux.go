package util

import (
	"syscall"
	"time"
)

func StatsTime(stat *syscall.Stat_t) (atime, ctime, mtime time.Time) {
	atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	mtime = time.Unix(int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec))
	return
}
