// Package lib contains: clones.go - commands.go - snapshot.go - uuid.go
//
// Commands provides all ZFS functions to manage the datasets and backups
//
package lib

import (
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

var (
	w = config.LogBook()
)

// TakeOrder takes a snapshot based on the order received from director
func TakeOrder(j int, DestDataset string, NotWritten bool) {
	// Check if dataset is configured
	index := -1
	for i := 0; i < j; i++ {
		pieces	:= config.Extract(i)
		dataset := pieces[4].(string)

		if dataset == DestDataset {
			index = i
			break
		} else {
			continue
		}
	}

	if index > -1 {
		// Extract data of dataset
		pieces	  := config.Extract(index)
		enable	  := pieces[0].(bool)
		docker	  := pieces[1].(bool)
		delClone  := pieces[2].(bool)
		clone	  := pieces[3].(string)
		dataset	  := pieces[4].(string)
		snapshot  := pieces[5].(string)
		retain	  := pieces[6].(int)
		getBackup := pieces[7].(bool)
		getClone  := pieces[8].(bool)

		if dataset == DestDataset && enable == true && docker == false {
			ds, err := Dataset(dataset)
			if err == nil {
				// Check the last snapshot in dataset
				SnapshotName := LastSnapshot(ds, dataset)
				snap, err := zfs.GetDataset(SnapshotName)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:52] it was not possible to get the snapshot '"+SnapshotName+"'.")
				}
				written := snap.Written

				// Create a new snapshot if something was written
				if NotWritten == false || NotWritten == true && written > 0 {
					DeleteClone(delClone, clone)
					s, SnapshotName := Snapshot(dataset, snapshot, ds)
					DeleteBackup(dataset, ds)
					Policy(dataset, ds, retain)
					Backup(getBackup, dataset, ds)
					Clone(getClone, clone, SnapshotName, s)
				}
			}
		} else if dataset == DestDataset && enable == false {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is disabled.")
		} else if dataset == DestDataset && docker == true {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is a docker dataset.")
		}
	} else {
		w.Notice("[NOTICE] the dataset '"+DestDataset+"' is not configured.")
	}
}

// DestroyOrder destroys a snapshot based on the order received from director
func DestroyOrder(SnapshotName string, Cloned bool) (bool, string) {
	// Define return variables
	var destroy bool

	// Get the snapshot
	snap, err := zfs.GetDataset(SnapshotName)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:84] it was not possible to get the snapshot '"+SnapshotName+"'.")
	}

	// Check if the snapshot was cloned
	clone := SearchClone(snap)

	if Cloned == false {
		err = snap.Destroy(zfs.DestroyRecursiveClones)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:93] it was not possible to destroy the snapshot '"+SnapshotName+"'.")
		} else {
			destroy = true
		}
	} else {
		err = snap.Destroy(zfs.DestroyDefault)
		if err != nil {
			destroy = false
		} else {
			destroy = true
		}
	}
	return destroy, clone
}

// Runner is a loop that executes 'ZFS' functions for each dataset enabled
func Runner(j int) int {
	for i := 0; i < j; i++ {
		// Extract all data stored in JSON file
		pieces := config.Extract(i)

		// This value returns if the dataset is enable
		enable := pieces[0].(bool)
		docker := pieces[1].(bool)

		// Get variables
		delClone  := pieces[2].(bool)
		clone	  := pieces[3].(string)
		dataset	  := pieces[4].(string)
		snapshot  := pieces[5].(string)
		retain	  := pieces[6].(int)
		getBackup := pieces[7].(bool)
		getClone  := pieces[8].(bool)

		// Execute functions
		if enable == true && docker == false {
			ds, err := Dataset(dataset)
			if err != nil {
				continue
			} else {
				DeleteClone(delClone, clone)
				s, SnapshotName := Snapshot(dataset, snapshot, ds)
				DeleteBackup(dataset, ds)
				Policy(dataset, ds, retain)
				Backup(getBackup, dataset, ds)
				Clone(getClone, clone, SnapshotName, s)
				continue
			}
		} else if enable == false && dataset != "" {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is disabled.")
			continue
		} else if docker == true {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is a docker dataset.")
		} else {
			continue
		}
	}
	return 0
}

// DeleteClone deletes an existing clone
func DeleteClone(delClone bool, clone string) {
	// Get clones dataset
	cl, err := zfs.GetDataset(clone)
	if err != nil {
		w.Info("[INFO] the clone '"+clone+"' does not exist.")
	} else {
		// Destroy clones dataset
		if delClone == true {
			err := cl.Destroy(zfs.DestroyRecursiveClones)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:164] it was not possible to destroy the clone '"+clone+"'.")
			} else {
				w.Info("[INFO] the clone '"+clone+"' has been destroyed.")
			}
		}
	}
}

// Dataset creates a dataset or get an existing one
func Dataset(dataset string) (*zfs.Dataset, error) {
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
			w.Err("[ERROR > lib/commands.go:191] it was not possible to create the dataset '"+dataset+"'.")
		} else {
			w.Info("[INFO] the dataset '"+dataset+"' has been created.")
		}
		ds, err = zfs.GetDataset(dataset)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:197] it was not possible to get the dataset '"+dataset+"'.")
		}
	}
	return ds, err
}

