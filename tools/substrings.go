// Package tools contains: calendar.go - cron.go - retention.go - substrings.go
//
// Functions to get substrings
// Sort a list of snapshots
//
package tools

import (
	"sort"
	"strings"
)

// Arrange sorts the list of snapshots
func Arrange(SnapshotsList []string) []string {
	for i := 0; i < len(SnapshotsList); i++ {
		SnapshotsList[i] = NumberMonth(SnapshotsList[i])
	}
	sort.Strings(SnapshotsList)
	return SnapshotsList
}

// After gets substring after a string
func After(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

// Before gets substring before a string
func Before(value string, a string) string {
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

// Between gets a substring between two strings
func Between(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

// Reverse gets substring
func Reverse(value string, a string) string {
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}
