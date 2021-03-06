// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go
//
// Snapshot provides the functions to get the properties of snapshots
//
package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/tools"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
	"github.com/pborman/uuid"
)

// CreateTime returns the date of snapshot's creation
func CreateTime(SnapshotName string) (int, time.Month, int, int, int, int) {
	SnapshotName = tools.NumberMonth(SnapshotName)
	SnapshotName = tools.After(SnapshotName, "@")

	// Extract year, month and day
	date  := tools.Reverse(SnapshotName, "_")
	year  := tools.Before(date, "-")
	date   = date[5:len(date)]
	month := tools.Before(date, "-")
	day   := tools.Between(date, "-", "_")

	// Extract hour, minute and second
	timer := tools.After(SnapshotName, "_")
	hour  := tools.Before(timer, ":")
	timer  = timer[3:len(timer)]
	min   := tools.Before(timer, ":")
	sec   := timer[3:len(timer)]

	y, _ := strconv.Atoi(year)
	d, _ := strconv.Atoi(day)
	h, _ := strconv.Atoi(hour)
	m, _ := strconv.Atoi(min)
	s, _ := strconv.Atoi(sec)

	Month := tools.NameMonthZero(month)
	return y, Month, d, h, m, s
}

// DatasetName returns the dataset name of snapshot
func DatasetName(SnapshotName string) string {
	dataset := tools.Before(SnapshotName, "@")
	return dataset
}

// InfoKV extracts the hostname, uuid, name and flags of snapshot KV pair
func InfoKV(pair string) (string, string, string) {
	uuid := tools.Before(pair, ":")
	name := tools.Reverse(pair, ":")
	var flag string
	if strings.Contains(name, "#") {
		flag = tools.Reverse(name, "#")
		flag = fmt.Sprintf("#%s", flag)
		name = tools.Before(name, "#")
	}
	return uuid, name, flag
}

// LastSnapshot returns the name of last snapshot in 'dataset'
func LastSnapshot(ds *zfs.Dataset, prefix string) string {
	var LastSnapshot string

	// List of snapshots
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:73] it was not possible to access of snapshots list in dataset '"+ds.Name+"'.")
	} else {
		// Find the correct snapshot
		for i := len(list)-1; i > -1; i-- {
			RealDataset := DatasetName(list[i].Name)
			RealPrefix := Prefix(list[i].Name)
			if RealDataset == ds.Name && RealPrefix == prefix {
				LastSnapshot = list[i].Name
				break
			} else {
				continue
			}
		}
	}
	return LastSnapshot
}

// Prefix returns the prefix of snapshot name
func Prefix(SnapshotName string) string {
	var prefix string
	if strings.Contains(SnapshotName, "@") {
		prefix = tools.After(SnapshotName, "@")
		prefix = tools.Before(prefix, "_")
	} else {
		prefix = tools.Before(SnapshotName, "_")
	}
	return prefix
}

// RealList returns the correct amount of snapshots and the index of backup snapshot
func RealList(ds *zfs.Dataset, prefix string) (int, []int) {
	var amount []int
	backup := -1

	// List of snapshots
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:110] it was not possible to access of snapshots list in dataset '"+ds.Name+"'.")
	} else {
		// Check the number of snapshot in the correct dataset
		for i := 0; i < len(list); i++ {
			// Check the dataset
			RealDataset := DatasetName(list[i].Name)
			RealPrefix := Prefix(list[i].Name)
			if RealDataset == ds.Name {
				// Is it the backup snapshot?
				if prefix != "" {
					if RealPrefix == "BACKUP" {
						backup = i
					} else if RealPrefix == prefix {
						amount = append(amount, i)
					}
					continue
				} else {
					if RealPrefix == "BACKUP" {
						backup = i
					} else {
						amount = append(amount, i)
					}
					continue
				}
			} else {
				continue
			}
		}
	}
	return backup, amount
}

// SearchName searchs the name of snapshot from its uuid
func SearchName(uuid string) string {
	var snapshot string
	search := fmt.Sprintf("zfs get -rHp -t snapshot -o name,value :uuid | awk '{if ($2 == \"%s\") print $1}'", uuid)
	cmd, err := exec.Command("sh", "-c", search).Output()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:148] it was not possible to execute the command 'zfs get :uuid'.")
	} else {
		out := bytes.Trim(cmd, "\x0A")
		snapshot = string(out)
	}
	return snapshot
}

// SearchUUID searchs the uuid of snapshot from its name
func SearchUUID(snap *zfs.Dataset) string {
	uuid, err := snap.GetProperty(":uuid")
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:160] it was not possible to find the uuid of the snapshot '"+snap.Name+"'.")
	}
	return uuid
}

// SnapName defines the name of the snapshot: PREFIX_yyyy-Month-dd_HH:MM:SS
func SnapName(prefix string) string {
	year, month, day := time.Now().Date()
	hour, min, sec   := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", prefix, year, month, day, hour, min, sec)
	return snapDate
}

// SnapBackup defines the name of a backup snapshot: BACKUP_from_yyyy-Month-dd
func SnapBackup(ds *zfs.Dataset) string {
	var backup string
	// Get the older snapshot
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:179] it was not possible to access of snapshots list in dataset '"+ds.Name+"'.")
	} else {
		_, amount := RealList(ds, "")
		OlderSnapshot := list[amount[0]].Name

		// Get date of last snapshot
		OlderSnapshot = tools.After(OlderSnapshot, "@")
		date := tools.Reverse(OlderSnapshot, "_")
		date = tools.Before(date, "_")
		backup = fmt.Sprintf("%s_%s", "BACKUP_from", date)
	}
	return backup
}

// SnapRenamed returns true if a snapshot was renamed
func SnapRenamed(SnapshotReceived string, SnapshotToCheck string) bool {
	received := tools.After(SnapshotReceived, "@")
	toCheck := tools.After(SnapshotToCheck, "@")
	if received == toCheck {
		return false
	}
	return true
}

// SnapCloned searchs the name of the dataset where a snapshot was cloned
func SnapCloned(snap *zfs.Dataset) (bool, string) {
	var cloned bool
	clone, err := snap.GetProperty("clones")
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:208] it was not possible to find the clone of the snapshot '"+snap.Name+"'.")
	} else {
		if clone == "" {
			cloned = false
		} else {
			cloned = true
		}
	}
	return cloned, clone
}

// Source returns if a snapshot has the status local or received
func Source(uuid string) string {
	var source string
	search := fmt.Sprintf("zfs get -rHp -t snapshot -o value,source :uuid | awk '{if ($1 == \"%s\") print $2}'", uuid)
	cmd, err := exec.Command("sh", "-c", search).Output()
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:225] it was not possible to execute the command 'zfs get :uuid'.")
	} else {
		out := bytes.Trim(cmd, "\x0A")
		source = string(out)
	}
	return source
}

// UUID asigns a new uuid
func UUID(snap *zfs.Dataset) error {
	id := uuid.New()
	err := snap.SetProperty(":uuid", id)
	return err
}

// Written search changes in dataset
func Written(ds *zfs.Dataset, SnapshotName string) bool {
	var written bool
	changes, err := ds.Diff(SnapshotName)
	if err != nil {
		w.Err("[ERROR > lib/snapshot.go:245] it was not possible to search changes in dataset '"+ds.Name+"'.")
	} else {
		if len(changes) > 0 {
			if changes[0].Change == zfs.Modified {
				written = true
			}
		}
	}
	return written
}
