// Package lib contains: commands.go - destroy.go - !policy.go - snapshot.go - take.go - uuid.go
//
// Commands provides all ZFS functions to manage the datasets and backups
//
package lib

import (
	"fmt"
	"os"
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
	"github.com/hashicorp/consul/api"
)

var (
	w = config.LogBook()
)

// Runner is a loop that executes 'ZFS' functions for every dataset enabled
func Runner(index int, director bool) int {
	var code int
	// Check every dataset
	if index < 0 {
		code = 1
		return code
	} else {
		// Extract all data stored in JSON file
		pieces := config.Extract(index)

		// This value returns if the dataset is enable
		enable := pieces[0].(bool)
		docker := pieces[1].(bool)

		// Get variables
		dataset	    := pieces[2].(string)
		consul	    := pieces[3].(bool)
		datacenter  := pieces[4].(string)
		snapshot    := pieces[5].(string)
		retention   := pieces[6].(int)
		getBackup   := pieces[7].(bool)
		getClone    := pieces[8].(bool)
		clone	    := pieces[9].(string)
		delClone    := pieces[10].(bool)

		// Execute functions
		if enable == true && docker == false {
			ds, err := Dataset(dataset)
			if err != nil {
				code = 1
				return code
			} else {
				DeleteClone(delClone, clone)
				snap, SnapshotName := Snapshot(dataset, snapshot, ds, consul, datacenter)
				DeleteBackup(dataset, ds)
				if director == false {
					Policy(dataset, ds, retention, consul, datacenter)
				}
				Backup(getBackup, dataset, ds)
				Clone(getClone, clone, SnapshotName, snap)
				code = 0
				return code
			}
		} else if enable == false && dataset != "" {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is disabled.")
			code = 0
			return code
		} else if docker == true {
			w.Notice("[NOTICE] the dataset '"+dataset+"' is a docker dataset.")
			code = 0
			return code
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
			w.Err("[ERROR > lib/commands.go:86] it was not possible to create the dataset '"+dataset+"'.")
		} else {
			w.Info("[INFO] the dataset '"+dataset+"' has been created.")
		}
		ds, err := zfs.GetDataset(dataset)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:92] it was not possible to get the dataset '"+dataset+"'.")
		}
		return ds, err
	}
	return ds, err
}

// DeleteClone deletes an existing clone
func DeleteClone(delClone bool, clone string) {
	if delClone == true {
		// Get clones dataset
		cl, err := zfs.GetDataset(clone)
		if err != nil {
			w.Info("[INFO] the clone '"+clone+"' does not exist.")
		} else {
			// Destroy clones dataset
			err := cl.Destroy(zfs.DestroyRecursiveClones)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:110] it was not possible to destroy the clone '"+clone+"'.")
			} else {
				w.Info("[INFO] the clone '"+clone+"' has been destroyed.")
			}
		}
	}
}

// Snapshot creates a new snapshot
func Snapshot(dataset string, name string, ds *zfs.Dataset, consul bool, datacenter string) (*zfs.Dataset, string) {
	// Create a new snapshot
	snap, err := ds.Snapshot(SnapName(name), false)

	// Check if it was created
	if err != nil {
		w.Err("[ERROR > lib/commands.go:123] it was not possible to create a new snapshot.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
		// Assign an uuid to the snapshot
		err = UUID(snap)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:131] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
		}
		if consul == true {
			// Create a new client
			client, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				w.Err("[ERROR > lib/commands.go:137]@[CONSUL] it was not possible to create a new client.")
			}
			kv := client.KV()

			// Resolve hostname
			hostname, err := os.Hostname()
			if err != nil {
				w.Err("[ERROR > lib/commands.go:144] it was not possible to resolve the hostname.")
			}

			// KV write options
			snapUUID := SearchUUID(snap)
			key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, snapUUID)
			value := snap.Name

			// Add the key and value of KV pair
			p := &api.KVPair{Key: key, Value: []byte(value)}
			q := &api.WriteOptions{Datacenter: datacenter}

			// Create a new KV
			_, err = kv.Put(p, q)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:159]@[CONSUL] it was not possible to create a new KV.")
			}
		}
		return snap, snap.Name
	}
	return snap, ""
}

