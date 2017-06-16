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
	"regexp"
	"strings"
)

var (
	// SyslogPath returns the path of syslog system service
	SyslogPath string
	// SyslogFilePath /var/log/zeplic.log
	SyslogFilePath string
	// SyslogLocal applies the priority
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

// SyslogLocalVar returns the priority of LOG_LOCAL variable
func SyslogLocalVar() syslog.Priority {
	s, _ := ioutil.ReadFile(SyslogPath)
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
}

// LogFile check if a Logfile already exists
func LogFile() error {
	_, err := os.Stat(SyslogFilePath)
	if os.IsNotExist(err) {
		fmt.Printf("\nThe Logfile '%s' does not exist. Please, type 'sudo make install' again...\n\n", SyslogFilePath)
		os.Exit(1)
	}
	return err
}

// LogBook creates a new connection with the syslog service
func LogBook() (*syslog.Writer, error) {
	// Establishe a new connection to the system log daemon
	sysLog, err := syslog.New(SyslogLocalVar()|syslog.LOG_ERR|syslog.LOG_WARNING|syslog.LOG_NOTICE|syslog.LOG_INFO, "zeplic")
	if err != nil {
		fmt.Printf("\n[ERROR] config/syslog.go:90 *** Unable to establish a new connection to the system log daemon ***\n\n")
		os.Exit(1)
	}
	return sysLog, nil
}
