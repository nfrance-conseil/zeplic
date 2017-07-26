// Package config contains: json.go - signal.go - syslog.go - version.go
//
// Json reads and extracts the information JSON configuration file
//
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	w = LogBook()
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

// JSON reads the 'JSON' file and checks how many datasets are there
func JSON() (int, string, error) {
	jsonFile := LocalFilePath
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("[INFO] The file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Pool
	err = json.Unmarshal(configFile, &values)
	if err != nil {
		w.Err("[ERROR > config/json.go:60] it was not possible to parse the JSON configuration file.")
	}
	return len(values.Dataset), jsonFile, nil
}

// Extract returns the value of each configuration file field
func Extract(i int) ([]interface{}) {
	_, path, _ := JSON()
	configFile, _ := ioutil.ReadFile(path)
	var values Pool
	json.Unmarshal(configFile, &values)

	enable	    := values.Dataset[i].Enable
	docker	    := values.Dataset[i].Docker
	dataset	    := values.Dataset[i].Name
	consul	    := values.Dataset[i].Consul.Enable
	datacenter  := values.Dataset[i].Consul.Datacenter
	prefix	    := values.Dataset[i].Prefix
	retention   := values.Dataset[i].Retention
	getBackup   := values.Dataset[i].Backup
	getClone    := values.Dataset[i].Clone.Enable
	clone	    := values.Dataset[i].Clone.Name
	delClone    := values.Dataset[i].Clone.Delete

	pieces := []interface{}{enable, docker, dataset, consul, datacenter, prefix, retention, getBackup, getClone, clone, delClone}
	return pieces
}
