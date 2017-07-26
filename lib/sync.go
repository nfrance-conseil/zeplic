// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Sync writes a check KV to synchronize zeplic
//
package lib

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// Sync put a new check KV
func Sync(hostname string, datacenter string, dataset string) {
	// Create a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > lib/sync.go:17]@[CONSUL] it was impossible to get a new client.")
	}
	// Get a handle to the KV API
	kv := client.KV()

	// PUT a new KV pair
	year, month, day := time.Now().Date()
	key := fmt.Sprintf("zeplic/%s/KV-to-sync", hostname)
	value := fmt.Sprintf("%s@zCHECK_%d_%s_%02d_00:00:00", dataset, year, month, day)
	p := &api.KVPair{Key: key, Value: []byte(value)}
	q := &api.WriteOptions{Datacenter: datacenter}
	_, err = kv.Put(p, q)
	if err != nil {
		w.Err("[ERROR > lib/sync.go:30]@[CONSUL] it was impossible to put a new KV pair.")
	}
}
