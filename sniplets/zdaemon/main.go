package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

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
		return
	}
}

func Writer() error {
	read, _ := ioutil.ReadFile("/root/test")
	contents := string(read)

	fileHandle, _ := os.Create("/root/stest")
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()
	fmt.Fprintln(writer, contents)
	writer.Flush()

	fmt.Println("Writed!")
	return nil
}

func Worker() {
	for {
		select {
		case <- quit:
			done <- struct{}{}
			return
		default:
		}
		go Writer()
	}
//	done <- struct{}{}
//	ticker.Stop()
}

func TermHandler(sig os.Signal) error {
	quit <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func ReloadHandler(sig os.Signal) error {
	return nil
}
