// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Functions to create a snapshot
//
package lib

import (
	"github.com/nfrance-conseil/zeplic/config"
)

// TakeOrder creates a new snapshot based on the order received from director
func TakeOrder(DestDataset string, SnapshotName string, NotWritten bool) int {
	// Define index of pieces
	index := -1

	// Define dataset variable
	var dataset string

	values := config.JSON()
	for i := 0; i < len(values.Dataset); i++ {
		dataset = values.Dataset[i].Name
		if dataset == DestDataset {
			index = i
			break
		} else {
			continue
		}
	}

	// Call Runner function
	var code int
	if index > -1 {
		code = Runner(index, true, SnapshotName, NotWritten)
	} else {
		w.Notice("[NOTICE] the dataset '"+DestDataset+"' is not configured.")
		code = 1
	}
	return code
}
