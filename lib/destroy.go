// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
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
func DestroyOrder(SnapshotUUID []string, Renamed bool, NotWritten bool, Cloned bool) int {
	var code int
	for i := 0 ; i < len(SnapshotUUID); i++ {
		uuid, name, _ := InfoKV(SnapshotUUID[i])

		// Define return variables
		RealSnapshotName := SearchName(uuid)
		RealDataset := DatasetName(RealSnapshotName)

		// Define interface
		var pieces []interface{}
		// Define index of pieces
		index := -1
		// Extract JSON information
		j, _, _ := config.JSON()

		for k := 0; k < j; k++ {
			pieces := config.Extract(i)
			dataset := pieces[2].(string)
			if dataset == RealDataset {
				index = i
				break
			} else {
				continue
			}
		}

		// Dataset configured
		if index > -1 {
			// Dataset enabled
			if pieces[0].(bool) == true {
				// Get the snapshot
				snap, err := zfs.GetDataset(name)
				if err != nil {
					w.Err("[ERROR > lib/destroy.go:49] it was not possible to get the snapshot '"+name+"'.")
					code = 1
					return code
				}

				// Check if the snapshot was renamed
				WasRenamed := WasRenamed(name, RealSnapshotName)

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
						w.Info("[INFO] the snapshot '"+name+"' was renamed to '"+RealSnapshotName+"'.")
						code = 0
					} else {
						if Cloned == true && WasCloned == true {
							if WasRenamed == true {
								w.Info("[INFO] the snapshot '"+name+"' (renamed as "+RealSnapshotName+") has dependent clones: '"+CloneName+"'.")
								code = 0
							} else {
								w.Info("[INFO] the snapshot '"+name+"' has dependent clones: '"+CloneName+"'.")
								code = 0
								return code
							}
						} else {
							err = snap.Destroy(zfs.DestroyRecursiveClones)
							if err != nil {
								w.Err("[ERROR > lib/destroy.go:85] it was not possible to destroy the snapshot '"+name+"'.")
								code = 1
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
								key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, uuid)
								q1 := &api.QueryOptions{Datacenter: pieces[4].(string)}

								// Get KV pair
								pair, _, err := kv.Keys(key, "", q1)
								if err != nil {
									w.Err("[ERROR > lib/destroy.go:112]@[CONSUL] it was not possible to get the KV pairs.")
								}

								_, _, flag := InfoKV(pair[0])
								var value string
								if flag != "" {
									value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
								} else {
									value = fmt.Sprintf("%s#%s", name, "deleted")
								}
								p := &api.KVPair{Key: key, Value: []byte(value)}
								q2 := &api.WriteOptions{Datacenter: pieces[4].(string)}

								// Edit KV pair
								_, err = kv.Put(p, q2)
								if err != nil {
									w.Err("[ERROR > lib/destroy.go:128]@[CONSUL] it was not possible to edit the KV pair.")
									code = 1
									return code
								}

								if Renamed == true {
									w.Info("[INFO] the snapshot '"+name+"' (renamed as "+RealSnapshotName+") has been destroyed.")
									code = 0
								} else {
									w.Info("[INFO] the snapshot '"+name+"' has been destroyed.")
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
	return code
}
