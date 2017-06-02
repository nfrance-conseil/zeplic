// Package config contains: json.go - message.go - syslog.go
//
// Syslog establishes a new connection with the syslog daemon
// and writes (firstly, can create the file) in the log file,
// all messages return by the functions
//
package config

import (
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	SyslogFilePath string
	SyslogLocal syslog.Priority
)

func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func SyslogLocalVar() syslog.Priority {
	s, _ := ioutil.ReadFile(SyslogFilePath)
	scan := string(s)
	exp := regexp.MustCompile(`.*local.*zeplic.log`)
	match := exp.FindStringSubmatch(scan)
	expJoin := strings.Join(match, " ")
	log := between(expJoin, "local", ".")
	switch log {
	case "0":
		SyslogLocal := syslog.LOG_LOCAL0
		return SyslogLocal
	case "1":
		SyslogLocal := syslog.LOG_LOCAL1
		return SyslogLocal
	case "2":
		SyslogLocal := syslog.LOG_LOCAL2
		return SyslogLocal
	case "3":
		SyslogLocal := syslog.LOG_LOCAL3
		return SyslogLocal
	case "4":
		SyslogLocal := syslog.LOG_LOCAL4
		return SyslogLocal
	case "5":
		SyslogLocal := syslog.LOG_LOCAL5
		return SyslogLocal
	case "6":
		SyslogLocal := syslog.LOG_LOCAL6
		return SyslogLocal
	default:
		SyslogLocal := syslog.LOG_LOCAL7
		return SyslogLocal
	}
	return SyslogLocal
}

// LogCreate creates a new log file if it does not exist
func LogCreate() error {
	// Open or create log file
	logFile := "/var/log/zeplic.log"
	_, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		file, err := os.Create(logFile)
		// Send a HUP signal to syslog daemon
		exec.Command("sh", "-c", "pkill -SIGHUP syslogd").Run()
		defer file.Close()
		if err != nil {
			fmt.Printf("\n[ERROR] config/syslog.go:83 *** Error creating log file '%s' ***\n\n", logFile)
			os.Exit(1)
		}
	}
	return err
}

// LogBook creates a new connection with the syslog service
func LogBook() (*syslog.Writer, error) {
	// Establishe a new connection to the system log daemon
	sysLog, err := syslog.New(SyslogLocalVar()|syslog.LOG_ERR|syslog.LOG_WARNING|syslog.LOG_NOTICE|syslog.LOG_INFO, "zeplic")
	if err != nil {
		fmt.Printf("\n[ERROR] config/syslog.go:98 *** Unable to establish a new connection to the system log daemon ***\n\n")
		os.Exit(1)
	}
	return sysLog, nil
}
