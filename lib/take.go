// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Functions to create a snapshot
//
package lib

import (
	"github.com/nfrance-conseil/zeplic/config"
)

// TakeOrder creates a new snapshot based on the order received from director
func TakeOrder(DestDataset string, SnapshotName string, NotWritten bool) {
	// Define index of pieces
	index := -1

	// Define dataset variable
	var dataset string

	values := config.Local()
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
	if index > -1 {
		go Runner(index, true, SnapshotName, NotWritten)
	} else {
		w.Notice("[NOTICE] the dataset '"+DestDataset+"' is not configured.")
	}
}
