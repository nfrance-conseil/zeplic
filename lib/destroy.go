// Package lib contains: commands.go - destroy.go - !policy.go - snapshot.go - take.go - uuid.go
//
// Functions to destroy a snapshot
//
package lib

import (
	"fmt"
	"os"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
	"github.com/hashicorp/consul/api"
)

// DestroyOrder destroys a snapshot based on the order received from director
func DestroyOrder(SnapshotUUID string, SnapshotName string, Renamed bool, NotWritten bool, Cloned bool) int {
	// Define return variables
	RealSnapshotName := SearchName(SnapshotUUID)
	RealDataset := DatasetName(RealSnapshotName)

	// Define interface
	var pieces []interface{}
	// Define index of pieces
	index := -1
	// Extract JSON information
	j,_, _ := config.JSON()

	for i := 0; i < j; i++ {
		pieces := config.Extract(i)
		dataset := pieces[2].(string)
		if dataset == RealDataset {
			index = i
			break
		} else {
			continue
		}
	}

	var code int
	// Dataset configured
	if index > -1 {
		// Dataset enabled
		if pieces[0].(bool) == true {
			// Get the snapshot
			snap, err := zfs.GetDataset(SnapshotName)
			if err != nil {
				w.Err("[ERROR > lib/destroy.go:46] it was not possible to get the snapshot '"+SnapshotName+"'.")
				code = 1
				return code
			}

			// Check if the snapshot was renamed
			WasRenamed := WasRenamed(SnapshotName, RealSnapshotName)

			// Check if the snapshot was cloned
			WasCloned, CloneName := WasCloned(snap)

			// Something new was written
			var NothingWasWritten bool
			amount := snap.Written
			if amount == 0 {
				NothingWasWritten = true
			}

			if NotWritten == false || NotWritten == true && NothingWasWritten == false {
				// Was renamed
				if Renamed == true && WasRenamed == true {
					w.Info("[INFO] the snapshot '"+SnapshotName+"' was renamed to '"+RealSnapshotName+"'.")
					code = 0
					return code
				} else {
					if Cloned == true && WasCloned == true {
						if WasRenamed == true {
							w.Info("[INFO] the snapshot '"+SnapshotName+"' (renamed as "+RealSnapshotName+") has dependent clones: '"+CloneName+"'.")
							code = 0
							return code
						} else {
							w.Info("[INFO] the snapshot '"+SnapshotName+"' has dependent clones: '"+CloneName+"'.")
							code = 0
							return code
						}
					} else {
						err = snap.Destroy(zfs.DestroyRecursiveClones)
						if err != nil {
							w.Err("[ERROR > lib/destroy.go:84] it was not possible to destroy the snapshot '"+SnapshotName+"'.")
							code = 1
							return code
						} else {
							// Create a new client
							client, err := api.NewClient(api.DefaultConfig())
							if err != nil {
								w.Err("[ERROR > lib/destroy.go:91]@[CONSUL] it was not possible to create a new client.")
								code = 1
								return code
							}
							kv := client.KV()

							// Resolve hostname
							hostname, err := os.Hostname()
							if err != nil {
								w.Err("[ERROR > lib/destroy.go:100] it was not possible to resolve the hostname.")
								code = 1
								return code
							}

							// KV write options
							key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, SnapshotUUID)
							value := fmt.Sprintf("%s#%s", SnapshotName, "deleted")

							// Update the key and value of KV pair
							p := &api.KVPair{Key: key, Value: []byte(value)}
							q := &api.WriteOptions{Datacenter: pieces[4].(string)}

							// Edit KV pair
							_, err = kv.Put(p, q)
							if err != nil {
								w.Err("[ERROR > lib/commands.go:116]@[CONSUL] it was not possible to edit the KV pair.")
								code = 1
								return code
							}

							if Renamed == true {
								w.Info("[INFO] the snapshot '"+SnapshotName+"' (renamed as "+RealSnapshotName+") has been destroyed.")
								code = 0
								return code
							} else {
								w.Info("[INFO] the snapshot '"+SnapshotName+"' has been destroyed.")
								code = 0
								return code
							}
						}
					}
				}
			}
		} else {
			w.Notice("[NOTICE] the dataset '"+RealDataset+"' is disabled.")
			code = 0
		}
	} else {
		w.Notice("[NOTICE] the dataset '"+RealDataset+"' is not configured.")
		code = 1
	}
	return code
}
