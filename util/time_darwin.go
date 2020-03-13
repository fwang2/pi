package util

import (
	"syscall"
	"time"
)

func StatsTime(stat *syscall.Stat_t) (atime, ctime, mtime time.Time) {
	atime = time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
	ctime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	mtime = time.Unix(int64(stat.Mtimespec.Sec), int64(stat.Mtimespec.Nsec))
	return
}
