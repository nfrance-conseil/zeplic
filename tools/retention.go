// Package tools contains: calendar.go - cron.go - retention.go - substrings.go
//
// Retention extracts the information of JSON retention
//
package tools

import (
	"strconv"
)

// Retention returns the policy retention
func Retention(retention string) (int, int, int, int) {
	D := Before(retention, "d")
	retention = retention[len(D)+1:]
	W := Before(retention, "w")
	retention = retention[len(W)+1:]
	M := Before(retention, "m")
	retention = retention[len(M)+1:]
	Y := Before(retention, "y")

	Dint, _ := strconv.Atoi(D)
	Wint, _ := strconv.Atoi(W)
	Mint, _ := strconv.Atoi(M)
	Yint, _ := strconv.Atoi(Y)

	return Dint, Wint, Mint, Yint
}
