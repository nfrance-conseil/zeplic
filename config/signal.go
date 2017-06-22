// Package config contains: json.go - signal.go - syslog.go - version.go
//
// Signal gets the process ID of zeplic and sends to the process SIGTERM/SIGHUP
//
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

var (
	// PidFilePath gets the path of pid file
	PidFilePath string
	// Variable to connect with syslog service
	w, _ = LogBook()
)

// Leave sends a SIGTERM signal to zeplic process ID
func Leave() int {
	pidBytes, _ := ioutil.ReadFile(PidFilePath)
	pid, _ := strconv.Atoi(string(pidBytes))

	// Search if zeplic process exists
	search := fmt.Sprintf("ps -o command= -p %d", pid)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	process := string(out)

	var err int
	if process == "" || pid == 0 {
		err = 1
	} else {
		err = 0
		syscall.Kill(pid, syscall.SIGTERM)
		w.Notice("[NOTICE] zeplic graceful shutdown...")
	}
	return err
}

// Pid writes the pid of zeplic pid file
func Pid() error {
	pid := os.Getpid()
	pidBytes := []byte(strconv.Itoa(pid))
	ioutil.WriteFile(PidFilePath, pidBytes, 0644)
	return nil
}
/*
// Reload sends a SIGHUP signal to zeplic process ID
func Reload() (int, string) {
	pidBytes, _ := ioutil.ReadFile(PidFilePath)
	pid, _ := strconv.Atoi(string(pidBytes))

	// Restart original zeplic process
	search := fmt.Sprintf("ps -o command= -p %d", pid)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	process := string(out)

	var err int
	if process == "" || pid == 0 {
		err = 1
	} else {
		err = 0
		syscall.Kill(pid, syscall.SIGHUP)
		w.Notice("[NOTICE] restarting zeplic...")
	}
	return err, process
}
*/
