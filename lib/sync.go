// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go
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
	key := fmt.Sprintf("zeplic/%s/SyncKV%d", hostname, index)
	value := fmt.Sprintf("%s@zCHECK_%d-%s-%02d_%02d:00:00", dataset, year, month, day, hour)
	go PutKV(key, value, datacenter)
}

// Resync checks if it is time to execute Update() function
func Resync(timezone []string) bool {
	var resync bool
	loc, _ := time.LoadLocation("UTC")
	year, month, day := time.Now().Date()
	hour, minute, _ := time.Now().Clock()
	actual := time.Date(year, month, day, hour, minute, 00, 0, loc)

	// Check if timezone struct is correct
	if len(timezone) < 1 || len(timezone) > 2 {
		w.Err("[ERROR > lib/sync.go:38] the length of resync struct is not valid.")
	} else {
		var minor time.Time
		var major time.Time
		for i := 0; i < 1; i++ {
			minorH := tools.Before(timezone[i], ":")
			minorM := tools.After(timezone[i], ":")
			H, _ := strconv.Atoi(minorH)
			M, _ := strconv.Atoi(minorM)
			minor  = time.Date(year, month, day, H, M, 00, 0, loc)
			majorH := tools.Before(timezone[i+1], ":")
			majorM := tools.After(timezone[i+1], ":")
			H, _ = strconv.Atoi(majorH)
			M, _ = strconv.Atoi(majorM)
			major  = time.Date(year, month, day, H, M, 00, 0, loc)
		}
		diff0 := major.Sub(minor).Seconds()
		if diff0 < 0 {
			w.Err("[ERROR > lib/sync.go:56] the time zone to resynchronize must belong to the same day.")
		} else if diff0 > 0 && diff0 < 300 {
			w.Err("[ERROR > lib/sync.go:58] the time zone to resynchronize must be greater than 5 minutes.")
		} else {
			diff1 := actual.Sub(minor).Seconds()
			diff2 := major.Sub(actual).Seconds()
			if diff1 >= 0 && diff2 >= 0 {
				resync = true
			}
		}
	}
	return resync
}

// Update updates the KV pairs data in Consul
func Update(datacenter string, dataset string) {
	// Get all KV
	keyfix := fmt.Sprintf("zeplic/%s/", Host())
	pairs := ListKV(keyfix, datacenter)

	// List of pairs
	var PairsList []string
	if len(pairs) > 0 {
		for f := 0; f < len(pairs); f++ {
			if !strings.Contains(pairs[f].Key, "SyncKV") {
				pair := fmt.Sprintf("%s:%s", pairs[f].Key, string(pairs[f].Value[:]))
				PairsList = append(PairsList, pair)
			}
		}
	}
	var KVList []string
	for g := 0; g < len(PairsList); g++ {
		snapString := tools.After(PairsList[g], keyfix)
		KVList = append(KVList, snapString)
	}
	for h := 0; h < len(KVList); h++ {
		_, snap, flag := InfoKV(KVList[h])
		RealDataset := DatasetName(snap)
		if RealDataset != dataset {
			KVList = append(KVList[:h], KVList[h+1:]...)
			h--
			continue
		} else if strings.Contains(flag, "#NotWritten") {
			KVList = append(KVList[:h], KVList[h+1:]...)
			h--
			continue
		} else {
			continue
		}
	}

	// Extract all snapshots of dataset
	var SnapshotsList []string
	ds, err := zfs.GetDataset(dataset)
	if err != nil {
		w.Err("[ERROR > lib/sync.go:110] it was not possible to get the dataset '"+dataset+"'.")
	} else {
		list, err := ds.Snapshots()
		if err != nil {
			w.Err("[ERROR > lib/sync.go:114] it was not possible to access of snapshots list.")
		} else {
			// Extract information of each snapshot
			_, amount := RealList(ds, "")
			for i := 0; i < len(amount); i++ {
				snap, err := zfs.GetDataset(list[amount[i]].Name)
				if err != nil {
					w.Err("[ERROR > lib/sync.go:121] it was not possible to get the snapshot '"+list[amount[i]].Name+"'.")
				} else {
					// Remove backup snapshots
					prefix := Prefix(snap.Name)
					if prefix != "BACKUP" {
						snapUUID := SearchUUID(snap)
						// Create list of snapshots
						SnapshotsList = append(SnapshotsList, snapUUID)
					}
				}
			}
		}
	}

	// Create indexing lists
	var found	 bool
	var CreateList	 []int
	var DeleteList	 []int
	var SourceList   []string

	// Snapshots removed
	for j := 0; j < len(KVList); j++ {
		uuid, _, _ := InfoKV(KVList[j])
		for m := 0; m < len(SnapshotsList); m++ {
			found = false
			if strings.Contains(SnapshotsList[m], uuid) {
				index := fmt.Sprintf("%d:%d", j, m)
				SourceList = append(SourceList, index)
				found = true
				break
			} else {
				continue
			}
		}
		if found == false {
			DeleteList = append(DeleteList, j)
		}
	}

	// Snaphots created
	for k := 0; k < len(SnapshotsList); k++ {
		for m := 0; m < len(KVList); m++ {
			found = false
			uuid, _, _ := InfoKV(KVList[m])
			if strings.Contains(uuid, SnapshotsList[k]) {
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
go Up(KVList, SnapshotsList, SourceList, DeleteList, CreateList)
	// Update KV pairs
/*	for m := 0; m < len(CreateList); m++ {
		snapUUID := SnapshotsList[CreateList[m]]
		snapName := SearchName(snapUUID)

		key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), snapUUID)
		value := snapName

		// Create a new KV
		go PutKV(key, value, datacenter)
	}

	for n := 0; n < len(DeleteList); n++ {
		pair := KVList[DeleteList[n]]
		uuid, name, flag := InfoKV(pair)

		var destroy bool
		var value   string
		if flag != "" {
			if strings.Contains(flag, "#deleted") {
				continue
			} else {
				if strings.Contains(flag, "#sent") {
					value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
					destroy = true
				} else if strings.Contains(flag, "#sync") {
					value = fmt.Sprintf("%s#%s#%s", name, "sync", "deleted")
					destroy = true
				}
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
		jString := tools.Before(SourceList[p], ":")
		mString := tools.After(SourceList[p], ":")
		j, _ := strconv.Atoi(jString)
		m, _ := strconv.Atoi(mString)

		source := Source(SnapshotsList[m])
		if source == "received" {
			uuid, name, flag := InfoKV(KVList[j])
			if !strings.Contains(flag, "#sync") {
				key := fmt.Sprintf("%s/%s/%s", "zeplic", Host(), uuid)
				var value string
				if flag != "" {
					value = fmt.Sprintf("%s#%s#%s", name, "sync", "deleted")
				} else {
					value = fmt.Sprintf("%s#%s", name, "sync")
				}

				// Edit KV pair
				go PutKV(key, value, datacenter)
			}
			continue
		} else {
			continue
		}
	}*/
}

// Up *** test for Update() function ***
func Up(KVList []string, SnapshotsList []string, SourceList []string, DeleteList []int, CreateList []int) {
	fmt.Println("")
	fmt.Println("")
	fmt.Println("********************")
	fmt.Printf("KVList: %s\n", KVList)
	fmt.Printf("SnapshotsList: %s\n", SnapshotsList)
	fmt.Printf("SourceList: %s\n", SourceList)
	fmt.Printf("DeleteList: %d\n", DeleteList)
	fmt.Printf("CreateList: %d\n", CreateList)
}
