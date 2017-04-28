package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/mistifyio/go-zfs"
)

type settings struct {
	Dataset string
	Clones  string
	Retain  int
}

func config() (string, string, int) {
	configFile, err := ioutil.ReadFile("/etc/zeplic.d/config.json")
	var jsontype settings
	json.Unmarshal(configFile, &jsontype)
	ok(err)
	return jsontype.Dataset, jsontype.Clones, jsontype.Retain
}

func ok(err error) {
	if err != nil {
		fmt.Printf("\033[31mUnexpected error! %s\033[39m\n", err.Error())
		// Call to Error() function in 'go-zfs' package
		os.Stderr.WriteString(err.Error())
	}
}

func snapName() string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", "SNAP", year, month, day, hour, min, sec)
	// snapName == SNAP_yyyy-Month-dd_HH:MM:SS	
	return snapDate
}

func between(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func main () {
	//JSON
	dataset, clones, retain := config()

	// Get clones dataset
	cl, err := zfs.GetDataset(clones)
	ok(err)
	if cl != nil {
		// Destroy clones dataset
		cl.Destroy(zfs.DestroyRecursiveClones)
		ok(err)
	}

	// Get dataset (called in JSON file)
	ds, err := zfs.GetDataset(dataset)
	ok(err)
	// Destroy dataset (optional)
/*	ds.Destroy(zfs.DestroyRecursive)
	ok(err)*/
	if ds == nil {
		// Create dataset if it does not exist
		zfs.CreateFilesystem(dataset, nil)
		ok(err)
	}

	// Return the number of existing snapshots
	count, err := zfs.Snapshots(dataset)
	ok(err)
	k := len(count)
	if k > 0 {
		// Save the last #Retain(JSON file) snapshots
		for ; k > retain; k-- {
			list, err := zfs.Snapshots(dataset)
			ok(err)
			justList := fmt.Sprintf("%s", list)
			take := between(justList, "{", " ")
			snap, err := zfs.GetDataset(take)
			ok(err)
			snap.Destroy(zfs.DestroyDefault)
			ok(err)
		}
	}

	// Create a new snapshot
	s, err := ds.Snapshot(snapName(), false)
	ok(err)

	// Create a clone of last snapshot
	s.Clone(clones, nil)
	ok(err)

	// Rollback of last snapshot
/*	s.Rollback(true)
	ok(err)*/
}
