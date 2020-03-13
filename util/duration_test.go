package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMinutes(t *testing.T) {
	s := "3m"
	du, _ := ParseDuration(s)
	assert.Equal(t, 180.0, du.Seconds())
}

func TestHourMin(t *testing.T) {
	s := "1h1m"
	du, _ := ParseDuration(s)
	assert.Equal(t, (61 * time.Minute).Minutes(), du.Minutes())
}

func TestDayHour(t *testing.T) {
	s := "1d1h"
	du, _ := ParseDuration(s)
	assert.Equal(t, (25 * time.Hour).Minutes(), du.Minutes())

}
