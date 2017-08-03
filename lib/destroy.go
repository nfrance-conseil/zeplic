// Package lib contains: cleaner.go - commands.go - consul.go - destroy.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Functions to destroy a snapshot
//
package lib

import (
	"fmt"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/tools"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// DestroyOrder destroys a snapshot based on the order received from director
func DestroyOrder(SnapshotUUID []string, Renamed bool, NotWritten bool, Cloned bool) int {
	var code int
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
					w.Err("[ERROR > lib/destroy.go:44] it was not possible to get the snapshot '"+name+"'.")
					code = 1
				} else {
					// Check if the snapshot was renamed
					WasRenamed := SnapRenamed(name, RealSnapshotName)

					// Check if the snapshot was cloned
					WasCloned, CloneName := SnapCloned(snap)

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
								}
							} else {
								err = snap.Destroy(zfs.DestroyRecursiveClones)
								if err != nil {
									w.Err("[ERROR > lib/destroy.go:77] it was not possible to destroy the snapshot '"+name+"'.")
									code = 1
								} else {
									// KV write options
									keyfix := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), uuid)
									datacenter := values.Dataset[index].Consul.Datacenter

									// Get KV pairs
									pairs := ListKV(keyfix, datacenter)
									pair := fmt.Sprintf("%s:%s", pairs[0].Key, string(pairs[0].Value[:]))
									snapfix := fmt.Sprintf("%s/%s/", "zeplic", Host())
									pair = tools.After(pair, snapfix)

									_, _, flag := InfoKV(pair)
									var value string
									if flag != "" {
										value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
									} else {
										value = fmt.Sprintf("%s#%s", name, "deleted")
									}

									// Edit KV pair
									code := PutKV(keyfix, value, datacenter)
									if code == 1 {
										break
									} else {
										if Renamed == true {
											w.Info("[INFO] the snapshot '"+name+"' (renamed as "+RealSnapshotName+") has been destroyed.")
											code = 0
										} else {
											w.Info("[INFO] the snapshot '"+name+"' has been destroyed.")
											code = 0
										}
									}
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
	}
	return code
}
