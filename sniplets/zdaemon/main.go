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
	signal = flag.String("z", "", `zdaemon -z ...
		  quit -- graceful shutdown
		reload -- reloading the configuration file`)
)

var (
	done = make(chan struct{})
	quit = make(chan struct{})
	reload = make(chan struct{})
)

func main() {
	// Start syslog daemon service
/*	go config.LogCreate()
	w, _ := config.LogBook()*/

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
		d, _ := cntxt.Search()
/*		if err != nil {
			w.Warning("[WARNING] unable send signal to the daemon!")
		}*/
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
//			w.Notice("[NOTICE] zdaemon graceful shutdown.")
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

/*	err = daemon.ServeSignals()
	if err != nil {
//		w.Notice("[NOTICE] zdaemon has been stopped.")
//		w.Notice("[NOTICE] zdaemon graceful shutdown.")
		return
	}*/
}

func Writer() error {
	read, _ := ioutil.ReadFile("/root/test")
	contents := string(read)

	fileHandle, _ := os.Create("/root/stest")
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, contents)
	writer.Flush()

	return nil
}

func TermHandler(sig os.Signal) error {
//	w.Notice("[NOTICE] zdaemon graceful shutdown.")
	quit <- struct{}{}
	if sig == syscall.SIGQUIT {
//		w.Notice("[NOTICE] zdaemon graceful shutdown.")
		<-done
	}
	return nil
}

func ReloadHandler(sig os.Signal) error {
//	w.Warning("[WARNING] zeplic configuration reloaded!")
/*	reload <- struct{}{}
	if sig == syscall.SIGHUP {
		<-done
	}*/
	return nil
}
