// zeplic main package
//
// Version * - May 2017
//
// ZEPLIC is an application to manage ZFS datasets.
// It establishes a connection with the syslog system service,
// reads the dataset configuration of a JSON file
// and execute a sequence of ZFS functions:
//
// Get a dataset, get a list of snapshots, create a snapshot,
// delete it, create a clone, roll back snapshot...
//
package main

import (
	"os"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/api"
)

func main () {
	// Start syslog daemon service
	go config.LogCreate()

	// Read JSON configuration file
	j, _, _ := config.JSON()

	// Invoke RealMain() function
	os.Exit(api.RealMain(j))
}
