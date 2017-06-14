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

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/director"
//	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/sevlyar/go-daemon"
)

func main () {
	// Create log file if it does not exit
	go config.LogCreate()

	// Start syslog system service
	w, _ := config.LogBook()

	// Read JSON configuration file
/*	j, _, _ := config.JSON()

	// Invoke RealMain() function
	os.Exit(lib.RealMain(j))*/

	// Define flag variable and channels
	var signal = flag.String("z", "", "")
//	var quit = make(chan struct{})

	// Show zeplic help
	flag.Usage = func() {
		fmt.Printf("Usage: zeplic -z <command>\n\n")
		fmt.Printf("   agent\tListen ZFS orders from director\n")
		fmt.Printf("   director\tSend ZFS orders to agent\n")
		fmt.Printf("   quit\t\tGracefully shutdown\n")
		fmt.Printf("   reload\tRestart zeplic to sleep state\n")
		fmt.Printf("   run\t\tStart zeplic as background\n")
		fmt.Printf("   slave\tReceive a new snapshot from agent\n")
		fmt.Printf("   version\tShow version of zeplic\n")
		fmt.Println("")
	}

	// Checking if the command-line arguments are correct
	if len(os.Args) != 3 || os.Args[1] != "-z" {
		flag.Usage()
		os.Exit(1)
	} else {
		flag.CommandLine.SetOutput(ioutil.Discard)
		flag.Parse()
	}

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
		fmt.Printf("[INFO] run case inoperative...\n\n")
		os.Exit(1)

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
		flag.Usage()
		os.Exit(1)
	}
}
