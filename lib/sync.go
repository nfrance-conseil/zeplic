// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Sync writes a check KV to synchronize zeplic and resynchronize all pairs
//
package lib

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/utils"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
	"github.com/hashicorp/consul/api"
)

// Sync put a new check KV
func Sync(hostname string, datacenter string, dataset string, index int) {
	// Create a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > lib/sync.go:22]@[CONSUL] it was impossible to get a new client.")
	}
	// Get a handle to the KV API
	kv := client.KV()

	// PUT a new KV pair
	year, month, day := time.Now().Date()
	hour, _, _ := time.Now().Clock()
	key := fmt.Sprintf("zeplic/%s/syncKV%d", hostname, index)
	value := fmt.Sprintf("%s@zCHECK_%d-%s-%02d_%02d:00:00", dataset, year, month, day, hour)
	p := &api.KVPair{Key: key, Value: []byte(value)}
	q := &api.WriteOptions{Datacenter: datacenter}
	_, err = kv.Put(p, q)
	if err != nil {
		w.Err("[ERROR > lib/sync.go:35]@[CONSUL] it was impossible to put a new KV pair.")
	}
}

// Update updates the KV data in Consul
func Update(datacenter string, dataset string) {
	// Resolve hostname
	hostname, err := os.Hostname()
	if err != nil {
		w.Err("[ERROR > lib/sync.go:44] it was not possible to resolve the hostname.")
	}

	// Create a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > lib/sync.go:50]@[CONSUL] it was impossible to get a new client.")
	}
	// Get a handle to the KV API
	kv := client.KV()

	// Get all KV
	keyfix := fmt.Sprintf("zeplic/%s/", hostname)
	q := &api.QueryOptions{Datacenter: datacenter}
	pairs, _, err := kv.List(keyfix, q)
	if err != nil {
		w.Err("[ERROR > lib/sync.go:60]@[CONSUL] it was impossible to get the list of KV pairs.")
	}

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
		snapString := utils.After(PairsList[g], keyfix)
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
		w.Err("[ERROR > lib/sync.go:89] it was not possible to get the dataset '"+dataset+"'.")
	}
	list, err := ds.Snapshots()
	if err != nil {
		w.Err("[ERROR > lib/sync.go:93] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
	}
	count := len(list)
	_, amount := RealList(count, list, dataset)

	// Extract information of each snapshot
	for i := 0; i < amount; i++ {
		snap, err := zfs.GetDataset(list[i].Name)
		if err != nil {
			w.Err("[ERROR > lib/sync.go:102] it was not possible to get the snapshot '"+snap.Name+"'.")
		}
		snapUUID := SearchUUID(snap)
		// Create list of snapshots
		SnapshotsList = append(SnapshotsList, snapUUID)
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

		key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, snapUUID)
		value := snapName

		// Add the key and value of KV pair
		p := &api.KVPair{Key: key, Value: []byte(value)}
		q := &api.WriteOptions{Datacenter: datacenter}

		// Create a new KV
		_, err = kv.Put(p, q)
		if err != nil {
			w.Err("[ERROR > lib/sync.go:158]@[CONSUL] it was not possible to create a new KV.")
		}
	}
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
			key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, uuid)
			p := &api.KVPair{Key: key, Value: []byte(value)}
			q := &api.WriteOptions{Datacenter: datacenter}

			// Edit KV pair
			_, err = kv.Put(p, q)
			if err != nil {
				w.Err("[ERROR > lib/sync.go:187]@[CONSUL] it was not possible to edit the KV pair.")
			}
			destroy = false
		}
	}
	for p := 0; p < len(SourceList); p++ {
		partner := SourceList[p]
		jString := utils.Before(partner, ":")
		mString := utils.After(partner, ":")
		j, _ := strconv.Atoi(jString)
		m, _ := strconv.Atoi(mString)

		source := Source(SnapshotsList[m])
		if source == "received" {
			uuid, name, flag := InfoKV(PairsList[j])
			if !strings.Contains(flag, "#sent") {
				key := fmt.Sprintf("%s/%s/%s", "zeplic", hostname, uuid)
				var value string
				if flag != "" {
					value = fmt.Sprintf("%s#%s#%s", name, "sent", "deleted")
				} else {
					value = fmt.Sprintf("%s#%s", name, "sent")
				}

				p := &api.KVPair{Key: key, Value: []byte(value)}
				q := &api.WriteOptions{Datacenter: datacenter}

				// Edit KV pair
				_, err = kv.Put(p, q)
				if err != nil {
					w.Err("[ERROR > lib/sync.go:217]@[CONSUL] it was not possible to edit the KV pair.")
				}
			}
			continue
		} else {
			continue
		}
	}
}
