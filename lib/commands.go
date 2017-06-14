// Package lib contains: comnands.go - snapshot.go - uuid.go
package lib

import (
//	"os/exec"
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/mistifyio/go-zfs"
//	"github.com/pborman/uuid"
)

// RealMain is a loop that executes 'ZFS' functions for each dataset enabled
func RealMain(j int) int {
	// Start syslog system service
	w, _ := config.LogBook()

	for i := 0; i < j; i++ {
		// We going to extract all data stored in JSON file
		pieces := config.Extract(i)

		// This value returns if the dataset is enable
		takedataset := pieces[0].(bool)

		// Execute the functions
		if takedataset == true {
			// Get clones dataset
			clone := pieces[1].(string)
			cl, err := zfs.GetDataset(clone)
			if err != nil {
				w.Info("[INFO] the clone '"+clone+"' does not exist.")
			} else {
				// Destroy clones dataset
				err := cl.Destroy(zfs.DestroyRecursiveClones)
				if err != nil {
					w.Err("[ERROR] it was not possible to destroy the clone '"+clone+"'.")
				} else {
					w.Info("[INFO] the clone '"+clone+"' has been destroyed.")
				}
			}

			// Get dataset (called in JSON file)
			dataset := pieces[2].(string)
			ds, err := zfs.GetDataset(dataset)

			// Destroy dataset (optional)
/*			err := ds.Destroy(zfs.DestroyRecursive)
			if err != nil {
				w.Err("[ERROR] it was not possible to destroy the dataset '"+dataset+"'.")
			} else {
				w.Info("[INFO] the dataset '"+dataset+"' has been destroyed.")
			}
			ds, err = zfs.GetDataset(dataset)
			*/
			if err != nil {
				w.Info("[INFO] the dataset '"+dataset+"' does not exist.")

				// Create dataset if it does not exist
				_, err := zfs.CreateFilesystem(dataset, nil)
				if err != nil {
					w.Err("[ERROR] it was not possible to create the dataset '"+dataset+"'.")
				} else {
					w.Info("[INFO] the dataset '"+dataset+"' has been created.")
				}
				ds, _ = zfs.GetDataset(dataset)
			}

			// Create a new snapshot
			snap := pieces[3].(string)
			s, err := ds.Snapshot(SnapName(snap), false)
			if err != nil {
				w.Err("[ERROR] it was not possible to create the snapshot '"+dataset+"@"+SnapName(snap)+"'.")
			} else {
				w.Info("[INFO] the snapshot '"+dataset+"@"+SnapName(snap)+"' has been created.")
			}

			// Assign an uuid to the snapshot
			list, err := zfs.Snapshots(dataset)
			count := len(list)
			take := list[count-1].Name
			if strings.Contains(take, "BACKUP") {
				take = list[count-1].Name
			}
/*			id := uuid.New()
			args := make([]string, 1, 4)
			args[0] = "zfs"
			args = append(args, "set")
			id = strings.Join([]string{":uuid=", id}, "")
			args = append(args, id)
			args = append(args, take)
			idset := strings.Join(args, " ")
			exec.Command("sh", "-c", idset).Run()*/
			go UUID(take)

			// Delete the backup snapshot
			list, err = zfs.Snapshots(dataset)
			if err != nil {
				w.Err("[ERROR] it was not possible to access of snapshots list.")
			}
			count = len(list)
			for k := 0; k < count; k++ {
				take := list[k].Name
				if strings.Contains(take, "BACKUP") {
					snap, err := zfs.GetDataset(take)
					if err != nil {
						w.Err("[ERROR] it was not possible to get the snapshot '"+take+"'.")
					}
					err = snap.Destroy(zfs.DestroyDefault)
					if err != nil {
						w.Err("[ERROR] it was not possible to destroy the snapshot '"+take+"'.")
					} else {
						w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
					}
				}
			}

			// Return the number of snapshots we want to keep
			retain := pieces[4].(int)
			list, err = zfs.Snapshots(dataset)
			if err != nil {
				w.Err("[ERROR] it was not possible to access of snapshots list.")
			}
			count = len(list)
			for k := 0; count > retain; k++ {
				take := list[k].Name
				snap, err := zfs.GetDataset(take)
				if err != nil {
					w.Err("[ERROR] it was not possible to get the snapshot '"+take+"'.")
				}
				err = snap.Destroy(zfs.DestroyDefault)
				if err != nil {
					w.Err("[ERROR] it was not possible to destroy the snapshot '"+take+"'.")
				} else {
					w.Info("[INFO] the snapshot '"+take+"' has been destroyed.")
				}
				list, err = zfs.Snapshots(dataset)
				if err != nil {
					w.Err("[ERROR] it was not possible to access of snapshots list.")
				}
				count = len(list)
			}

			// Create a backup snapshot
			backup := pieces[5].(bool)
			if backup == true {
				_, err := ds.Snapshot(SnapBackup(), false)
				if err != nil {
					w.Err("[ERROR] it was not possible to create the backup snapshot '"+dataset+"@"+SnapBackup()+"'.")
				} else {
					w.Info("[INFO] the backup snapshot '"+dataset+"@"+SnapBackup()+"' has been created.")
				}
			}

			// Create a clone of last snapshot
			takeclone := pieces[6].(bool)
			if takeclone == true {
				_, err = s.Clone(clone, nil)
				if err != nil {
					w.Err("[ERROR] it was not possible to clone the snapshot '"+dataset+"@"+SnapName(snap)+"'.")
				} else {
					w.Info("[INFO] the snapshot '"+dataset+"@"+SnapName(snap)+"' has been clone.")
				}
			}

			// Rollback of last snaphot
			takerollback := pieces[7].(bool)
			if takerollback == true {
				s.Rollback(true)
				if err != nil {
					w.Err("[ERROR] it was not possible to rolling back the snapshot '"+SnapName(snap)+"'.")
				} else {
					w.Info("[INFO] the snapshot '"+dataset+"@"+SnapName(snap)+"' has been restored.")
				}
			}
		}
	}
	return 0
}
