package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"github.com/mistifyio/go-zfs"
)

const (
	CONN_HOST = "192.168.99.5"
	CONN_PORT = "7766"
	CONN_TYPE = "tcp"
)

func main() {
	// Extract the last snapshots in dataset
	list, _ := zfs.Snapshots("tank/test")
	count := len(list)
	take1 := list[count-2].Name
	take2 := list[count-1].Name

	// If the last snapshot is the backup snapshot, it will be rejected
	if strings.Contains(take2, "BACKUP") {
		take1 = list[count-3].Name
		take2 = list[count-2].Name
	}

	// Get the last two snapshots
	ds1, _ := zfs.GetDataset(take1)
	ds2, _ := zfs.GetDataset(take2)

	// New connection
	conn, _ := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)

	// Read data of server side
	buff := bufio.NewReader(conn)
	n, _ := buff.ReadByte()

	// Execute the correct case
	switch string(n) {
	case "0":
		// Send only the last snapshot available
		ds2.SendSnapshot(conn, zfs.SendDefault)
		conn.Close()
		fmt.Printf("\n[INFO] the snapshot '%s' has been sent.\n", take2)
	default :
		// Send an incremental stream of snapshots
		zfs.SendSnapshotIncremental(conn, ds1, ds2, zfs.IncrementalStream)
		conn.Close()
		fmt.Printf("\n[INFO] the snapshot '%s' has been sent.\n", take2)
	}
}
