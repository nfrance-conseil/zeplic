// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Snapshot makes the structure of snapshot's names
//
package lib

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/calendar"
	"github.com/nfrance-conseil/zeplic/utils"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// CreateTime returns the date of snapshot's creation
func CreateTime(SnapshotName string) (int, time.Month, int, int, int, int) {
	SnapshotName = calendar.NumberMonth(SnapshotName)
	SnapshotName = utils.After(SnapshotName, "@")

	// Extract year, month and day
	date  := utils.Reverse(SnapshotName, "_")
	year  := utils.Before(date, "-")
	date   = date[5:len(date)]
	month := utils.Before(date, "-")
	day   := utils.Between(date, "-", "_")

	// Extract hour, minute and second
	timer := utils.After(SnapshotName, "_")
	hour  := utils.Before(timer, ":")
	timer  = timer[3:len(timer)]
	min   := utils.Before(timer, ":")
	sec   := timer[3:len(timer)]

	y, _ := strconv.Atoi(year)
	d, _ := strconv.Atoi(day)
	h, _ := strconv.Atoi(hour)
	m, _ := strconv.Atoi(min)
	s, _ := strconv.Atoi(sec)

	Month := calendar.NameMonthZero(month)
	return y, Month, d, h, m, s
}

// DatasetName returns the dataset name of snapshot
func DatasetName(SnapshotName string) string {
	dataset := utils.Before(SnapshotName, "@")
	return dataset
}

// InfoKV extracts the hostname, uuid, name and flag of snapshot KV pair
func InfoKV(pair string) (string, string, string) {
	uuid := utils.Before(pair, ":")
	name := utils.Reverse(pair, ":")
	var flag string
	if strings.Contains(name, "#") {
		flag = utils.Reverse(name, "#")
		name = utils.Before(name, "#")
	}
	return uuid, name, flag
}

// Prefix returns the prefix of snapshot name
func Prefix(SnapshotName string) string {
	prefix := utils.Between(SnapshotName, "@", "_")
	return prefix
}

// SnapName defines the name of the snapshot: PREFIX_yyyy-Month-dd_HH:MM:SS
func SnapName(prefix string) string {
	year, month, day := time.Now().Date()
	hour, min, sec   := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", prefix, year, month, day, hour, min, sec)
	return snapDate
}

// SnapBackup defines the name of a backup snapshot: BACKUP_from_yyyy-Month-dd
func SnapBackup(dataset string, ds *zfs.Dataset) string {
	// Get the older snapshot
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:82] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
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

// WasRenamed returns true if a snapshot was renamed
func WasRenamed(SnapshotReceived string, SnapshotToCheck string) bool {
	received := utils.After(SnapshotReceived, "@")
	toCheck := utils.After(SnapshotToCheck, "@")
	if received == toCheck {
		return false
	}
	return true
}

// WasCloned searchs the name of the dataset where a snapshot was cloned
func WasCloned(snap *zfs.Dataset) (bool, string) {
	var cloned bool
	clone, err := snap.GetProperty("clones")
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:120] it was not possible to find the clone of the snapshot '"+snap.Name+"'.")
	}
	if clone == "" {
		cloned = false
	} else {
		cloned = true
	}
	return cloned, clone
}
