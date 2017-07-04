// Package config contains: json.go - signal.go - syslog.go - version.go
//
// Syslog establishes a new connection with the syslog daemon
// and writes in the log file, all messages return by the functions
//
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/nfrance-conseil/zeplic/utils"
)

var (
	// SyslogFilePath returns the path of syslog config file
	SyslogFilePath string
	// Facility applies the priority
	Facility syslog.Priority
)

// Facility returns the priority of LOCAL variable
func Logger(log string) syslog.Priority {
	switch log {
	case "LOCAL0":
		Facility := syslog.LOG_LOCAL0
		return Facility
	case "LOCAL1":
		Facility := syslog.LOG_LOCAL1
		return Facility
	case "LOCAL2":
		Facility := syslog.LOG_LOCAL2
		return Facility
	case "LOCAL3":
		Facility := syslog.LOG_LOCAL3
		return Facility
	case "LOCAL4":
		Facility := syslog.LOG_LOCAL4
		return Facility
	case "LOCAL5":
		Facility := syslog.LOG_LOCAL5
		return Facility
	case "LOCAL6":
		Facility := syslog.LOG_LOCAL6
		return Facility
	case "LOCAL7":
		Facility := syslog.LOG_LOCAL7
		return Facility
	default:
		return Facility
	}
}

// Log extracts the interface of JSON file
type Log struct {
	Enable bool   `json:"enable"`
	Mode   string `json:"mode"`
	Info   string `json:"info"`
}

// LogBook checks the configuration of syslog and creates a new connection with the service
func LogBook() *syslog.Writer {
	jsonFile := SyslogFilePath
	configFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil
	} else {
		var values Log
		json.Unmarshal(configFile, &values)

		// Variables
		enable := values.Enable
		mode := values.Mode
		info := values.Info

		switch {
		case enable:
			switch mode {
			// Local
			case "local":
				facility := Logger(info)
				// Establishe a new connection to the system log daemon
				sysLog, err := syslog.New(facility|syslog.LOG_DEBUG|syslog.LOG_ERR|syslog.LOG_WARNING|syslog.LOG_NOTICE|syslog.LOG_INFO, "zeplic")
				if err != nil {
					fmt.Printf("[ERROR] config/syslog.go:86 *** Unable to establish a new connection with syslog service ***\n\n")
					os.Exit(1)
				}
				return sysLog
			// Remote
			case "remote":
				protocol := string(utils.Before(info, ":")) // TCP | UDP
				addr := string(utils.Reverse(info, ":")) // IP address and Port
				// Establishe a new connection to the system log daemon
				sysLog, err := syslog.Dial(protocol, addr, syslog.LOG_DEBUG|syslog.LOG_ERR|syslog.LOG_WARNING|syslog.LOG_NOTICE|syslog.LOG_INFO, "zeplic")
				if err != nil {
					fmt.Printf("[ERROR] config/syslog.go:97 *** Unable to establish a new connection with syslog service ***\n\n")
					os.Exit(1)
				}
				return sysLog
			default:
				fmt.Printf("\n[ERROR] config/syslog.go:81 *** The mode chosen in your syslog config file is not correct (local | remote) ***\n\n")
				os.Exit(1)
				return nil
			}
		default:
			return nil
		}
	}
}
