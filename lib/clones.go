// Package lib contains: clones.go - commands.go - snapshot.go - uuid.go
//
// Clones searchs the dataset where the snapshot was cloned
//
package lib

import (
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// SearchClone searchs the name of the dataset where a snapshot was cloned
func SearchClone(ds *zfs.Dataset) string {
	clone, err := ds.GetProperty("clones")
	if err != nil {
		w.Err("[ERROR] it was not possible to find the clone of '"+ds.Name+"'.")
	}
	return clone
}
