// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
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
func Runner(index int, director bool, SnapshotName string, NotWritten bool) int {
	var code int
	// Check every dataset
	if index < 0 {
		code = 1
	} else {
		// Extract all data stored in JSON file
		values := config.JSON()

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
				// Resolve hostname
				hostname, err := os.Hostname()
				if err != nil {
					w.Err("[ERROR > lib/commands.go:51] it was not possible to resolve the hostname.")
				}

				// Run functions
				DeleteClone(delClone, clone)
				var snap *zfs.Dataset
				var snapName string
				if SnapshotName != "" {
					prefix = SnapshotName
					snap, snapName = TakeSnapshot(prefix, NotWritten, ds, consul, datacenter, hostname)
				} else {
					snap, snapName = Snapshot(prefix, ds, consul, datacenter, hostname)
				}
				DeleteBackup(dataset, ds)
				if director == false {
					Policy(dataset, ds, retention, consul, datacenter, hostname)
				}
				Backup(getBackup, dataset, ds)
				Clone(getClone, clone, snapName, snap)
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
			w.Err("[ERROR > lib/commands.go:96] it was not possible to create the dataset '"+dataset+"'.")
		} else {
			w.Info("[INFO] the dataset '"+dataset+"' has been created.")
		}
		ds, err := zfs.GetDataset(dataset)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:102] it was not possible to get the dataset '"+dataset+"'.")
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
				w.Err("[ERROR > lib/commands.go:120] it was not possible to destroy the clone '"+clone+"'.")
			} else {
				w.Info("[INFO] the clone '"+clone+"' has been destroyed.")
			}
		}
	}
}

// Snapshot creates a new snapshot
func Snapshot(prefix string, ds *zfs.Dataset, consul bool, datacenter string, hostname string) (*zfs.Dataset, string) {
	// Create a new snapshot
	snap, err := ds.Snapshot(SnapName(prefix), false)

	// Check if it was created
	if err != nil {
		w.Err("[ERROR > lib/commands.go:133] it was not possible to create a new snapshot.")
	} else {
		w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
		// Assign an uuid to the snapshot
		err = UUID(snap)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:141] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
		}
		if consul == true {
			// Create a new client
			client, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				w.Err("[ERROR > lib/commands.go:147]@[CONSUL] it was not possible to create a new client.")
			}
			kv := client.KV()

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
				w.Err("[ERROR > lib/commands.go:163]@[CONSUL] it was not possible to create a new KV.")
			}
		}
		return snap, snap.Name
	}
	return snap, ""
}

// TakeSnapshot creates a new snapshot
func TakeSnapshot(SnapshotName string, NotWritten bool, ds *zfs.Dataset, consul bool, datacenter string, hostname string) (*zfs.Dataset, string) {
	// Something new was written?
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:176] it was not possible to access of snapshots list.")
	}
	count := len(list)
	_, amount := RealList(count, list, ds.Name)

	var snap *zfs.Dataset
	var create bool
	if amount == 0 {
		// Create a new snapshot
		snap, err = ds.Snapshot(SnapshotName, false)
		create = true
	} else {
		snap, err = zfs.GetDataset(list[amount-1].Name)
		if err != nil {
			w.Err("[ERROR > lib/commands.go:190] it was not possible to get the snapshot '"+snap.Name+"'.")
			return nil, ""
		}
		written := snap.Written
		if NotWritten == false || NotWritten == true && written > 0 {
			// Create a new snapshot
			snap, err = ds.Snapshot(SnapshotName, false)
			create = true
		}
	}

	// Check if it was created
	if err != nil {
		w.Err("[ERROR > lib/commands.go:204] it was not possible to create a new snapshot.")
	} else {
		if create == true {
			w.Info("[INFO] the snapshot '"+snap.Name+"' has been created.")
			// Assign an uuid to the snapshot
			err = UUID(snap)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:210] it was not possible to assign an uuid to the snapshot '"+snap.Name+"'.")
			}
			if consul == true {
				// Create a new client
				client, err := api.NewClient(api.DefaultConfig())
				if err != nil {
					w.Err("[ERROR > lib/commands.go:216]@[CONSUL] it was not possible to create a new client.")
				}
				kv := client.KV()

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
					w.Err("[ERROR > lib/commands.go:232]@[CONSUL] it was not possible to create a new KV.")
				}
			}
			return snap, snap.Name
		}
	}
	return snap, ""
}

