package util

import (
	"testing"
)

func TestBinsToNum(t *testing.T) {
	s := "3t, 4k, 1m, 2g"
	nbins := BinsToNum(s)
	if nbins[0] != 4096 || nbins[1] != MiB || nbins[2] != 2*GiB || nbins[3] != 3*TiB {
		t.Errorf("Failed at converting %v", nbins)
	}
}
