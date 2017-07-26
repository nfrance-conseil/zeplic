// Package lib contains: actions.go - cleaner.go - commands.go - destroy.go - snapshot.go - sync.go - take.go - uuid.go
//
// Functions to create a snapshot
//
package lib

import (
	"github.com/nfrance-conseil/zeplic/config"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
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

	var code int
	if index > -1 {
		// Skip if something new was written
		ds, err := zfs.GetDataset(dataset)
		if err != nil {
			w.Err("[ERROR > lib/take.go:36] it was not possible to get the dataset '"+dataset+"'.")
			code = 1
			return code
		}
		list, err := ds.Snapshots()
		if err != nil {
			w.Err("[ERROR > lib/take.go:42] it was not possible to access of snapshots list.")
			code = 1
			return code
		}
		count := len(list)
		_, amount := RealList(count, list, dataset)
		if amount == 0 {
			// Call to Runner function
			code = Runner(index, true, SnapshotName)
			return code
		} else if amount > 0 {
			snap, err := zfs.GetDataset(list[amount-1].Name)
			if err != nil {
				w.Err("[ERROR > lib/take.go:55] it was not possible to get the snapshots '"+snap.Name+"'.")
				code = 1
				return code
			}
			written := snap.Written

			if NotWritten == false || NotWritten == true && written > 0 {
				// Call to Runner function
				code = Runner(index, true, SnapshotName)
				return code
			}
		}
	} else {
		w.Notice("[NOTICE] the dataset '"+DestDataset+"' is not configured.")
		code = 1
		return code
	}
	return code
}
