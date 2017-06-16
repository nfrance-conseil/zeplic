// Package lib contains: comnands.go - snapshot.go - uuid.go
//
// Commands provides all ZFS functions to manage the datasets and backups
//
package lib

import (
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/mistifyio/go-zfs"
)

var (
	w, _ = config.LogBook()
)

// CommandOrder takes a snapshot based on the order received from director
func CommandOrder(j int, dataset string) {
	// Check if dataset is configured
	index := -1
	for i := 0; i < j; i++ {
		pieces := config.Extract(i)
		take := pieces[2].(string)

		if take == dataset {
			index = i
			break
		} else {
			continue
		}
	}

	if index > -1 {
		// Extract data of dataset
		pieces := config.Extract(index)
		enable := pieces[0].(bool)
		take := pieces[2].(string)
		snapshot := pieces[3].(string)
		retain := pieces[4].(int)

		if take == dataset && enable == true {
			ds := Dataset(take)
			Snapshot(take, snapshot, ds)
			DeleteBackup(take)
			Policy(take, retain)
			Backup(pieces[5].(bool), take, ds)
		} else if take == dataset && enable == false {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is disabled.")
		}
	} else {
		w.Notice("[NOTICE] the dataset '"+dataset+"' is not configured.")
	}
}

// RealMain is a loop that executes 'ZFS' functions for each dataset enabled
func RealMain(j int) int {
	for i := 0; i < j; i++ {
		// Extract all data stored in JSON file
		pieces := config.Extract(i)

		// This value returns if the dataset is enable
		takedataset := pieces[0].(bool)

		// Execute the functions
		if takedataset == true {
			clone := pieces[1].(string)
			dataset := pieces[2].(string)
			snapshot := pieces[3].(string)
			DeleteClone(clone)
			ds := Dataset(dataset)
			s, SnapshotName := Snapshot(dataset, snapshot, ds)
			DeleteBackup(dataset)
			Policy(dataset, pieces[4].(int))
			Backup(pieces[5].(bool), dataset, ds)
			Clone(pieces[6].(bool), clone, dataset, SnapshotName, s)
		} else if takedataset == false && pieces[2].(string) != "" {
			w.Notice("[NOTICE] the dataset '"+pieces[2].(string)+"' is disabled.")
		}
	}
	return 0
}

// DeleteClone deletes an existing clone
func DeleteClone(clone string) {
	// Get clones dataset
	cl, err := zfs.GetDataset(clone)
	if err != nil {
		w.Info("[INFO] the clone '"+clone+"' does not exist.")
	} else {
		// Destroy clones dataset
		err := cl.Destroy(zfs.DestroyRecursiveClones)
		if err != nil {
			w.Err("[ERROR] it was not possible to destroy the clone '"+clone+"'.")
		} else {
			w.Info("[INFO] the clone '"+clone+"' has been destroyed.")
		}
	}
}

// Dataset creates a dataset or get an existing one
func Dataset(dataset string) (*zfs.Dataset) {
	ds, err := zfs.GetDataset(dataset)

	// Destroy dataset (optional)
/*	err := ds.Destroy(zfs.DestroyRecursive)
	if err != nil {
		w.Err("[ERROR] it was not possible to destroy the dataset '"+dataset+"'.")
	} else {
		w.Info("[INFO] the dataset '"+dataset+"' has been destroyed.")
	}
	ds, err = zfs.GetDataset(dataset)*/

	if err != nil {
		w.Info("[INFO] the dataset '"+dataset+"' does not exist.")

		// Create dataset if it does not exist
		_, err := zfs.CreateFilesystem(dataset, nil)
		if err != nil {
			w.Err("[ERROR] it was not possible to create the dataset '"+dataset+"'.")
		} else {
			w.Info("[INFO] the dataset '"+dataset+"' has been created.")
		}
		ds, _ = zfs.GetDataset(dataset)
	}
	return ds
}

