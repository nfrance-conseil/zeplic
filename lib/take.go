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
	// Extract JSON information
	j,_, _ := config.JSON()

	// Define dataset variable
	var dataset string

	for i := 0; i < j; i++ {
		pieces := config.Extract(i)
		dataset = pieces[2].(string)
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
		return code
	} else {
		w.Notice("[NOTICE] the dataset '"+DestDataset+"' is not configured.")
		code = 1
		return code
	}
	return code
}
