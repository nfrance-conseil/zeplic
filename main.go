// zeplic main package
//
// Description...
//
package main

import (
	"os"

	"zeplic/config"
	"zeplic/api"
)

func main () {
	// Read JSON configuration file
	go config.LogCreate()

	// Read JSON configuration file
	j, _, _ := config.Json()

	// Invoke RealMain() function
	os.Exit(api.RealMain(j))
}
