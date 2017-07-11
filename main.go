// zeplic main package - July 2017 version 0.1.0-rc1
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
	"fmt"
	"net"
	"os"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/order"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/pborman/getopt/v2"
)

var (
	// Variable to connect with syslog service
	w = config.LogBook()
)

func main() {
	// Available flags
	optAgent    := getopt.BoolLong("agent", 'a', "Listen ZFS orders from director")
	optDirector := getopt.BoolLong("director", 'd', "Send ZFS orders to agent")
	optHelp     := getopt.BoolLong("help", 0, "Show help menu")
	optQuit	    := getopt.BoolLong("quit", 0, "Gracefully shutdown")
	optRun	    := getopt.BoolLong("run", 'r', "Execute ZFS functions")
	optSlave    := getopt.BoolLong("slave", 's', "Receive a new snapshot from agent")
	optVersion  := getopt.BoolLong("version", 'v', "Show version of zeplic")
	getopt.Parse()

	if len(os.Args) == 1 || len(os.Args) > 2 {
		fmt.Printf("zeplic --help\n\n")
		os.Exit(0)
	}

	// Cases...
	switch {

	// AGENT
	case *optAgent:
		go config.Pid()

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
			stop = order.HandleRequestAgent(connAgent)
		}

	// DIRECTOR
	case *optDirector:
//		config.Pid()
		fmt.Printf("[INFO] director case inoperative...\n\n")
		os.Exit(0)

	// HELP
	case *optHelp:
		getopt.Usage()
		fmt.Println("")
		os.Exit(0)

	// QUIT
	case *optQuit:
		err := config.Leave()
		if err == 1 {
			fmt.Printf("[INFO] zeplic is not running...\n\n")
			os.Exit(0)
		} else {
			os.Exit(0)
		}

	// RUN
	case *optRun:
		// Read JSON configuration file
		j, _, _ := config.JSON()

		// Invoke Runner() function
		os.Exit(lib.Runner(j))

	// SLAVE
	case *optSlave:
		go config.Pid()

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
			stop = order.HandleRequestSlave(connSlave)
		}

	// VERSION
	case *optVersion:
		version := config.ShowVersion()
		fmt.Printf("%s", version)
		os.Exit(0)
	}
}
