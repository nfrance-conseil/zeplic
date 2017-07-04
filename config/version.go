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
	// Commit gets the last commit
	Commit	    string
	// Version show the version of zeplic
	Version     string
)

// ShowVersion shows the version of zeplic
func ShowVersion() string {
	version := fmt.Sprintf("zeplic v%s\nBuilt on %s [Commit: %s]\n\n", Version, BuildTime, Commit)
	return version
}
