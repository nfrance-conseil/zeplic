// Package calendar contains: cron.go - format.go - retention.go
//
// Retention extracts the information of JSON retention
//
package calendar

import (
	"strconv"

	"github.com/nfrance-conseil/zeplic/utils"
)

// Retention returns the policy retention
func Retention(retention string) (int, int, int, int) {
	D := utils.Before(retention, "d")
	retention = retention[len(D)+1:]
	W := utils.Before(retention, "w")
	retention = retention[len(W)+1:]
	M := utils.Before(retention, "m")
	retention = retention[len(M)+1:]
	Y := utils.Before(retention, "y")

	Dint, _ := strconv.Atoi(D)
	Wint, _ := strconv.Atoi(W)
	Mint, _ := strconv.Atoi(M)
	Yint, _ := strconv.Atoi(Y)

	return Dint, Wint, Mint, Yint
}
