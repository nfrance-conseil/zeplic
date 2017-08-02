// Package config contains: json.go - signal.go - syslog.go - version.go
//
// Signal gets the process ID of zeplic and sends to the process SIGTERM/SIGHUP
//
package config

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// PidFilePath gets the path of pid file
var PidFilePath string

// Leave sends a SIGTERM signal to zeplic process ID
func Leave() int {
	// Open pid file
	file, err := os.OpenFile(PidFilePath, os.O_RDWR, 0644)
	if err != nil {
		w.Err("[ERROR > config/signal.go:21] it was not possible to open the file '"+PidFilePath+"'.")
	}
	defer file.Close()

	// Read pid file
	var pidString = make([]byte, 16)
	n, err := file.Read(pidString)
	if err != nil {
		w.Err("[ERROR > config/signal.go:29] it was not possible to read the file '"+PidFilePath+"'.")
	}

	var check int
	if n == 0 {
		check = 1
	} else {
		var pidSlice []int
		b := bytes.Count(pidString, []byte{10})
		for i := 0; i < b; i++ {
			buff := bytes.NewBuffer(pidString)
			line, _ := buff.ReadBytes(10)
			line = bytes.TrimSuffix(line, []byte{10})
			pidInt, _ := strconv.Atoi(string(line))
			pidSlice = append(pidSlice, pidInt)
			pidString = pidString[len(line)+1:]
		}

		// SIGTERM to zeplic
		for j := 0; j < len(pidSlice); j++ {
			syscall.Kill(pidSlice[j], syscall.SIGTERM)
		}

		// Clean the file
		os.Remove(PidFilePath)
		os.Create(PidFilePath)

		w.Notice("[NOTICE] zeplic graceful shutdown...")
	}
	return check
}

// Pid writes the pid of zeplic pid file
func Pid() error {
	// Get pid
	pid := os.Getpid()

	// Open pid file
	file, err := os.OpenFile(PidFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		w.Err("[ERROR > config/signal.go:69] it was not possible to open the file '"+PidFilePath+"'.")
	}
	defer file.Close()

	// Write in pid file
	pidString := strconv.Itoa(pid)
	toWrite := fmt.Sprintf("%s\n", pidString)
	_, err = file.WriteString(toWrite)
	if err != nil {
		w.Err("[ERROR > config/signal.go:78] it was not possible to write the pid in '"+PidFilePath+"'.")
	}
	return nil
}
