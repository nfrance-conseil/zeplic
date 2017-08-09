// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Tracker searchs in local datasets
//
package lib

import (
	"strconv"

	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// Status for response
const (
	WasRenamed = iota + 1 // The snapshot sent was renamed on destination
	WasWritten	      // The snapshot sent was written on destination
	NothingToDo	      // The snapshot sent already existed on destination
	Zerror		      // Any error string
	NotEmpty	      // Need an incremental stream
	Incremental	      // Ready to send an incremental stream
	MostActual	      // The last snapshot on destination is the most actual
)

// Delivery compares the local dataset with the list of UUIDs received
func Delivery(MapUUID []string, SnapshotName string) ([]byte, bool, bool, *zfs.Dataset, *zfs.Dataset) {
	var found string
	index := -1
	for i := len(MapUUID)-1; i > -1; i-- {
		name := SearchName(MapUUID[i])
		_, err := zfs.GetDataset(name)
		if err != nil {
			continue
		} else {
			found = name
			index = i
			break
		}
	}

	// Define all variables
	var send bool
	var incremental bool
	var ds1 *zfs.Dataset
	var ds2 *zfs.Dataset

	// Struct for the flag
	ack := make([]byte, 0)

	// Choose the correct option
	if found == "" {
		ack = nil
		ack = strconv.AppendInt(ack, Zerror, 10)
		send = true
		incremental = false
	} else if found == SnapshotName {
		ack = nil
		ack = strconv.AppendInt(ack, NothingToDo, 10)
		send = false
		incremental = false
	} else if found != "" && found != SnapshotName {
		dataset := DatasetName(SnapshotName)
		ds, err := zfs.GetDataset(dataset)
		if err != nil {
			w.Err("[ERROR > lib/tracker.go:62] it was not possible to get the dataset '"+dataset+"'.")
		} else {
			list, err := ds.Snapshots()
			if err != nil {
				w.Err("[ERROR > lib/tracker.go:66] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
			} else {
				// Search the index of snapshot to send
				var number int
				_, amount := RealList(ds, "")
				for i := 0; i < len(amount); i++ {
					if list[amount[i]].Name == SnapshotName {
						number = i
						break
					} else {
						continue
					}
				}

				if index < number {
					ds1, err = zfs.GetDataset(found)
					if err != nil {
						w.Err("[ERROR > lib/tracker.go:83] it was not possible to get the snapshot '"+found+"'.")
					}
					ds2, err = zfs.GetDataset(SnapshotName)
					if err != nil {
						w.Err("[ERROR > lib/tracker.go:87] it was not possible to get the snapshot '"+SnapshotName+"'.")
					}
					ack = nil
					ack = strconv.AppendInt(ack, Incremental, 10)
					send = true
					incremental = true
				} else {
					ack = nil
					ack = strconv.AppendInt(ack, MostActual, 10)
					send = false
					incremental = false
				}
			}
		}
	}
	return ack, send, incremental, ds1, ds2
}