// Snapshot creates a new snapshot
func Snapshot(dataset, name string, ds *zfs.Dataset) (*zfs.Dataset, string) {
	SnapshotName := SnapName(name)
	s, err := ds.Snapshot(SnapshotName, false)

	// Get the snapshot created
	list, err := zfs.Snapshots(dataset)
	count := len(list)
	take := list[count-1].Name
	if strings.Contains(take, "BACKUP") {
		take = list[count-1].Name
	}
	// Check if it was created
	if err != nil {
		w.Err("[ERROR] it was not possible to create the snapshot '"+dataset+"@"+SnapshotName+"'.")
	} else {
		w.Info("[INFO] the snapshot '"+dataset+"@"+SnapshotName+"' has been created.")
		// Assign an uuid to the snapshot
		go UUID(take)
	}
	return s, SnapshotName
}

// DeleteBackup deletes the backup snapshot
func DeleteBackup(dataset string) {
	// Delete the backup snapshot
	list, err := zfs.Snapshots(dataset)
	if err != nil {
		w.Err("[ERROR] it was not possible to access of snapshots list.")
	}
	count := len(list)
	for k := 0; k < count; k++ {
		take := list[k].Name
		if strings.Contains(take, "BACKUP") {
			snap, err := zfs.GetDataset(take)
			if err != nil {
				w.Err("[ERROR] it was not possible to get the snapshot '"+take+"'.")
			}
			err = snap.Destroy(zfs.DestroyDefault)
			if err != nil {
				w.Err("[ERROR] it was not possible to destroy the snapshot '"+take+"'.")
			} else {
				w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
			}
		}
	}
}

// Policy apply the retention policy
func Policy(dataset string, retain int) {
	list, err := zfs.Snapshots(dataset)
	if err != nil {
		w.Err("[ERROR] it was not possible to access of snapshots list.")
	}
	count := len(list)
	for k := 0; count > retain; k++ {
		take := list[k].Name
		snap, err := zfs.GetDataset(take)
		if err != nil {
			w.Err("[ERROR] it was not possible to get the snapshot '"+take+"'.")
		}
		err = snap.Destroy(zfs.DestroyDefault)
		if err != nil {
			w.Err("[ERROR] it was not possible to destroy the snapshot '"+take+"'.")
		} else {
			w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
		}
		list, err = zfs.Snapshots(dataset)
		if err != nil {
			w.Err("[ERROR] it was not possible to access of snapshots list.")
		}
		count = len(list)
	}
}

// Backup creates a backup snapshot
func Backup(backup bool, dataset string, ds *zfs.Dataset) {
	if backup == true {
		_, err := ds.Snapshot(SnapBackup(dataset), false)

		// Get the snapshot created
		list, err := zfs.Snapshots(dataset)
		count := len(list)
		take := list[count-1].Name

		// Check if it was created
		if err != nil {
			w.Err("[ERROR] it was not possible to create the backup snapshot '"+dataset+"@"+SnapBackup(dataset)+"'.")
		} else {
			w.Info("[INFO] the backup snapshot '"+dataset+"@"+SnapBackup(dataset)+"' has been created.")
			// Assign an uuid to the snapshot
			go UUID(take)
		}
	}
}

// Clone creates a clone of last snapshot
func Clone(takeclone bool, clone string, dataset string, SnapshotName string, s *zfs.Dataset) {
	if takeclone == true {
		_, err := s.Clone(clone, nil)
		if err != nil {
			w.Err("[ERROR] it was not possible to clone the snapshot '"+dataset+"@"+SnapshotName+"'.")
		} else {
			w.Info("[INFO] the snapshot '"+dataset+"@"+SnapshotName+"' has been clone.")
		}
	}
}

// Rollback of last snapshot
func Rollback(rollback bool, dataset string, SnapshotName string, s *zfs.Dataset) {
	if rollback == true {
		err := s.Rollback(true)
		if err != nil {
			w.Err("[ERROR] it was not possible to rolling back the snapshot '"+SnapshotName+"'.")
		} else {
			w.Info("[INFO] the snapshot '"+dataset+"@"+SnapshotName+"' has been restored.")
		}
	}
}
