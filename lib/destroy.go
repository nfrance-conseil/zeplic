// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Functions to destroy a snapshot
//
package lib

import (
	"fmt"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// DestroyOrder destroys a snapshot based on the order received from director
func DestroyOrder(SnapshotUUID []string, SkipIfRenamed bool, SkipIfCloned bool) {
	// Should I destroy the snapshot?
	for i := 0 ; i < len(SnapshotUUID); i++ {
		uuid, name, _ := InfoKV(SnapshotUUID[i])

		// Define return variables
		RealSnapshotName := SearchName(uuid)
		RealDataset := DatasetName(RealSnapshotName)

		// Define index of pieces
		index := -1

		values := config.Local()
		for k := 0; k < len(values.Dataset); k++ {
			dataset := values.Dataset[k].Name
			if dataset == RealDataset {
				index = k
				break
			} else {
				continue
			}
		}

		// Dataset configured
		if index > -1 {
			// Dataset enabled
			if values.Dataset[index].Enable == true {
				// Get the snapshot
				snap, err := zfs.GetDataset(name)
				if err != nil {
					w.Err("[ERROR > lib/destroy.go:43] it was not possible to get the snapshot '"+name+"'.")
				} else {
					// Check if the snapshot was renamed
					Renamed := SnapRenamed(name, RealSnapshotName)

					// Check if the snapshot was cloned
					Cloned, CloneName := SnapCloned(snap)

					// Was renamed
					if SkipIfRenamed == true && Renamed == true {
						w.Info("[INFO] the snapshot '"+name+"' was renamed to '"+RealSnapshotName+"'.")
					} else if SkipIfCloned == true && Cloned == true {
						if Renamed == true {
							w.Info("[INFO] the snapshot '"+name+"' (renamed as "+RealSnapshotName+") has dependent clones: '"+CloneName+"'.")
						} else {
							w.Info("[INFO] the snapshot '"+name+"' has dependent clones: '"+CloneName+"'.")
						}
					} else {
						err = snap.Destroy(zfs.DestroyRecursiveClones)
						if err != nil {
							w.Err("[ERROR > lib/destroy.go:63] it was not possible to destroy the snapshot '"+name+"'.")
						} else {
							// KV write options
							key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), uuid)
							datacenter := values.Dataset[index].Consul.Datacenter

							// Get KV pairs
							pairs := ListKV(key, datacenter)
							pair := fmt.Sprintf("%s:%s", uuid, string(pairs[0].Value[:]))
							_, _, flag := InfoKV(pair)
							var value string
							if flag != "" {
								value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
							} else {
								value = fmt.Sprintf("%s#%s", name, "deleted")
							}

							// Edit KV pair
							code := PutKV(key, value, datacenter)
							if code == 1 {
								break
							} else {
								if Renamed == true {
									w.Info("[INFO] the snapshot '"+name+"' (renamed as "+RealSnapshotName+") has been destroyed.")
								} else {
									w.Info("[INFO] the snapshot '"+name+"' has been destroyed.")
								}
							}
						}
					}
				}
			} else {
				w.Notice("[NOTICE] the dataset '"+RealDataset+"' is disabled.")
			}
		} else {
			w.Notice("[NOTICE] the dataset '"+RealDataset+"' is not configured.")
		}
	}
}