// Snapshot creates a new snapshot
func Snapshot(dataset string, name string, ds *zfs.Dataset) (*zfs.Dataset, string) {
	SnapshotName := SnapName(name)
	s, err := ds.Snapshot(SnapshotName, false)

	// Check if it was created
	if err != nil {
		w.Err("[ERROR > lib/commands.go:208] it was not possible to create the snapshot '"+dataset+"@"+SnapshotName+"'.")
	} else {
		// Get the snapshot created
		SnapshotName = s.Name
		w.Info("[INFO] the snapshot '"+SnapshotName+"' has been created.")
		// Assign an uuid to the snapshot
		snap, err := zfs.GetDataset(SnapshotName)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:218] it was not possible to get the snapshot '"+snap.Name+"'.")
		}
		err = UUID(snap)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:222] it was not possible to assign an uuid to the snapshot '"+SnapshotName+"'.")
		}
	}
	return s, SnapshotName
}

// DeleteBackup deletes the backup snapshot
func DeleteBackup(dataset string, ds *zfs.Dataset) {
	// Get the backup snapshot
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:233] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
	count := len(list)
	for i := 0; i < count; i++ {
		take := list[i].Name
		dsName := DatasetName(take)
		if dsName == dataset && strings.Contains(take, "BACKUP") {
			snap, err := zfs.GetDataset(take)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:242] it was not possible to get the snapshot '"+take+"'.")
			}
			err = snap.Destroy(zfs.DestroyDefault)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:246] it was not possible to destroy the snapshot '"+take+"'.")
			} else {
				w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
			}
		} else {
			continue
		}
	}
}

// Policy applies the retention policy
func Policy(dataset string, ds *zfs.Dataset, retain int) {
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:260] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
	count := len(list)

	// Check the number of snapshot in the correct dataset
	_, amount := RealList(count, list, dataset)

	for k := 0; amount > retain; k++ {
		take := list[k].Name
		// Check the dataset
		dsName := DatasetName(take)
		if dsName == dataset {
			snap, err := zfs.GetDataset(take)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:274] it was not possible to get the snapshot '"+take+"'.")
			}
			err = snap.Destroy(zfs.DestroyDefault)
			if err != nil {
				clone := SearchClone(snap)
				w.Warning("[WARNING] the snapshot '"+take+"' has dependent clones: '"+clone+"'.")
			} else {
				w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
				amount--
				continue
			}
		} else {
			continue
		}
	}
}

// Backup creates a backup snapshot
func Backup(backup bool, dataset string, ds *zfs.Dataset) {
	if backup == true {
		_, err := ds.Snapshot(SnapBackup(dataset, ds), false)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:296] it was not possible to create the backup snapshot '"+dataset+"@"+SnapBackup(dataset, ds)+"'.")
		}

		// Get the backup snapshot created
		list, err := ds.Snapshots()
		if err != nil {
			w.Err("[ERROR > lib/commands.go:302] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
		}
		count := len(list)
		for i := 0; i < count; i++ {
			take := list[i].Name
			dsName := DatasetName(take)
			if dsName == dataset && strings.Contains(take, "BACKUP") {
				w.Info("[INFO] the backup snapshot '"+take+"' has been created.")
				// Assign an uuid to the snapshot
				snap, err := zfs.GetDataset(take)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:313] it was not possible to get the snapshot '"+snap.Name+"'.")
				}
				err = UUID(snap)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:317] it was not possible to assign an uuid to the snapshot '"+take+"'.")
				}
				break
			} else {
				continue
			}
		}
	}
}

// Clone creates a clone of last snapshot
func Clone(takeclone bool, clone string, SnapshotName string, s *zfs.Dataset) {
	if takeclone == true {
		_, err := s.Clone(clone, nil)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:332] it was not possible to clone the snapshot '"+SnapshotName+"'.")
		} else {
			w.Info("[INFO] the snapshot '"+SnapshotName+"' has been clone.")
		}
	}
}

// Rollback of last snapshot
func Rollback(SnapshotName string, s *zfs.Dataset) {
	err := s.Rollback(true)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:343] it was not possible to rolling back the snapshot '"+SnapshotName+"'.")
	} else {
		w.Info("[INFO] the snapshot '"+SnapshotName+"' has been restored.")
	}
}

// RealList returns the correct amount of snapshots in dataset
func RealList (count int, list []*zfs.Dataset, dataset string) (int, int) {
	amount := count

	// Check the number of snapshot in the correct dataset
	for i := count-1; i > -1; i-- {
		// Check the dataset
		take := list[i].Name
		dsName := DatasetName(take)
		if dsName != dataset {
			amount--
			continue
		} else {
			continue
		}
	}
	backup := amount

	// Search if exist the backup snapshot
	for j := 0; j < amount; j++ {
		take := list[j].Name
		if strings.Contains(take, "BACKUP") {
			amount--
			continue
		} else {
			continue
		}
	}

	return backup, amount
}

// LastSnapshot returns the name of last snapshot in 'dataset'
func LastSnapshot(ds *zfs.Dataset, dataset string) string {
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:385] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
	count := len(list)

	// Check the number of snapshot in the correct dataset
	_, amount := RealList(count, list, dataset)

	LastSnapshot := list[amount-1].Name
	return LastSnapshot
}