// DeleteBackup deletes the backup snapshot
func DeleteBackup(dataset string, ds *zfs.Dataset) {
	// Search if the backup snapshot exists
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:172] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
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
					w.Err("[ERROR > lib/commands.go:185] it was not possible to get the backup snapshot '"+backup.Name+"'.")
				} else {
					err = backup.Destroy(zfs.DestroyDefault)
					if err != nil {
						w.Err("[ERROR > lib/commands.go:189] it was not possible to destroy the snapshot '"+backup.Name+"'.")
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

// Policy applies the retention policy
func Policy(dataset string, ds *zfs.Dataset, retention int, consul bool, datacenter string) {
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:206] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
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
				w.Err("[ERROR > lib/commands.go:222] it was not possible to get the snapshot '"+take+"'.")
			}
			err = snap.Destroy(zfs.DestroyDefault)
			if err != nil {
				cloned, clone := WasCloned(snap)
				if cloned == true {
					w.Warning("[WARNING] the snapshot '"+take+"' has dependent clones: '"+clone+"'.")
				} else {
					w.Err("[ERROR > lib/commands.go:227] it was not possible to destroy the snapshot '"+take+"'.")
				}
			} else {
				w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
				amount--

				if consul == true {
					// Create a new client
					client, err := api.NewClient(api.DefaultConfig())
					if err != nil {
						w.Err("[ERROR > lib/commands.go:241]@[CONSUL] it was not possible to create a new client.")

					}
					kv := client.KV()

					// Resolve hostname
					hostname, err := os.Hostname()
					if err != nil {
						w.Err("[ERROR > lib/commands.go:249] it was not possible to resolve the hostname.")
					}

					// KV write options
					key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, uuid)
					value := fmt.Sprintf("%s#%s", snap.Name, "deleted")

					// Update the key and value of KV pair
					p := &api.KVPair{Key: key, Value: []byte(value)}
					q := &api.WriteOptions{Datacenter: datacenter}

					// Edit KV pair
					_, err = kv.Put(p, q)
					if err != nil {
						w.Err("[ERROR > lib/commands.go:263]@[CONSUL] it was not possible to edit the KV pair.")
					}
				}
				continue
			}
		} else {
			continue
		}
	}
}

// Backup creates a backup snapshot
func Backup(getBackup bool, dataset string, ds *zfs.Dataset) {
	if getBackup == true {
		// Create the backup snapshot
		list, err := ds.Snapshots()
		if err != nil {
			w.Err("[ERROR > lib/commands.go:280] it was not possible to access of snapshots list.")
		}
		count := len(list)
		_, amount := RealList(count, list, dataset)
		if amount > 0 {
			backup, err := ds.Snapshot(SnapBackup(dataset, ds), false)

			// Check if it was created
			if err != nil {
				w.Err("[ERROR > lib/commands.go:287] it was not possible to create the backup snapshot.")
			} else {
				w.Info("[INFO] the backup snapshot '"+backup.Name+"' has been created.")
				err = UUID(backup)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:294] it was not possible to assign an uuid to the backup snapshot '"+backup.Name+"'.")
				}
			}
		} else {
			w.Info("[INFO] there is no snapshot to backup.")
		}
	}
}

// Clone creates a clone of last snapshot
func Clone(getClone bool, clone string, SnapshotName string, snap *zfs.Dataset) {
	if getClone == true && SnapshotName != "" {
		_, err := snap.Clone(clone, nil)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:308] it was not possible to clone the snapshot '"+SnapshotName+"'.")
		} else {
			w.Info("[INFO] the snapshot '"+SnapshotName+"' has been clone.")
		}
	}
}

// Rollback of last snapshot
func Rollback(SnapshotName string, snap *zfs.Dataset) {
	err := snap.Rollback(true)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:319] it was not possible to rolling back the snapshot '"+SnapshotName+"'.")
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
		w.Err("[ERROR > lib/commands.go:360] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
	count := len(list)

	// Check the number of snapshot in the correct dataset
	_, amount := RealList(count, list, dataset)

	var LastSnapshot string
	if amount == 0 {
		LastSnapshot = ""
	} else {
		LastSnapshot = list[amount-1].Name
	}
	return LastSnapshot
}
