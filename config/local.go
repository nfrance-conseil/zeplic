// Package config contains: local.go - server.go - signal.go - syslog.go - version.go
//
// Local reads and extracts the information JSON configuration file
//
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// LocalFilePath returns the path of JSON config file
var LocalFilePath string

// Sync contains Consul options
type Sync struct {
	Enable	   bool	  `json:"enable"`
	Datacenter string `json:"datacenter"`
}

// Copy contains Clone options
type Copy struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
	Delete	bool	`json:"delete"`
}

// Data contains the information of each dataset
type Data struct {
	Enable	   bool	  `json:"enable"`
	Docker	   bool	  `json:"docker"`
	Name	   string `json:"name"`
	Consul	   Sync
	Prefix	   string `json:"snap_prefix"`
	Retention  int	  `json:"snap_retention"`
	Backup	   bool	  `json:"backup"`
	Clone	   Copy
}

// Pool extracts the interface of JSON file
type Pool struct {
	Dataset	[]Data	`json:"local_datasets"`
}

// Local reads the local json file and checks how many datasets are there and its data
func Local() Pool {
	jsonFile := LocalFilePath
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("[NOTICE] The file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Pool
	err = json.Unmarshal(configFile, &values)
	if err != nil {
		w.Err("[ERROR > config/local.go:56] it was not possible to parse the JSON configuration file.")
	}
	return values
}
