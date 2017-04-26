package main

import (
	"fmt"
	"time"

	"github.com/mistifyio/go-zfs"
)

func ok(err error) {
	if err != nil {
		fmt.Printf("\033[31mUnexpected error! %s\033[39m\n", err.Error())
	}
}

func snapName() string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", "SNAP", year, month, day, hour, min, sec)
	return snapDate
}

func main () {
	// Get Clones
	c, err := zfs.GetDataset("tank/clones")
	ok(err)
	if c != nil {
		// Destroy Clones
		c.Destroy(zfs.DestroyRecursiveClones)
		ok(err)
	}

	// Get Dataset
	f, err := zfs.GetDataset("tank/test")
	ok(err)
	if f != nil {
		// Destroy Dataset
		f.Destroy(zfs.DestroyRecursive)
		ok(err)
	}
	// Create Dataset
	zfs.CreateFilesystem("tank/test", nil)
	ok(err)

	// Create Snapshot
	s, err := f.Snapshot(snapName(), false)
	ok(err)

	// Create Clone
	s.Clone("tank/clones", nil)
	ok(err)

	// Rollback
	s.Rollback(true)
	ok(err)
}
