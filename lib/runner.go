// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Runner provides all ZFS functions to manage the datasets and backups
//
package lib

import (
	"fmt"
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

var (
	w = config.LogBook()
)

// Runner is a loop that executes 'ZFS' functions for every dataset enabled
func Runner(index int, director bool, SnapshotName string, NotWritten bool) int {
	var code int
	var snap *zfs.Dataset

	// Extract all data stored in JSON file
	values := config.Local()

	// This value returns if the dataset is enable
	enable	    := values.Dataset[index].Enable
	docker	    := values.Dataset[index].Docker
	dataset	    := values.Dataset[index].Name
	consul	    := values.Dataset[index].Consul.Enable
	datacenter  := values.Dataset[index].Consul.Datacenter
	prefix	    := values.Dataset[index].Prefix
	retention   := values.Dataset[index].Retention
	getBackup   := values.Dataset[index].Backup
	getClone    := values.Dataset[index].Clone.Enable
	clone	    := values.Dataset[index].Clone.Name
	delClone    := values.Dataset[index].Clone.Delete

	// Execute functions
	if enable == true && docker == false {
		ds, err := Dataset(dataset)
		if err != nil {
			code = 1
		} else {
			// Create a snapshot
			if strings.Contains(SnapshotName, "@") {
				snap, err = zfs.GetDataset(SnapshotName)
				if err != nil {
					w.Err("[ERROR > lib/runner.go:48] it was not possible to get the snapshot '"+SnapshotName+"'.")
				}
			} else if SnapshotName != "" && !strings.Contains(SnapshotName, "@") {
				prefix = SnapshotName
				snap = TakeSnapshot(prefix, NotWritten, ds, consul, datacenter)
			} else {
				snap = Snapshot(prefix, ds, consul, datacenter)
			}

			// Run ZFS functions...
			if snap != nil {
				// Delete an existing clone?
				cl, err := zfs.GetDataset(clone)
				if delClone == true && err == nil {
					go DeleteClone(cl)
				}
				// Delete an existing backup snapshot?
				backup, _ := RealList(ds)
				if backup != -1 {
					go DeleteBackup(ds, backup)
				}
				// Local retention policy?
				if director == false {
					go Policy(ds, retention, consul, datacenter)
				}
				// Create a backup snaphot?
				if getBackup == true {
					go Backup(ds)
				}
				// Clone the last snapshot?
				if getClone == true {
					go Clone(clone, snap)
				}
			}
			code = 0
		}
	} else if enable == false && dataset != "" {
		w.Notice("[NOTICE] the dataset '"+dataset+"' is disabled.")
		code = 0
	} else if docker == true {
		w.Notice("[NOTICE] the dataset '"+dataset+"' is a docker dataset.")
		code = 0
	}
	return code
}

// Dataset creates a dataset or get an existing one
func Dataset(dataset string) (*zfs.Dataset, error) {
	// Get the dataset
	ds, err := zfs.GetDataset(dataset)
	if err != nil {
		w.Info("[INFO] the dataset '"+dataset+"' does not exist.")

		// Create dataset if it does not exist
		_, err := zfs.CreateFilesystem(dataset, nil)
		if err != nil {
			w.Err("[ERROR > lib/runner.go:104] it was not possible to create the dataset '"+dataset+"'.")
		} else {
			w.Info("[INFO] the dataset '"+dataset+"' has been created.")
			ds, err = zfs.GetDataset(dataset)
			if err != nil {
				w.Err("[ERROR > lib/runner.go:109] it was not possible to get the dataset '"+dataset+"'.")
			}
		}
	}
	return ds, err
}

// Snapshot creates a new snapshot
func Snapshot(prefix string, ds *zfs.Dataset, consul bool, datacenter string) *zfs.Dataset {
	// Create a new snapshot
	snap, err := ds.Snapshot(SnapName(prefix), false)
	if err != nil {
		w.Err("[ERROR > lib/runner.go:121] it was not possible to create a new snapshot.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
		// Assign an uuid to the snapshot
		err = UUID(snap)
		if err != nil {
			w.Err("[ERROR > lib/runner.go:127] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
		}

		// Consul KV put
		if consul == true {
			// KV write options
			snapUUID := SearchUUID(snap)
			key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
			value := snap.Name

			// Create a new KV
			go PutKV(key, value, datacenter)
		}
	}
	return snap
}

// TakeSnapshot creates a new snapshot
func TakeSnapshot(SnapshotName string, SkipIfNotWritten bool, ds *zfs.Dataset, consul bool, datacenter string) *zfs.Dataset {
	var snap *zfs.Dataset
	_, amount := RealList(ds)

	if len(amount) == 0 {
		// Create a new snapshot
		snap, err := ds.Snapshot(SnapshotName, false)
		if err != nil {
			w.Err("[ERROR > lib/runner.go:153] it was not possible to create a new snapshot.")
		} else {
			w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
			// Assign an uuid to the snapshot
			err = UUID(snap)
			if err != nil {
				w.Err("[ERROR > lib/runner.go:159] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
			} else {
				// Consul KV put
				if consul == true {
					// KV write options
					snapUUID := SearchUUID(snap)
					key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
					value := snap.Name

					// Create a new KV
					go PutKV(key, value, datacenter)
				}
			}
		}
	} else {
		prefix := Prefix(SnapshotName)
		LastSnapshot := LastSnapshot(ds, prefix)

		// Search changes in dataset
		written := Written(ds, LastSnapshot)
		if SkipIfNotWritten == false || SkipIfNotWritten == true && written == true {
			// Create a new snapshot
			snap, err := ds.Snapshot(SnapshotName, false)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:183] it was not possible to create a new snapshot.")
			} else {
				w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
				// Assign an uuid to the snapshot
				err = UUID(snap)
				if err != nil {
					w.Err("[ERROR > lib/runner.go:189] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
				} else {
					// Consul KV put
					if consul == true {
						// KV write options
						snapUUID := SearchUUID(snap)
						key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
						value := snap.Name

						// Create a new KV
						go PutKV(key, value, datacenter)
					}
				}
			}
		}
	}
	return snap
}

// DeleteBackup deletes the backup snapshot
func DeleteBackup(ds *zfs.Dataset, backup int) {
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/runner.go:212] it was not possible to access of snapshots list.")
	} else {
		// Get the backup snapshot
		bk, err := zfs.GetDataset(list[backup].Name)
		if err != nil {
			w.Err("[ERROR > lib/runner.go:217] it was not possible to get the backup snapshot '"+bk.Name+"'.")
		} else {
			bk.Destroy(zfs.DestroyDefault)
			if err != nil {
				w.Err("[ERROR > lib/runner.go:221] it was not possible to destroy the snapshot '"+bk.Name+"'.")
			} else {
				w.Info("[INFO] the snapshot '"+bk.Name+"' has been destroyed.")
			}
		}
	}
}

// DeleteClone deletes an existing clone
func DeleteClone(cl *zfs.Dataset) {
	// Destroy clones dataset
	err := cl.Destroy(zfs.DestroyRecursiveClones)
	if err != nil {
		w.Err("[ERROR > lib/runner.go:234] it was not possible to destroy the clone '"+cl.Name+"'.")
	} else {
		w.Info("[INFO] the clone '"+cl.Name+"' has been destroyed.")
	}
}

// Policy applies the retention policy
func Policy(ds *zfs.Dataset, retention int, consul bool, datacenter string) {
	var SnapshotUUID []string
	// Check the number of snapshot in the correct dataset
	_, amount := RealList(ds)

	// Retention policy
	if len(amount) > retention {
		list, err := ds.Snapshots()
		if err != nil {
			w.Err("[ERROR > lib/runner.go:250] it was not possible to access of snapshots list in dataset '"+ds.Name+"'.")
		} else {
			for i := len(amount)-1; i > retention-1; i-- {
				snap, err := zfs.GetDataset(list[i].Name)
				if err != nil {
					w.Err("[ERROR > lib/runner.go:255] it was not possible to get the snapshot '"+snap.Name+"'.")
				} else {
					uuid := SearchUUID(snap)
					pair := fmt.Sprintf("%s:%s", uuid, snap.Name)
					SnapshotUUID = append(SnapshotUUID, pair)
				}
			}

			// Destroy snapshot
			go DestroyOrder(SnapshotUUID, false, false)
		}
	}
}

// Backup creates a backup snapshot
func Backup(ds *zfs.Dataset) {
	_, amount := RealList(ds)
	if len(amount) > 0 {
		// Create the backup snapshot
		backup, err := ds.Snapshot(SnapBackup(ds), false)
		if err != nil {
			w.Err("[ERROR > lib/runner.go:276] it was not possible to create the backup snapshot.")
		} else {
			w.Info("[INFO] the backup snapshot '"+backup.Name+"' has been created.")
			// Assign an uuid to the backup snapshot
			err = UUID(backup)
			if err != nil {
				w.Err("[ERROR > lib/runner.go:282] it was not possible to assign an uuid to the backup snapshot '"+backup.Name+"'.")
			}
		}
	}
}

// Clone creates a clone of last snapshot
func Clone(clone string, snap *zfs.Dataset) {
	_, err := snap.Clone(clone, nil)
	if err != nil {
		w.Err("[ERROR > lib/runner.go:292] it was not possible to clone the snapshot '"+snap.Name+"'.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been clone.")
	}
}

// Rollback of last snapshot
func Rollback(snap *zfs.Dataset) {
	err := snap.Rollback(true)
	if err != nil {
		w.Err("[ERROR > lib/runner.go:302] it was not possible to rolling back the snapshot '"+snap.Name+"'.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been restored.")
	}
}
