// Functions to read and extract the information that the JSON file contains
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type Copy struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
}

type Data struct {
	Enable	bool	`json:"enable"`
	Name	string	`json:"name"`
	Snap	string	`json:"snapshot"`
	Retain	int	`json:"retain"`
	Backup	bool	`json:"backup"`
	Clone	Copy
	Roll	bool	`json:"rollback"`
}

type Pool struct {
	Dataset	[]Data	`json:"datasets"`
}

// Json() reads the JSON file and checks how many datasets are there
func Json() (int, string, error) {
	jsonFile := "/usr/local/etc/zeplic.d/config.json"
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("\nThe file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Pool
	err = json.Unmarshal(configFile, &values)
	if err != nil {
		errors.New("\n[ERR] config/json.go:40 ~> func Json() *** Impossible to parse the JSON configuration file ***\n\n")
		os.Exit(1)
	}
	return len(values.Dataset), jsonFile, nil
}

// Extract() returns the value of each configuration file field
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
