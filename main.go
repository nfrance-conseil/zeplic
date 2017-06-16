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
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
//	"time"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/director"
	"github.com/nfrance-conseil/zeplic/lib"
//	"github.com/sevlyar/go-daemon"
)

// Define flag variable and channels
var (
	signal = flag.String("z", "", "")
//	done = make(chan struct{})
//	quit = make(chan struct{})
)

func main () {
	// Checking if the command-line arguments are correct
	switch len(os.Args) {
	case 2:
		if os.Args[1] != "--help" {
			fmt.Printf("zeplic --help\n")
			fmt.Println("")
			os.Exit(1)
		} else {
			config.Usage()
			os.Exit(1)
		}
	case 3:
		if os.Args[1] != "-z" {
			fmt.Printf("zeplic --help\n")
			fmt.Println("")
			os.Exit(1)
		} else {
			// Check if a Logfile already exists
//			go config.LogFile()
			flag.CommandLine.SetOutput(ioutil.Discard)
			flag.Parse()
		}
	default:
		fmt.Printf("zeplic --help\n")
		fmt.Println("")
		os.Exit(1)
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
	switch *signal {

	// AGENT
	case "agent":
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
	case "director":
		fmt.Printf("[INFO] director case inoperative...\n\n")
		os.Exit(1)

	// QUIT
	case "quit":
		fmt.Printf("[INFO] quit case inoperative...\n\n")
		os.Exit(1)

	// RUN
	case "run":
		// Start syslog system service
//		w, _ := config.LogBook()

		// Read JSON configuration file
		j, _, _ := config.JSON()

		// Invoke RealMain() function
		os.Exit(lib.RealMain(j))

	// RELOAD
	case "reload":
		fmt.Printf("[INFO] reload case inoperative...\n\n")
		os.Exit(1)

	// SLAVE
	case "slave":
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
	case "version":
		fmt.Printf("[INFO] version case inoperative...\n\n")
		os.Exit(1)

	// Show zeplic help
	default:
		config.Usage()
		os.Exit(1)
	}
}
