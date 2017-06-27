// Package lib contains: clones.go - commands.go - snapshot.go - uuid.go
//
// Snapshot makes the structure of snapshot's names
//
package lib

import (
	"fmt"
	"time"

	"github.com/nfrance-conseil/zeplic/utils"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// DatasetName returns the dataset name of snapshot
func DatasetName(SnapshotName string) string {
	dataset := utils.Before(SnapshotName, "@")
	return dataset
}

// SnapName defines the name of the snapshot: NAME_yyyy-Month-dd_HH:MM:SS
func SnapName(name string) string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", name, year, month, day, hour, min, sec)
	return snapDate
}

// SnapBackup defines the name of a backup snapshot: BACKUP_from_yyyy-Month-dd
func SnapBackup(dataset string, ds *zfs.Dataset) string {
	// Get the older snapshot
	list, _ := ds.Snapshots()
	count := len(list)

	var oldSnapshot string
	for i := 0; i < count; i++ {
		take := list[i].Name
		dsName := DatasetName(take)
		if dsName == dataset {
			oldSnapshot = take
			break
		} else {
			continue
		}
	}

	// Get date
	rev := utils.Reverse(oldSnapshot, "_")
	date := utils.Before(rev, "_")
	backup := fmt.Sprintf("%s_%s", "BACKUP_from", date)
	return backup
}

// Renamed returns true if a snapshot was renamed
func Renamed(SnapshotReceived string, SnapshotToCheck string) bool {
	received := utils.After(SnapshotReceived, "@")
	toCheck := utils.After(SnapshotToCheck, "@")
	if received == toCheck {
		return false
	}
	return true
}
