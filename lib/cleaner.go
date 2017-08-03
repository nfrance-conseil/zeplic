// Package lib contains: cleaner.go - commands.go - consul.go - destroy.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Cleaner removes all keys with the flag #deleted at the datacenter indicated
//
package lib

import (
	"fmt"
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
)

// Cleaner deletes all KV pairs with #deleted flag in dataset
func Cleaner(RealDataset string) int {
	// Search the dataset
	var datacenter string

	values := config.Local()
	for i := 0; i < len(values.Dataset); i++ {
		dataset := values.Dataset[i].Name
		if RealDataset == dataset {
			datacenter = values.Dataset[i].Consul.Datacenter
			break
		} else {
			continue
		}
	}

	var code int
	if datacenter != "" {
		// Get KV pairs
		keyfix := fmt.Sprintf("zeplic/%s/", Host())
		pairs := ListKV(keyfix, datacenter)
		for i := 0; i < len(pairs); i++ {
			value := string(pairs[i].Value[:])
			if strings.Contains(value, "#deleted") && strings.Contains(value, RealDataset) {
				// Destroy KV pairs with #deleted flag
				code := DeleteKV(pairs[i].Key, string(pairs[i].Value), datacenter)
				if code == 1 {
					break
				} else {
					continue
				}
			} else {
				code = 0
				continue
			}
		}
	} else {
		w.Err("[ERROR > lib/cleaner.go:22] the dataset '"+RealDataset+"' has not a datacenter configured.")
		code = 1
	}
	return code
}
