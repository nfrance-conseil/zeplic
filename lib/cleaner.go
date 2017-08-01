// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Cleaner removes all keys with the flag #deleted at the datacenter indicated
//
package lib

import (
	"fmt"
	"os"
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/hashicorp/consul/api"
)

// Cleaner deletes all KV pairs with #deleted flag in dataset
func Cleaner(RealDataset string) int {
	// Search the dataset
	var datacenter string
	j, _, _ := config.JSON()
	for i := 0; i < j; i++ {
		pieces := config.Extract(i)
		dataset := pieces[2].(string)
		if RealDataset == dataset {
			datacenter = pieces[4].(string)
			break
		} else {
			continue
		}
	}

	var code int
	if datacenter != "" {
		// Resolve hostname
		hostname, err := os.Hostname()
		if err != nil {
			w.Err("[ERROR > lib/cleaner.go:35] it was not possible to resolve the hostname.")
			code = 1
			return code
		}

		// Create a new client
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			w.Err("[ERROR > lib/cleaner.go:43]@[CONSUL] it was not possible to create a new client.")
			code = 1
			return code
		}
		kv := client.KV()

		// KV write options
		q1 := &api.QueryOptions{Datacenter: datacenter}

		// Get KV pairs
		keyfix := fmt.Sprintf("zeplic/%s/", hostname)
		pairs, _, err := kv.List(keyfix, q1)
		if err != nil {
			w.Err("[ERROR > lib/cleaner.go:56]@[CONSUL] it was not possible to get the KV pairs.")
			code = 1
			return code
		} else {
			for i := 0; i < len(pairs); i++ {
				value := string(pairs[i].Value[:])
				if strings.Contains(value, "#deleted") && strings.Contains(value, RealDataset) {
					key := pairs[i].Key
					pair := fmt.Sprintf("%s:%s", key, value)

					// Destroy KV pairs with #deleted flag
					q2 := &api.WriteOptions{Datacenter: datacenter}
					_, err = kv.Delete(key, q2)
					if err != nil {
						w.Err("[ERROR > lib/cleaner.go:71]@[CONSUL] it was not possible to destroy the KV pair '"+pair+"'.")
						code = 1
						break
					} else {
						w.Info("[INFO]@[CONSUL] the KV pair '"+pair+"' has been destroyed.")
						code = 0
						continue
					}
				} else {
					code = 0
					continue
				}
			}
			return code
		}
	} else {
		w.Err("[ERROR > lib/cleaner.go:24] the dataset '"+RealDataset+"' has not a datacenter configured.")
		code = 1
		return code
	}
}
