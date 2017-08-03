// Package lib contains: cleaner.go - commands.go - consul.go - destroy.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Sync writes a check KV to synchronize zeplic and resynchronize all pairs
//
package lib

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/tools"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// Sync put a new check KV
func Sync(hostname string, datacenter string, dataset string, index int) {
	// Actual time
	year, month, day := time.Now().Date()
	hour, _, _ := time.Now().Clock()

	// PUT a new KV pair
	key := fmt.Sprintf("zeplic/%s/syncKV%d", hostname, index)
	value := fmt.Sprintf("%s@zCHECK_%d-%s-%02d_%02d:00:00", dataset, year, month, day, hour)
	go PutKV(key, value, datacenter)
}

// Update updates the KV data in Consul
func Update(datacenter string, dataset string) {
	// Get all KV
	keyfix := fmt.Sprintf("zeplic/%s/", Host())
	pairs := ListKV(keyfix, datacenter)

	// List of pairs
	var PairsList []string
	if len(pairs) > 0 {
		for f := 0; f < len(pairs); f++ {
			pair := fmt.Sprintf("%s:%s", pairs[f].Key, string(pairs[f].Value[:]))
			PairsList = append(PairsList, pair)
		}
	}
	var KVList []string
	for g := 0; g < len(PairsList); g++ {
		snapString := tools.After(PairsList[g], keyfix)
		KVList = append(KVList, snapString)
	}
	for h := 0; h < len(KVList); h++ {
		if !strings.Contains(KVList[h], dataset) {
			KVList = append(KVList[:h], KVList[h+1:]...)
			continue
		} else {
			continue
		}
	}

	// Extract all snapshots of dataset
	var SnapshotsList []string
	ds, err := zfs.GetDataset(dataset)
	if err != nil {
		w.Err("[ERROR > lib/sync.go:59] it was not possible to get the dataset '"+dataset+"'.")
	} else {
		list, err := ds.Snapshots()
		if err != nil {
			w.Err("[ERROR > lib/sync.go:63] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
		} else {
			count := len(list)
			_, amount := RealList(count, list, dataset)

			// Extract information of each snapshot
			for i := 0; i < amount; i++ {
				snap, err := zfs.GetDataset(list[i].Name)
				if err != nil {
					w.Err("[ERROR > lib/sync.go:72] it was not possible to get the snapshot '"+snap.Name+"'.")
				} else {
					snapUUID := SearchUUID(snap)
					// Create list of snapshots
					SnapshotsList = append(SnapshotsList, snapUUID)
				}
			}

			// Create indexing lists
			var found	 bool
			var CreateList	 []int
			var DeleteList	 []int
			var SourceList   []string
			for j := 0; j < len(PairsList); j++ {
				uuid, _, _ := InfoKV(PairsList[j])
				for m := 0; m < len(SnapshotsList); m++ {
					if strings.Contains(SnapshotsList[m], uuid) {
						index := fmt.Sprintf("%d:%d", j, m)
						SourceList = append(SourceList, index)
						break
					} else {
						if m == len(SnapshotsList)-1 {
							DeleteList = append(DeleteList, j)
						}
						continue
					}
				}
			}

			for k := 0; k < len(SnapshotsList); k++ {
				for m := 0; m < len(PairsList); m++ {
					if strings.Contains(PairsList[m], SnapshotsList[k]) {
						found = true
						break
					} else {
						continue
					}
				}
				if found == false {
					CreateList = append(CreateList, k)
				}
			}

			// Update KV pairs
			for m := 0; m < len(CreateList); m++ {
				snapUUID := SnapshotsList[CreateList[m]]
				snapName := SearchName(snapUUID)

				key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
				value := snapName

				// Create a new KV
				go PutKV(key, value, datacenter)

				for n := 0; n < len(DeleteList); n++ {
					pair := PairsList[DeleteList[n]]
					uuid, name, flag := InfoKV(pair)

					var destroy bool
					var value   string
					if flag != "" {
						if strings.Contains(flag, "#deleted") {
							continue
						} else {
							value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
							destroy = true
						}
					} else {
						value = fmt.Sprintf("%s#%s", name, "deleted")
						destroy = true
					}

					if destroy == true {
						key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), uuid)

						// Edit KV pair
						go PutKV(key, value, datacenter)
						destroy = false
					}
				}

				for p := 0; p < len(SourceList); p++ {
					partner := SourceList[p]
					jString := tools.Before(partner, ":")
					mString := tools.After(partner, ":")
					j, _ := strconv.Atoi(jString)
					m, _ := strconv.Atoi(mString)

					source := Source(SnapshotsList[m])
					if source == "received" {
						uuid, name, flag := InfoKV(PairsList[j])
						if !strings.Contains(flag, "#sent") {
							key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), uuid)
							var value string
							if flag != "" {
								value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
							} else {
								value = fmt.Sprintf("%s#%s", name, "sent")
							}

							// Edit KV pair
							go PutKV(key, value, datacenter)
						}
						continue
					} else {
						continue
					}
				}
			}
		}
	}
}
