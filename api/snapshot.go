// Package api contains: commands.go - snapshot.go
//
// Snapshot makes the structure of snapshot's names
//
package api

import (
	"fmt"
	"time"
)

// SnapName defines the name of the snapshot: NAME_yyyy-Month-dd_HH:MM:SS
func SnapName(name string) string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", name, year, month, day, hour, min, sec)
	return snapDate
}

// SnapBackup defines the name of a backup snapshot: BACKUP_yyyy-Month-dd_HH:MM:SS
func SnapBackup() string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	backup := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", "BACKUP", year, month, day, hour, min, sec)
	return backup
}
