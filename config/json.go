// Package config contains: json.go - syslog.go
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

// Copy contains Clone options
type Copy struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
}

// Data contains the information of each dataset
type Data struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
	Snap	string	`json:"snapshot"`
	Retain	int	`json:"retain"`
	Backup	bool	`json:"backup"`
	Clone	Copy
	Roll	bool	`json:"rollback"`
}

// Pool extracts the interface of JSON file
type Pool struct {
	Dataset	[]Data	`json:"datasets"`
}

// Json reads the JSON file and checks how many datasets are there
func Json() (int, string, error) {
	w, _ := LogBook()
	jsonFile := "/usr/local/etc/zeplic.d/config.json"
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("\nThe file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Pool
	err = json.Unmarshal(configFile, &values)
	if err != nil {
		w.Err("[ERROR] config/json.go:40 *** Impossible to parse the JSON configuration file ***")
		os.Exit(1)
	}
	return len(values.Dataset), jsonFile, nil
}

// Extract returns the value of each configuration file field
func Extract(i int) ([]interface{}) {
	_, path, _ := Json()
	configFile, _ := ioutil.ReadFile(path)
	var values Pool
	json.Unmarshal(configFile, &values)

	takedataset := values.Dataset[i].Enable
	clone := values.Dataset[i].Clone.Name
	dataset	:= values.Dataset[i].Name
	snap := values.Dataset[i].Snap
	retain := values.Dataset[i].Retain
	takebackup := values.Dataset[i].Backup
	takeclone := values.Dataset[i].Clone.Enable
	takerollback := values.Dataset[i].Roll

	pieces := []interface{}{takedataset, clone, dataset, snap, retain, takebackup, takeclone, takerollback}
	return pieces
}
