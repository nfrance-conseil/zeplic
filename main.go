// zeplic main package - June 2017 version
//
// ZEPLIC is an application to manage ZFS datasets.
// It establishes a connection with the syslog system service,
// make a synchronisation with Consul,
// reads the dataset configuration of a JSON file
// and execute ZFS functions:
//
// Get a dataset, get a list of snapshots, create a snapshot,
// delete it, create a clone, roll back snapshot, send a snapshot...
//
package main

import (
//	"flag"
	"fmt"
//	"io/ioutil"
	"net"
	"os"
//	"time"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/director"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/pborman/getopt/v2"
//	"github.com/sevlyar/go-daemon"
)

func main () {
	// Available flags
	optHelp := getopt.BoolLong("help", 0, "Help")
	optAgent := getopt.BoolLong("agent", 'a', "Listen ZFS orders from director")
	optDirector := getopt.BoolLong("director", 'd', "Send ZFS orders to agent")
	optQuit := getopt.BoolLong("quit", 0, "Gracefully shutdown")
	optReload := getopt.BoolLong("reload", 0, "Restart zeplic to sleep state")
	optRun := getopt.BoolLong("run", 'r', "Start zeplic as background")
	optSlave := getopt.BoolLong("slave", 's', "Receive a new snapshot from agent")
	optVersion := getopt.BoolLong("version", 'v', "Show version of zeplic")
	getopt.Parse()

	// Help case
	if *optHelp {
		getopt.Usage()
		fmt.Println("")
		os.Exit(0)
	}
/*
	// Create pid file
	cntxt := &daemon.Context{
		PidFileName: "/var/run/zeplic.pid",
		PidFilePerm: 0644,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[Z]"},
	}
	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			w.Warning("[WARNING] unable send signal to the daemon!")
		}
		daemon.SendCommands(d)
		return
	}
	d, err := cntxt.Reborn()
	if d != nil {
		return
	}
	if err != nil {
		os.Exit(1)
	}
	defer cntxt.Release()
*/
	// Checking the flag received
	switch {

	// AGENT
	case *optAgent:
		// Listen for incoming connections
		l, _ := net.Listen("tcp", ":7711")
		defer l.Close()
		fmt.Println("[AGENT:7711] Receiving orders from director...")

		// Loop to accept a new connection
		stop := true
		for stop {
			// Accept a new connection
			connAgent, _ := l.Accept()

			// Handle connection in a new goroutine
			stop = director.HandleRequestAgent(connAgent)
		}

	// DIRECTOR
	case *optDirector:
		fmt.Printf("[INFO] director case inoperative...\n\n")
		os.Exit(0)

	// QUIT
	case *optQuit:
		fmt.Printf("[INFO] quit case inoperative...\n\n")
		os.Exit(0)

	// RELOAD
	case *optReload:
		fmt.Printf("[INFO] reload case inoperative...\n\n")

	// RUN
	case *optRun:
		// Read JSON configuration file
		j, _, _ := config.JSON()

		// Invoke RealMain() function
		os.Exit(lib.Runner(j))

	// SLAVE
	case *optSlave:
		// Listen for incoming connections
		l, _ := net.Listen("tcp", ":7722")
		defer l.Close()
		fmt.Println("[SLAVE:7722] Receiving orders from agent...")

		// Loop to accept a new connection
		stop := true
		for stop {
			// Accept a new connection
			connSlave, _ := l.Accept()

			// Handle connection in a new goroutine
			stop = director.HandleRequestSlave(connSlave)
		}

	// VERSION
	case *optVersion:
		fmt.Printf("[INFO] version case inoperative...\n\n")
		os.Exit(0)
	}
}
