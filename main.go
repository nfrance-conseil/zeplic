package main

import (
	"github.com/mistifyio/go-zfs"
)

func main () {
	// Get Clones
	c, _ := zfs.GetDataset("tank/clones")
	if c != nil {
		// Destroy Clones
		c.Destroy(zfs.DestroyRecursiveClones)
	}

	// Get Dataset
	f, _ := zfs.GetDataset("tank/test")
	if f != nil {
		f.Destroy(zfs.DestroyRecursive)
		zfs.CreateFilesystem("tank/test", nil)

		// Create Snapshot
		f.Snapshot("snap1", false)
		s2, _ := f.Snapshot("snap2", false)
		s3, _ := f.Snapshot("snap3", false)

		// Create Clone
		s2.Clone("tank/clones", nil)

		// Rollback
		s3.Rollback(true)
	} else {
		// Create Dataset
		zfs.CreateFilesystem("tank/test", nil)

		// Create Snapshot
		f.Snapshot("snap1", false)
		s2, _ := f.Snapshot("snap2", false)
		s3, _ := f.Snapshot("snap3", false)

		// Create Clone
		s2.Clone("tank/clones", nil)

		// Rollback
		s3.Rollback(true)
	}
}

