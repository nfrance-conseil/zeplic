// Package config contains: json.go - signal.go - syslog.go - version.go
//
// Version gets the commit version and date of last zeplic build
//
package config

import (
	"fmt"
)

var (
	// BuildTime gets date of last build
	BuildTime   string
	// Version gets the version of last ommit
	Version     string
)

// ShowVersion shows the version of zeplic
func ShowVersion() string {
	version := fmt.Sprintf("zeplic preliminar version: %s - %s\n\n", Version, BuildTime)
	return version
}
