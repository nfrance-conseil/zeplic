// Package lib contains: commands.go - destroy.go - !policy.go - snapshot.go - take.go - uuid.go
//
// Snapshot makes the structure of snapshot's names
//
package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
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
	hour, min, sec   := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", name, year, month, day, hour, min, sec)
	return snapDate
}

// SnapBackup defines the name of a backup snapshot: BACKUP_from_yyyy-Month-dd
func SnapBackup(dataset string, ds *zfs.Dataset) string {
	// Get the older snapshot
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:35] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
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

// Renamed returns true if a snapshot was renamed
func WasRenamed(SnapshotReceived string, SnapshotToCheck string) bool {
	received := utils.After(SnapshotReceived, "@")
	toCheck := utils.After(SnapshotToCheck, "@")
	if received == toCheck {
		return false
	}
	return true
}

// Cloned searchs the name of the dataset where a snapshot was cloned
func WasCloned(snap *zfs.Dataset) (bool, string) {
	var cloned bool
	clone, err := snap.GetProperty("clones")
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:73] it was not possible to find the clone of the snapshot '"+snap.Name+"'.")
	}
	if clone == "" {
		cloned = false
	} else {
		cloned = true
	}
	return cloned, clone
}

// CreateTime returns the date of snapshot's creation
func CreateTime(snap *zfs.Dataset) (int64, int, string, int, int, int, int) {
	// Extract year, month and day
	date  := utils.Reverse(snap.Name, "_")
	year  := utils.Before(date, "-")
	date   = date[5:len(date)]
	month := utils.Before(date, "-")
	day   := utils.Between(date, "-", "_")

	// Extract hour, minute and second
	timer := utils.After(snap.Name, "_")
	hour  := utils.Before(timer, ":")
	timer  = timer[3:len(timer)]
	min   := utils.Before(timer, ":")
	sec   := timer[3:len(timer)]

	y, _ := strconv.Atoi(year)
	d, _ := strconv.Atoi(day)
	h, _ := strconv.Atoi(hour)
	m, _ := strconv.Atoi(min)
	s, _ := strconv.Atoi(sec)

	search := fmt.Sprintf("zfs get -rHp -t snapshot -o name,value creation | awk '{if ($1 == \"%s\") print $2}'", snap.Name)
	cmd, err := exec.Command("sh", "-c", search).Output()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:108] it was not possible to execute the command 'zfs get creation'.")
	}
	out := bytes.Trim(cmd, "\x0A")
	creation := string(out)
	unixTimeSnap, _ := strconv.ParseInt(creation, 10, 64)

	return unixTimeSnap, y, month, d, h, m, s
}
