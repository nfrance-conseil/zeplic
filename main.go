package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/mistifyio/go-zfs"
)

type settings struct {
	Dataset string
	Clones  string
}

func config() (string, string) {
	configFile, err := ioutil.ReadFile("/etc/zeplic.d/config.json")
	var jsontype settings
	json.Unmarshal(configFile, &jsontype)
	ok(err)
	return jsontype.Dataset, jsontype.Clones
}

func ok(err error) {
	if err != nil {
		fmt.Printf("\033[31mUnexpected error! %s\033[39m\n", err.Error())
		os.Stderr.WriteString(err.Error())
	}
}

func snapName() string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", "SNAP", year, month, day, hour, min, sec)
	return snapDate
}

func main () {
	//JSON
	dataset, clones := config()

	// Get Clones
	cl, err := zfs.GetDataset(clones)
	ok(err)
	if cl != nil {
		// Destroy Clones
		cl.Destroy(zfs.DestroyRecursiveClones)
		ok(err)
	}

	// Get Dataset
	ds, err := zfs.GetDataset(dataset)
	ok(err)
	// Destroy Dataset
/*	ds.Destroy(zfs.DestroyRecursive)
	ok(err)*/
	if ds == nil {
		// Create Dataset
		zfs.CreateFilesystem(dataset, nil)
		ok(err)
	}

	// Get Snapshots
/*	zfs.Snapshots(dataset)
	ok(err)*/
	// Create Snapshot
	s, err := ds.Snapshot(snapName(), false)
	ok(err)

	// Create Clone
	s.Clone(clones, nil)
	ok(err)

	// Rollback
/*	s.Rollback(true)
	ok(err)*/
}
