package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mistifyio/go-zfs"
)

type settings struct {
	Dataset string
	Clone  string
	Retain  int
}

func config() (string, string, int) {
	jsonFile := "/etc/zeplic.d/config.json"
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("\nThe file '"+jsonFile+"' does not exist!\n\n")
	}
	var jsontype settings
	json.Unmarshal(configFile, &jsontype)
//	syslog.Err("it was not possible to parse the JSON configuration file.")
	return jsontype.Dataset, jsontype.Clone, jsontype.Retain
}

// Define the name of the snapshot: SNAP_yyyy-Month-dd_HH:MM:SS
func snapName() string {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	snapDate := fmt.Sprintf("%s_%d-%s-%02d_%02d:%02d:%02d", "SNAP", year, month, day, hour, min, sec)
	return snapDate
}

// Get substring with the name of the snapshots
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
	// Open or create log file
	var logPath = "/var/log/zeplic.log"
	var _, err = os.Stat(logPath)
	if os.IsNotExist(err) {
		var file, err = os.Create(logPath)
		// Send a HUP signal to syslog daemon
		exec.Command("csh", "-c", "pkill -SIGHUP syslogd").Run()
		defer file.Close()
		if err != nil {
			fmt.Printf("\nError creating log file '%s'...\n\n", logPath)
			return
		}
	}
	// Establishe a new connection to the system log daemon
	sysLog, err := syslog.New(syslog.LOG_LOCAL0|syslog.LOG_ALERT|syslog.LOG_DEBUG|syslog.LOG_ERR|syslog.LOG_INFO|syslog.LOG_WARNING, "zeplic")
	if err != nil {
		log.Fatal(err)
	}

	//JSON
	dataset, clone, retain := config()

	// Get clones dataset
	cl, err := zfs.GetDataset(clone)
	if err != nil {
		sysLog.Info("the clone '"+clone+"' does not exist.")
	} else {
		// Destroy clones dataset
		err := cl.Destroy(zfs.DestroyRecursiveClones)
		if err != nil {
			sysLog.Err("it was not possible to destroy the clone '"+clone+"'.")
		}
	}

	// Get dataset (called in JSON file)
	ds, err := zfs.GetDataset(dataset)
	if err != nil {
		sysLog.Info("the dataset '"+dataset+"' does not exist.")
	}
	// Destroy dataset (optional)
/*	err := ds.Destroy(zfs.DestroyRecursive)
	if err != nil {
		sysLog.Err("it was not possible to destroy the dataset '"+dataset+"'.")
	}*/
	if ds == nil {
		// Create dataset if it does not exist
		zfs.CreateFilesystem(dataset, nil)
		sysLog.Err("it was not possible to create the dataset '"+dataset+"'.")
	}

	// Return the number of existing snapshots
	count, err := zfs.Snapshots(dataset)
	if err != nil {
		sysLog.Err("it was not possible to access of snapshots list.")
	}
	k := len(count)
	if k > 0 {
		// Save the last #Retain(JSON file) snapshots
		for ; k > retain; k-- {
			list, err := zfs.Snapshots(dataset)
			if err != nil {
				sysLog.Err("it was not possible to access of snapshots list.")
			}
			justList := fmt.Sprintf("%s", list)
			take := between(justList, "{", " ")
			snap, err := zfs.GetDataset(take)
			if err != nil {
				sysLog.Err("it was not possible to get the snapshot '"+take+"'.")
			}
			snap.Destroy(zfs.DestroyDefault)
			if err != nil {
				sysLog.Err("it was not possible to destroy the snapshot '"+take+"'.")
			}
		}
	}

	// Create a new snapshot
	s, err := ds.Snapshot(snapName(), false)
	if err != nil {
		sysLog.Err("it was not possible to create the snapshot '"+snapName()+"'.")
	}

	// Create a clone of last snapshot
	s.Clone(clone, nil)
	if err != nil {
		sysLog.Err("it was not possible to clone the snapshot '"+snapName()+"'.")
	}

	// Rollback of last snaphot
/*	s.Rollback(true)
	if err != nil {
		sysLog.Err("it was not possible to rolling back the snapshot '"+snapName()+"'.")
	}*/
}
