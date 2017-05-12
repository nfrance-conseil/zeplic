// Functions to read the log file (or to create it) and to establishe a new connection with the syslog daemon
package config

import (
	"errors"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
)

// LogCreate() creates a new log file if it does not exist
func LogCreate() error {
	// Open or create log file
	logFile := "/var/log/zeplic.log"
	_, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		file, err := os.Create(logFile)
		// Send a HUP signal to syslog daemon
		exec.Command("csh", "-c", "pkill -SIGHUP syslogd").Run()
		defer file.Close()
		if err != nil {
			fmt.Errorf("\n[ERR] config/syslog.go:18 ~> func LogCreate() *** Error creating log file '%s' ***\n\n", logFile)
			os.Exit(1)
		}
	}
	return nil
}

// LogBook() creates a new connection with the syslog service
func LogBook() (*syslog.Writer, error) {
	// Establishe a new connection to the system log daemon
	sysLog, err := syslog.New(syslog.LOG_LOCAL0|syslog.LOG_ALERT|syslog.LOG_DEBUG|syslog.LOG_ERR|syslog.LOG_INFO|syslog.LOG_WARNING, "zeplic")
	if err != nil {
		errors.New("\n[ERR] config/syslog.go:33 ~> func LogBook() *** Unable to establish a new connection to the system log daemon ***\n\n")
		os.Exit(1)
	}
	return sysLog, nil
}
