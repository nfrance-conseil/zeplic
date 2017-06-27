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

// ConfigFilePath returns the path of JSON config file
var ConfigFilePath string

// Copy contains Clone options
type Copy struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
	Delete	bool	`json:"delete"`
}

// Data contains the information of each dataset
type Data struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
	Snap	string	`json:"snapshot"`
	Retain	int	`json:"retain"`
	Backup	bool	`json:"backup"`
	Clone	Copy
}

// Pool extracts the interface of JSON file
type Pool struct {
	Dataset	[]Data	`json:"datasets"`
}

// JSON reads the 'JSON' file and checks how many datasets are there
func JSON() (int, string, error) {
	w := LogBook()
	jsonFile := ConfigFilePath
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("[INFO] The file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Pool
	err = json.Unmarshal(configFile, &values)
	if err != nil {
		w.Err("[ERROR] it was not possible to parse the JSON configuration file.")
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
	delClone    := values.Dataset[i].Clone.Delete
	clone	    := values.Dataset[i].Clone.Name
	dataset	    := values.Dataset[i].Name
	snapshot    := values.Dataset[i].Snap
	retain	    := values.Dataset[i].Retain
	getBackup   := values.Dataset[i].Backup
	getClone    := values.Dataset[i].Clone.Enable

	pieces := []interface{}{enable, delClone, clone, dataset, snapshot, retain, getBackup, getClone}
	return pieces
}
