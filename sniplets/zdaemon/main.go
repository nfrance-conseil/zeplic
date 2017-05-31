package main

import (
//	"bytes"
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
//	"time"

	"github.com/sevlyar/go-daemon"
)

var (
	signal = flag.String("z", "", `zdaemon -z ...
		  quit -- graceful shutdown
		reload -- reloading the configuration file`)
)

var (
	quit = make(chan struct{})
	done = make(chan struct{})
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
	go Worker()

//	quit <- struct{}{}
//	<-done

	err = daemon.ServeSignals()
	if err != nil {
		fmt.Printf("STOPPED!\n")
/*		w.Notice("[NOTICE] zeplic has been stopped.")*/
/*		w.Notice("[NOTICE] zeplic graceful shutdown.")*/
		return
	}
}

func Writer() error {
/*	fileHandle, _ := os.Create("/root/output")
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, "String I want to write")
	writer.Flush()*/

	leer, _ := ioutil.ReadFile("/root/test")
	contents := string(leer)

	fileHandle, _ := os.Create("/root/stest")
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, contents)
	writer.Flush()

	fmt.Println("Writed!")
	return nil
}

func Worker() {
//	ticker := time.NewTicker(time.Millisecond)
	for {
/*
		// Read JSON configuration file
		j, _, _ := config.JSON()

		// Invoke RealMain() function
		os.Exit(api.RealMain(j))
*/
		select {
		case <- quit:
			done <- struct{}{}
			return
		default:
		}
		go Writer()
//		time.Sleep(time.Second)
	}
//	done <- struct{}{}
//	ticker.Stop()
}

func TermHandler(sig os.Signal) error {
/*	w.Notice("[NOTICE] zeplic is being stopped...")*/
/*	w.Notice("[NOTICE] zeplic graceful shutdown.")*/
	quit <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func ReloadHandler(sig os.Signal) error {
/*	w.Warninge("[WARNING] zeplic configuration reloaded!")*/
/*	done <- struct{}{}
	if sig == syscall.SIGHUP {
		<-done
	}*/
	return nil
}
