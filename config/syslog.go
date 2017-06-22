// Package config contains: json.go - syslog.go
//
// Syslog establishes a new connection with the syslog daemon
// and writes (firstly, can create the file) in the log file,
// all messages return by the functions
//
package config

import (
	"fmt"
	"log/syslog"
	"os"
)

// LogBook creates a new connection with the syslog service
func LogBook() (*syslog.Writer, error) {
	// Establishe a new connection to the system log daemon
	sysLog, err := syslog.Dial("udp", "localhost:514", syslog.LOG_ERR|syslog.LOG_WARNING|syslog.LOG_NOTICE|syslog.LOG_INFO, "zeplic")
	if err != nil {
		fmt.Printf("\n[ERROR] config/syslog.go:90 *** Unable to establish a new connection to the system log daemon ***\n\n")
		os.Exit(1)
	}
	return sysLog, nil
}
