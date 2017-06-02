package main

import (
	"fmt"
	"net"
	"github.com/mistifyio/go-zfs"
)

const (
	CONN_HOST = ""
	CONN_PORT = "7766"
	CONN_TYPE = "tcp"
)

func main() {
	// Listen for incoming connections
	l, _ := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	// Close the listener when the application closes
	defer l.Close()
	fmt.Println("[ZEPLIC] Listening on port 7766...")
	for {
		// Listen for an incoming connection.
		conn, _ := l.Accept()
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests
func handleRequest(conn net.Conn) {
	// Number of snapshots in server side
	list, _ := zfs.Snapshots("tank/replication")
	i := len(list)
	if i == 0 {
		conn.Write([]byte("0"))
	} else {
		conn.Write([]byte("1"))
	}

	// Receive snapshot
	zfs.ReceiveSnapshot(conn, "tank/replication", false)

	// Update the list of snapshots
	list, _ = zfs.Snapshots("tank/replication")
	k := len(list)
	// Get the last snapshot available
	last := list[k-1].Name

	// Check if the snapshot (or incremental stream) received already existed
	if i == 0 && k == 1 {
		fmt.Printf("\n[INFO] the snapshot '%s' has been received\n", last)
	} else if i == k {
		fmt.Printf("\n[INFO] the snapshot received '%s' already existed.\n", last)
		fmt.Println(i)
		fmt.Println(k)
	} else {
		fmt.Printf("\n[INFO] the snapshot '%s' has been received.\n", last)
	}
}
