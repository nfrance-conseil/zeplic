package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
)

var (
	signal = flag.String("z", "", "\n  zdaemon -z... reload -- restart the configuration\n\t\t  quit -- graceful shutdown\n")
)

var (
	done = make(chan struct{})
	quit = make(chan struct{})
	reload = make(chan struct{})
)

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, TermHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, ReloadHandler)

	cntxt := &daemon.Context{
		PidFileName: "/var/run/zdaemon.pid",
		PidFilePerm: 0644,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[zdaemon]"},
	}
	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			fmt.Println("[WARNING] unable send signal to the daemon!")
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

	// Start daemon
	ticker := time.NewTicker(time.Minute)
	zeplic := true
	for zeplic {
		select {
		case <- ticker.C:
			go Writer()
		case <- quit:
			// Got a quit signal, stopping
			done <- struct{}{}
			zeplic = false
			ticker.Stop()
			return
/*		case <- reload:
			// Got a reload signal, reloading
			done <- struct{}{}
			zeplic = true
			return*/
		default:
			// No stop signal, continuing loop
		}
	}
}

// Writer writes a string into a file
func Writer() error {
	read, _ := ioutil.ReadFile("$HOME/rtest")
	contents := string(read)

	fileHandle, _ := os.Create("$HOME/wtest")
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, contents)
	writer.Flush()

	return nil
}

// TermHandler is the struct for quit signal
func TermHandler(sig os.Signal) error {
	quit <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return nil
}

// ReloadHandler is the struct for reload signal
func ReloadHandler(sig os.Signal) error {
/*	reload <- struct{}{}
	if sig == syscall.SIGHUP {
		<-done
	}*/
	return nil
}
