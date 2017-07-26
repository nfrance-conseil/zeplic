// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Cleaner removes all keys with the flag #deleted at the datacenter indicated
//
package lib

import (
	"github.com/hashicorp/consul/api"
)

// Cleaner deletes all KV pairs with #deleted flag in datacenter
func Cleaner(datacenter string) int {
	var code int
	// Create a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > lib/cleaner.go:15]@[CONSUL] it was not possible to create a new client.")
		code = 1
	}
	kv := client.KV()

	// KV write options
	q1 := &api.QueryOptions{Datacenter: datacenter}

	// Get KV pairs
	pairs, _, err := kv.Keys("#deleted", "", q1)
	if err != nil {
		w.Err("[ERROR > lib/cleaner.go:26]@[CONSUL] it was not possible to get the KV pairs.")
		code = 1
	}

	// Destroy KV pairs with #deleted flag
	q2 := &api.WriteOptions{Datacenter: datacenter}
	_, err = kv.DeleteTree("#deleted", q2)
	if err != nil {
		w.Err("[ERROR > lib/cleaner.go:34]@[CONSUL] it was not possible to destroy the #deleted pairs.")
		code = 1
	} else {
		// Inform to syslog
		for i := 0; i < len(pairs); i++ {
			w.Info("[INFO]@[CONSUL] the KV '"+pairs[i]+"' has been destroyed.")
		}
		code = 0
	}
	return code
}
