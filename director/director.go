// Package director contains: agent.go - director.go - slave.go
//
// Director sends an order to the agent
// Make orders from synchronisation between nodes
//
package director

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/nfrance-conseil/zeplic/config"
)

var (
	w = config.LogBook()
	// ServerFilePath returns the path of JSON config file
	ServerFilePath string
)

// Create contains the information to create a new snapshot
type Create struct {
	Creation   string `json:"creation"`
	Type       string `json:"type"`
	SyncOn     string `json:"sync_on"`
	SyncPolicy string `json:"sync_policy"`
}

// Policy contains the information of retention policy
type Policy struct {
	Daily   int  `json:"daily"`
	Weekly  bool `json:"weekly"`
	Monthly bool `json:"monthly"`
	Annual  bool `json:"annual"`
}

// Actions contains the information of replicate every snapshot
type Actions struct {
	Hostname	 string `json:"hostname"`
	Dataset		 string `json:"dataset"`
	Snapshot	 Create
	Retention	 Policy
	RollbackIfNeeded bool	`json:"rollback_needed"`
	SkipIfRenamed    bool	`json:"skip_renamed"`
	SkipIfNotWritten bool	`json:"skip_not_written"`
	SkipIfCloned     bool	`json:"skip_cloned"`
}

// Config extracts the interface of JSON server file
type Config struct {
	Snapshots []Config `json:"snapshots"`
}

// Status for DestDataset
const (
	DatasetTrue    = iota + 1 // Dataset not empty
	DatasetFalse		  // Dataset does not exist or empty
	DatasetDisable		  // Dataset disabled
	DatasetDocker             // Dataset docker
	DatasetNotConf		  // Dataset not configured
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

// ZFSDirectorsOrder is the struct for the director's orders
type ZFSDirectorsOrder struct {
	OrderUUID        string `json:"OrderUUID"`	  // mandatory
	Action           string `json:"Action"`		  // take_snapshot, send_snapshot, destroy_snapshot
	Destination      string `json:"Destination"`	  // hostname or IP for send
	SnapshotUUID     string `json:"SnapshotUUID"`	  // mandatory
	SnapshotName     string `json:"SnapshotName"`	  // name of snapshot
	DestDataset      string `json:"DestDataset"`	  // dataset for receive
	RollbackIfNeeded bool   `json:"RollbackIfNeeded"` // should I rollback if written is true on destination
	SkipIfRenamed    bool   `json:"SkipIfRenamed"`	  // should I do the stuff if a snapshot has been renamed
	SkipIfNotWritten bool   `json:"SkipIfNotWritten"` // should I take a snapshot if nothing is written
	SkipIfCloned     bool   `json:"SkipIfCloned"`	  // should I delete a snapshot if it was cloned
}

func Director() {
	jsonFile := ServerFilePath
	serverFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("[INFO] The file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Config
	err = json.Unmarshal(serverFile, &values)
	if err != nil {
		w.Err("[ERROR > order/director.go:XX] it was not possible to parse the JSON configuration file.")
	}

	// READ CONSUL KV PAIRS
	// zeplic/$HOSTNAME/$UUID:$NAME

	// CREATE ORDER AND SEND IT TO AGENT
}
