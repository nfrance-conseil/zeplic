// Package lib contains: cleaner.go - commands.go - consul.go - destroy.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Commands provides all ZFS functions to manage the datasets and backups
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
	// Check every dataset
	if index < 0 {
		code = 1
	} else {
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
				// Run functions
				// Delete an existing clone?
				if delClone == true {
					go DeleteClone(clone)
				}
				var snap *zfs.Dataset
				if SnapshotName != "" {
					prefix = SnapshotName
					snap = TakeSnapshot(prefix, NotWritten, ds, consul, datacenter)
				} else {
					snap = Snapshot(prefix, ds, consul, datacenter)
				}
				go DeleteBackup(dataset, ds)
				// Local retention policy?
				if director == false {
					go Policy(dataset, ds, retention, consul, datacenter)
				}
				// Create a backup snaphot?
				if getBackup == true {
					go Backup(dataset, ds)
				}
				// Clone the last snapshot?
				if getClone == true {
					go Clone(clone, snap)
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
			w.Err("[ERROR > lib/commands.go:94] it was not possible to create the dataset '"+dataset+"'.")
		} else {
			w.Info("[INFO] the dataset '"+dataset+"' has been created.")
		}
		ds, err := zfs.GetDataset(dataset)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:100] it was not possible to get the dataset '"+dataset+"'.")
		}
		return ds, err
	}
	return ds, err
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
			w.Err("[ERROR > lib/commands.go:117] it was not possible to destroy the clone '"+clone+"'.")
		} else {
			w.Info("[INFO] the clone '"+clone+"' has been destroyed.")
		}
	}
}

// Snapshot creates a new snapshot
func Snapshot(prefix string, ds *zfs.Dataset, consul bool, datacenter string) *zfs.Dataset {
	// Create a new snapshot
	snap, err := ds.Snapshot(SnapName(prefix), false)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:129] it was not possible to create a new snapshot.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
		// Assign an uuid to the snapshot
		err = UUID(snap)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:135] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
		}
		if consul == true {
			// KV write options
			snapUUID := SearchUUID(snap)
			key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
			value := snap.Name

			// Create a new KV
			code := PutKV(key, value, datacenter)
			if code == 1 {
				return nil
			}
		}
	}
	return snap
}

// TakeSnapshot creates a new snapshot
func TakeSnapshot(SnapshotName string, NotWritten bool, ds *zfs.Dataset, consul bool, datacenter string) *zfs.Dataset {
	var snap *zfs.Dataset
	var create bool

	// Something new was written?
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:161] it was not possible to access of snapshots list.")
	} else {
		count := len(list)
		_, amount := RealList(count, list, ds.Name)

		if amount == 0 {
			// Create a new snapshot
			snap, err = ds.Snapshot(SnapshotName, false)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:170] it was not possible to create a new snapshot.")
			} else {
				create = true
			}
		} else {
			snap, err = zfs.GetDataset(list[amount-1].Name)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:177] it was not possible to get the snapshot '"+snap.Name+"'.")
			} else {
				written := snap.Written
				if NotWritten == false || NotWritten == true && written > 0 {
					// Create a new snapshot
					snap, err = ds.Snapshot(SnapshotName, false)
					if err != nil {
						w.Err("[ERROR > lib/commands.go:184] it was not possible to create a new snapshot.")
					} else {
						create = true
					}
				}
			}
		}
	}

	// Check if it was created
	if create == true {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
		// Assign an uuid to the snapshot
		err = UUID(snap)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:199] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
		}
		if consul == true {
			// KV write options
			snapUUID := SearchUUID(snap)
			key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
			value := snap.Name

			// Create a new KV
			code := PutKV(key, value, datacenter)
			if code == 1 {
				return nil
			}
		}
	}
	return snap
}