// DeleteBackup deletes the backup snapshot
func DeleteBackup(dataset string, ds *zfs.Dataset) {
	// Search if the backup snapshot exists
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:246] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
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
					w.Err("[ERROR > lib/commands.go:259] it was not possible to get the backup snapshot '"+backup.Name+"'.")
				} else {
					err = backup.Destroy(zfs.DestroyDefault)
					if err != nil {
						w.Err("[ERROR > lib/commands.go:263] it was not possible to destroy the snapshot '"+backup.Name+"'.")
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
func Policy(dataset string, ds *zfs.Dataset, retention int, consul bool, datacenter string, hostname string) {
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:280] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
	count := len(list)

	// Check the number of snapshot in the correct dataset
	_, amount := RealList(count, list, dataset)

	// Create a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > lib/commands.go:290]@[CONSUL] it was impossible to create a new client.")
	}
	kv := client.KV()

	// Retention policy
	for k := 0; amount > retention; k++ {
		take := list[k].Name

		// Check the dataset
		dsName := DatasetName(take)
		if dsName == dataset {
			snap, err := zfs.GetDataset(take)
			uuid := SearchUUID(snap)
			if err != nil {
				w.Err("[ERROR > lib/commands.go:304] it was not possible to get the snapshot '"+take+"'.")
			}
			err = snap.Destroy(zfs.DestroyDefault)
			if err != nil {
				cloned, clone := WasCloned(snap)
				if cloned == true {
					w.Warning("[WARNING] the snapshot '"+take+"' has dependent clones: '"+clone+"'.")
				} else {
					w.Err("[ERROR > lib/commands.go:308] it was not possible to destroy the snapshot '"+take+"'.")
				}
			} else {
				w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
				amount--

				if consul == true {
					// KV write options
					key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, uuid)
					value := fmt.Sprintf("%s#%s", snap.Name, "deleted")

					// Update the key and value of KV pair
					p := &api.KVPair{Key: key, Value: []byte(value)}
					q := &api.WriteOptions{Datacenter: datacenter}

					// Edit KV pair
					_, err = kv.Put(p, q)
					if err != nil {
						w.Err("[ERROR > lib/commands.go:330]@[CONSUL] it was not possible to edit the KV pair.")
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
			w.Err("[ERROR > lib/commands.go:347] it was not possible to access of snapshots list.")
		}
		count := len(list)
		_, amount := RealList(count, list, dataset)
		if amount > 0 {
			backup, err := ds.Snapshot(SnapBackup(dataset, ds), false)

			// Check if it was created
			if err != nil {
				w.Err("[ERROR > lib/commands.go:354] it was not possible to create the backup snapshot.")
			} else {
				w.Info("[INFO] the backup snapshot '"+backup.Name+"' has been created.")
				err = UUID(backup)
				if err != nil {
					w.Err("[ERROR > lib/commands.go:361] it was not possible to assign an uuid to the backup snapshot '"+backup.Name+"'.")
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
			w.Err("[ERROR > lib/commands.go:375] it was not possible to clone the snapshot '"+SnapshotName+"'.")
		} else {
			w.Info("[INFO] the snapshot '"+SnapshotName+"' has been clone.")
		}
	}
}

// Rollback of last snapshot
func Rollback(SnapshotName string, snap *zfs.Dataset) {
	err := snap.Rollback(true)
	if err != nil {
		w.Err("[ERROR > lib/commands.go:386] it was not possible to rolling back the snapshot '"+SnapshotName+"'.")
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
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/commands.go:427] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
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
