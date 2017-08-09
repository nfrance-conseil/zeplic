// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
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
	var dataset    string

	values := config.Local()
	for i := 0; i < len(values.Dataset); i++ {
		dataset := values.Dataset[i].Name
		if RealDataset == dataset {
			datacenter = values.Dataset[i].Consul.Datacenter
			dataset = RealDataset
			break
		} else {
			continue
		}
	}

	var code int
	if dataset != "" {
		if datacenter != "" {
			// Get KV pairs
			keyfix := fmt.Sprintf("zeplic/%s/", Host())
			pairs := ListKV(keyfix, datacenter)
			for i := 0; i < len(pairs); i++ {
				value := string(pairs[i].Value[:])
				if strings.Contains(value, RealDataset) && (strings.Contains(value, "#deleted") || strings.Contains(value, "#NotWritten")) {
					// Destroy KV pairs with #deleted flag
					code = DeleteKV(pairs[i].Key, value, datacenter)
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
			w.Err("[ERROR > lib/cleaner.go:23] the dataset '"+RealDataset+"' has not a datacenter configured.")
			code = 1
		}
	} else {
		w.Notice("[NOTICE] the dataset '"+RealDataset+"' is not configured.")
		code = 1
	}
	return code
}