// DeleteBackup deletes the backup snapshot
func DeleteBackup(dataset string, ds *zfs.Dataset) {
	// Search if the backup snapshot exists
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:222] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	} else {
		count := len(list)
		backup, amount := RealList(count, list, dataset)

		if backup == amount {
			w.Info("[INFO] the backup snapshot does not exist in dataset '"+dataset+"'.")
		} else {
			for i := 0; i < backup; i++ {
				take := list[i].Name
				if strings.Contains(take, "BACKUP") {
					backup, err := zfs.GetDataset(take)
					if err != nil {
						w.Err("[ERROR > lib/commands.go:235] it was not possible to get the backup snapshot '"+backup.Name+"'.")
					} else {
						err = backup.Destroy(zfs.DestroyDefault)
						if err != nil {
							w.Err("[ERROR > lib/commands.go:239] it was not possible to destroy the snapshot '"+backup.Name+"'.")
						} else {
							w.Info("[INFO] the snapshot '"+backup.Name+"' has been destroyed.")
						}
					}
					continue
				} else {
					continue
				}
			}
		}
	}
}

// Policy applies the retention policy
func Policy(dataset string, ds *zfs.Dataset, retention int, consul bool, datacenter string) {
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:257] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	} else {
		count := len(list)

		// Check the number of snapshot in the correct dataset
		_, amount := RealList(count, list, dataset)

		// Retention policy
		for k := 0; amount > retention; k++ {
			take := list[k].Name

			// Check the dataset
			dsName := DatasetName(take)
			if dsName == dataset {
				snap, err := zfs.GetDataset(take)
				uuid := SearchUUID(snap)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:274] it was not possible to get the snapshot '"+take+"'.")
				}
				err = snap.Destroy(zfs.DestroyDefault)
				if err != nil {
					cloned, clone := SnapCloned(snap)
					if cloned == true {
						w.Warning("[WARNING] the snapshot '"+take+"' has dependent clones: '"+clone+"'.")
					} else {
						w.Err("[ERROR > lib/commands.go:280] it was not possible to destroy the snapshot '"+take+"'.")
					}
				} else {
					w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
					amount--

					if consul == true {
						// KV write options
						key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), uuid)
						value := fmt.Sprintf("%s#%s", snap.Name, "deleted")

						// Edit KV pair
						go PutKV(key, value, datacenter)
					}
				}
				continue
			} else {
				continue
			}
		}
	}
}

// Backup creates a backup snapshot
func Backup(dataset string, ds *zfs.Dataset) {
	// Create the backup snapshot
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:310] it was not possible to access of snapshots list.")
	} else {
		count := len(list)
		_, amount := RealList(count, list, dataset)
			if amount > 0 {
			backup, err := ds.Snapshot(SnapBackup(dataset, ds), false)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:317] it was not possible to create the backup snapshot.")
			} else {
				w.Info("[INFO] the backup snapshot '"+backup.Name+"' has been created.")
				err = UUID(backup)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:322] it was not possible to assign an uuid to the backup snapshot '"+backup.Name+"'.")
				}
			}
		}
	}
}

// Clone creates a clone of last snapshot
func Clone(clone string, snap *zfs.Dataset) {
	_, err := snap.Clone(clone, nil)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:333] it was not possible to clone the snapshot '"+snap.Name+"'.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been clone.")
	}
}

// Rollback of last snapshot
func Rollback(SnapshotName string, snap *zfs.Dataset) {
	err := snap.Rollback(true)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:343] it was not possible to rolling back the snapshot '"+SnapshotName+"'.")
	} else {
		w.Info("[INFO] the snapshot '"+SnapshotName+"' has been restored.")
	}
}

// RealList returns the correct amount of snapshots in dataset
func RealList(count int, list []*zfs.Dataset, dataset string) (int, int) {
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
	var LastSnapshot string
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:385] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	} else {
		count := len(list)

		// Check the number of snapshot in the correct dataset
		_, amount := RealList(count, list, dataset)
		if amount == 0 {
			LastSnapshot = ""
		} else {
			LastSnapshot = list[amount-1].Name
		}
	}
	return LastSnapshot
}
