// Snapshot makes the structure of snapshot's names
package api

import (
	"fmt"
	"strings"
	"time"
)

// SnapName() define the name of the snapshot: NAME_yyyy-Month-dd_HH:MM:SS
func SnapName(name string) string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", name, year, month, day, hour, min, sec)
	return snapDate
}

// SnapBackup() define the name of a backup snapshot: BACKUP_yyyy-Month-dd_HH:MM:SS
func SnapBackup() string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	backup := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", "BACKUP", year, month, day, hour, min, sec)
	return backup
}

// Get substring with the name of the snapshots
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

// Get substring before a string
func Before(value string, a string) string {
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func Chop(r int, s string) string {
	return s[r:]
}
