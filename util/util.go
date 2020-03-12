package util

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Define a set of common unit
const (
	KiB   = int64(1 << 10)
	MiB   = int64(1 << 20)
	GiB   = int64(1 << 30)
	TiB   = int64(1 << 40)
	PiB   = int64(1 << 50)
	EiB   = int64(1 << 60)
	GUARD = math.MaxInt64
)

var logger = NewLogger()

// StrBytes convert 4k to 4*KiB
func StrBytes(s string) int64 {
	idx := len(s) - 1
	num, _ := strconv.ParseInt(s[:idx], 10, 64)

	unit := string(s[idx])
	switch unit {
	case "c", "C":
		return num
	case "k", "K":
		return num * KiB
	case "m", "M":
		return num * MiB
	case "g", "G":
		return num * GiB
	case "t", "T":
		return num * TiB
	}
	return 0
}

// Item define what we can take
type Item struct {
	Name string
	Val  int64
}

// ItemList implement sort.Interface
type ItemList []Item

// Len returns length
func (a ItemList) Len() int { return len(a) }

// Less use Val
func (a ItemList) Less(i, j int) bool { return a[i].Val < a[j].Val }

// Swap two element
func (a ItemList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// SortedQueue provide an ordered queue with fixed length
type SortedQueue struct {
	capacity int
	items    ItemList
}

// NewSortedQueue returns a new queue
func NewSortedQueue(cap int) (oq *SortedQueue) {
	oq = new(SortedQueue)
	oq.capacity = cap
	oq.items = make([]Item, 0, cap) // fixed size
	return
}

// Put add new items
func (oq *SortedQueue) Put(it Item) {

	if len(oq.items) == oq.capacity {
		if oq.items[0].Val < it.Val {
			oq.items[0] = it
		}
	} else {
		oq.items = append(oq.items, it)
	}
	sort.Sort(oq.items)

}

// Items ... return items
func (oq *SortedQueue) Items() ItemList {
	return oq.items
}

// ShortNum ... shorten a number
func ShortNum(n int64) string {
	x := float64(n)

	K := float64(1000)
	M := float64(1000000)
	B := float64(1000000000)

	if x < K {
		return fmt.Sprintf("%d", n)
	}

	if x < M {
		return fmt.Sprintf("%.2f K", x/K)
	}

	if x < B {
		return fmt.Sprintf("%.2f Mi", x/M)
	}

	return fmt.Sprintf("%.2f B", x/B)

}

// ShortByte ... convert bytes to MiB GiB TiB
func ShortByte(sz int64) string {
	unitMap := map[string]int64{
		"mib": 1 << 20,
		"gib": 1 << 30,
		"tib": 1 << 40,
		"pib": 1 << 50,
		"eib": 1 << 60,
	}

	if sz < unitMap["mib"] {
		return fmt.Sprintf("%.2f KiB", float64(sz)/float64(1024))
	}
	if sz < unitMap["gib"] {
		return fmt.Sprintf("%.2f MiB", float64(sz)/float64(unitMap["mib"]))
	}
	if sz < unitMap["tib"] {
		return fmt.Sprintf("%.2f GiB", float64(sz)/float64(unitMap["gib"]))
	}
	if sz < unitMap["pib"] {
		return fmt.Sprintf("%.2f TiB", float64(sz)/float64(unitMap["tib"]))
	}
	if sz < unitMap["eib"] {
		return fmt.Sprintf("%.2f PiB", float64(sz)/float64(unitMap["pib"]))
	}

	return fmt.Sprintf("%.2f EiB", float64(sz)/float64(unitMap["eib"]))

}

// Comma produces a string form of the given number in base 10 with
// commas after every three orders of magnitude.
//
// e.g. Comma(834142) -> 834,142
func Comma(v int64) string {
	sign := ""

	// Min int64 can't be negated to a usable value, so it has to be special cased.
	if v == math.MinInt64 {
		return "-9,223,372,036,854,775,808"
	}

	if v < 0 {
		sign = "-"
		v = 0 - v
	}

	parts := []string{"", "", "", "", "", "", ""}
	j := len(parts) - 1

	for v > 999 {
		parts[j] = strconv.FormatInt(v%1000, 10)
		switch len(parts[j]) {
		case 2:
			parts[j] = "0" + parts[j]
		case 1:
			parts[j] = "00" + parts[j]
		}
		v = v / 1000
		j--
	}
	parts[j] = strconv.Itoa(int(v))
	return sign + strings.Join(parts[j:], ",")
}

// Commau for uint64 type
func Commau(v uint64) string {
	if v > math.MaxInt64 {
		log.Fatalln("Exceed max int64")
	}
	return Comma(int64(v))
}

// BinsToNum ... convert bin string to a sorted number slice
// Expect bins in 4k,8k ... must be one of {k,m,g,t} can't be
// Also, it provides GUARD as the largest int64.
func BinsToNum(bins string) []int64 {
	sbins := strings.Split(bins, ",")
	nbins := make([]int64, 0, len(sbins)+1) // plus one is for guard
	for _, v := range sbins {
		v = strings.TrimSpace(v)
		unit := strings.ToUpper(v[len(v)-1:])
		i, err := strconv.ParseInt(v[0:len(v)-1], 10, 64)
		if err != nil {
			log.Fatalf("Convert error: %v", err)
		}
		switch unit {
		case "K":
			nbins = append(nbins, i*KiB)
		case "M":
			nbins = append(nbins, i*MiB)
		case "G":
			nbins = append(nbins, i*GiB)
		case "T":
			nbins = append(nbins, i*TiB)
		default:
			log.Fatalf("Unknown unit in bins: %s\n", bins)
		}
	}
	nbins[len(nbins)-1] = GUARD
	sort.Slice(nbins, func(i, j int) bool { return nbins[i] < nbins[j] })
	return nbins
}

// InsertLeft ... evaluate x and increase counter
//
func InsertLeft(data []int64, counter []int64, x int64) {
	i := sort.Search(len(data),
		func(i int) bool {
			return x <= data[i]
		})
	if i < len(data) {
		counter[i]++
	} else {
		counter[len(data)-1]++
	}
}
