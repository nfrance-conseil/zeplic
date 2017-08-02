// Package director contains: agent.go - consul.go - director.go - extract.go - slave.go
//
// Extract keep the struct of server json file and extracts all data
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
)


// ServerFilePath returns the path of JSON config file
var ServerFilePath string

// Cold contains the information of backup snapshot
type Cold struct {
	Creation    string `json:"creation"`
	Prefix	    string `json:"prefix"`
	SyncOn      string `json:"sync_on"`
	SyncDataset string `json:"sync_dataset"`
	SyncPolicy  string `json:"sync_policy"`
	Retention   string `json:"retention"`
}

// Hot contains the information of synchronization snapshot
type Hot struct {
	Creation    string `json:"creation"`
	Prefix	    string `json:"prefix"`
	SyncOn      string `json:"sync_on"`
	SyncDataset string `json:"sync_dataset"`
	SyncPolicy  string `json:"sync_policy"`
	Retention   string `json:"retention"`
}

// Actions contains the information of replicate every snapshot
type Actions struct {
	Hostname	 string `json:"hostname"`
	Dataset		 string `json:"dataset"`
	Backup		 Cold
	Sync		 Hot
	RollbackIfNeeded bool	`json:"rollback_needed"`
	SkipIfRenamed    bool	`json:"skip_renamed"`
	SkipIfNotWritten bool	`json:"skip_not_written"`
	SkipIfCloned     bool	`json:"skip_cloned"`
}

// Config extracts the interface of JSON server file
type Config struct {
	Datacenter	string	`json:"datacenter"`
	Director      []Actions `json:"datasets"`
}

// Extract extracts all data from server json file
func Extract() Config {
	jsonFile := ServerFilePath
	serverFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("[NOTICE] The file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Config
	err = json.Unmarshal(serverFile, &values)
	if err != nil {
		w.Err("[ERROR > director/extract.go:71] it was not possible to parse the JSON configuration file.")
	}
	return values
}
