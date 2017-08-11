// zeplic main package
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
	"time"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/director"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/pborman/getopt/v2"
)

var (
	w = config.LogBook()
)

func main() {
	// Available flags
	optAgent    := getopt.BoolLong("agent", 'a', "Execute the orders from director")
	optCleaner  := getopt.BoolLong("cleaner", 'c', "Clean KV pairs with the flags #NotWritten #deleted")
	optDirector := getopt.BoolLong("director", 'd', "Execute 'zeplic' in synchronization mode")
	optHelp     := getopt.BoolLong("help", 0, "Show help menu")
	optQuit	    := getopt.BoolLong("quit", 0, "Gracefully shutdown")
	optRun	    := getopt.BoolLong("run", 'r', "Execute 'zeplic' in local mode")
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
		w.Notice("[NOTICE] 'zeplic --agent' is running... listen on port 7711!")

		// Loop to accept a new connection
		for {
			// Accept a new connection
			connAgent, _ := l.Accept()

			// Handle connection in a new goroutine
			go director.HandleRequestAgent(connAgent)
		}

	// CLEANER
	case *optCleaner:
		var dataset string
		fmt.Println("[CLEANER] Running zeplic cleaner's mode...")
		fmt.Printf("\nPlease, indicate dataset: ")
		fmt.Scanf("%s", &dataset)

		// Call to Cleaner function
		code := lib.Cleaner(dataset)
		if code == 1 {
			fmt.Printf("[CLEANER] An error has occurred while zeplic cleaned the KV pairs, please revise your syslog...\n\n")
		} else {
			fmt.Printf("Done!\n\n")
		}
		os.Exit(code)

	// DIRECTOR
	case *optDirector:
		alive := lib.Alive()
		if alive == true {
			go config.Pid()
			w.Notice("[NOTICE] 'zeplic --director' is running...")

			// Infinite loop to manage the datasets
			ticker := time.NewTicker(1 * time.Minute)
			for {
				select {
				case <- ticker.C:
					go director.Director()
				default:
					// No stop signal, continuing loop
				}
			}
		} else {
			fmt.Printf("[INFO] Consul server is not running...\n\n")
			os.Exit(0)
		}

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
		values := config.Local()

		// Invoke Runner() function
		var code int
		for i := 0; i < len(values.Dataset); i++ {
			code = lib.Runner(i, false, "", false)
			if code == 1 {
				fmt.Printf("[RUNNER] An error has occurred while running zeplic, please revise your syslog...\n\n")
				break
			} else {
				continue
			}
		}
		os.Exit(code)

	// SLAVE
	case *optSlave:
		go config.Pid()

		// Listen for incoming connections
		l, _ := net.Listen("tcp", ":7722")
		defer l.Close()
		w.Notice("[NOTICE] 'zeplic --slave' is running... listen on port 7722!")

		// Loop to accept a new connection
		for {
			// Accept a new connection
			connSlave, _ := l.Accept()

			// Handle connection in a new goroutine
			go director.HandleRequestSlave(connSlave)
		}

	// VERSION
	case *optVersion:
		version := config.ShowVersion()
		fmt.Printf("%s", version)
		os.Exit(0)
	}
}
