// Package order contains: agent.go - !director.go - slave.go
//
// Agent executes the orders received from director
//
package order

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// ZFSOrderFromDirector is the struct for ZFS orders from director
type ZFSOrderFromDirector struct {
	OrderUUID	 string	// mandatory
	Action		 string // take_snapshot, send_snapshot, destroy_snapshot
	Destination	 string // hostname or IP for send
	SnapshotUUID	 string // mandatory
	SnapshotName	 string	// name of snapshot
	DestDataset	 string // dataset for receive
	RollbackIfNeeded bool   // should I rollback if written is true on destination
	SkipIfRenamed	 bool   // should I do the stuff if a snapshot has been renamed
	SkipIfNotWritten bool   // should I take a snapshot if nothing is written
}

// ZFSOrderToSlave is the struct for ZFS orders to slave
type ZFSOrderToSlave struct {
	Hostname	string `json:"Source"`
	OrderUUID	string `json:"OrderUUID"`
	SnapshotUUID	string `json:"SnapshotUUID"`
	SnapshotName	string `json:"SnapshotName"`
	DestDataset	string `json:"DestDataset"`
}

// ZFSResponseFromSlave is the struct for ZFS response from slave
type ZFSResponseFromSlave struct {
	OrderUUID    string  // reference to a valid order
	IsSuccess    bool    // true or false
	Status	     int64   // 
	Error	     string  // error string if needed
}

// HandleRequestAgent incoming requests from director
func HandleRequestAgent (connAgent net.Conn) bool {
	// Start syslog system service
	w := config.LogBook()

	// Resolve hostname
	hostname, err := os.Hostname()
	if err != nil {
		w.Err("[ERROR] it was not possible to resolve the hostname.")
	}

	// Unmarshal orders from director
	var d ZFSOrderFromDirector
	director, err := ioutil.ReadAll(connAgent)
	if err != nil {
		w.Err("[ERROR] an error has occurred while reading from the socket.")
	}
	err = json.Unmarshal(director, &d)
	if err != nil {
		w.Err("[ERROR] it was impossible to parse the JSON struct from the socket.")
	}
	if d.OrderUUID == "" || d.Action == "" || (d.Action == "send_snapshot" && d.Destination == "") {
		w.Err("[ERROR] inconsistant data structure in ZFS order.")
	}

	// Switch for action order
	switch d.Action {

	// Create a new snapshot
	case "take_snapshot":
		// Check if the DestDataset exists and it is enable
		if d.DestDataset == "" {
			w.Err("[ERROR] inconsistant data structure in ZFS order.")
			break
		}

		// Read JSON configuration file
		j, _, _ := config.JSON()

		// Call to function CommandOrder for create the snapshot
		lib.TakeOrder(j, d.DestDataset)

	// Send snapshot to d.Destination
	case "send_snapshot":
		// Checking required information
		if d.SnapshotUUID == "" || d.Destination == "" || d.DestDataset == "" {
			w.Err("[ERROR] inconsistant data structure in ZFS order.")
			break
		}
		// Search the snapshot name from its uuid
		SnapshotName := lib.SearchName(d.SnapshotUUID)
		// Take the snapshot
		ds, _ := zfs.GetDataset(SnapshotName)

		// Create a new connection with the destination
		connToSlave, _ := net.Dial("tcp", d.Destination+":7722")

		// Struct for ZFS orders to slave
		ZFSOrderToSlave := ZFSOrderToSlave{hostname,d.OrderUUID,d.SnapshotUUID,SnapshotName,d.DestDataset}
		ots, err := json.Marshal(ZFSOrderToSlave)
		if err != nil {
			w.Err("[ERROR] it was impossible to encode the JSON struct.")
		}
		connToSlave.Write([]byte(ots))
		connToSlave.Write([]byte("\n"))

		// Read from destinantion if Dataset exists
		buff := bufio.NewReader(connToSlave)
		n, _ := buff.ReadByte()
		dsExist, _ := strconv.Atoi(string(n))

		switch dsExist {

		// Case: dataset exist on destination
		case DatasetTrue:
			// Check uuid of last snapshot on destination
			connToSlave.Close()

			// Reconnection to get ZFSResponse
			l2, _ := net.Listen("tcp", ":7733")
			defer l2.Close()
			fmt.Println("[AGENT:7733] Receiving response from slave...")

			conn2Agent, _ := l2.Accept()

			var r ZFSResponseFromSlave
			response, err := bufio.NewReader(conn2Agent).ReadBytes('\x0A')
			if err != nil {
				w.Err("[ERROR] an error has occurred while reading from the socket.")
				break
			}
			err = json.Unmarshal(response, &r)
			if err != nil {
				w.Err("[ERROR] it was impossible to parse the JSON struct from the socket.")
				break
			}
			if r.IsSuccess == true {
				switch r.Status {
				// Snapshot renamed
				case WasRenamed:
					w.Info("[INFO] the snapshot '"+SnapshotName+"' has been renamed to '"+r.Error+"'.")
				// Nothing to do
				case NothingToDo:
					w.Info("[INFO] the snapshot '"+SnapshotName+"' already existed.")
				}
			} else {
				switch r.Status {
				// Slave are snapshots
				case NotEmpty:
					// Take the uuid of last snapshot on destination
					slaveUUID := r.Error
					// Take the dataset name of snapshot to send to slave
					DatasetName := lib.DatasetName(SnapshotName)
					ds, _ := zfs.GetDataset(DatasetName)
					list, _ := ds.Snapshots()
					count := len(list)

					// Reject the backup snapshot
					for i := 0; i < count; i++ {
						if strings.Contains(list[i].Name, "BACKUP") {
							count--
						}
					}

					// Struct for the flag
					ack := make([]byte, 0)

					// Define variables
					var ds1 *zfs.Dataset
					var ds2 *zfs.Dataset
					var send bool
					index := -1

					// Check if the uuid received exists
					for i := 0; i < count; i++ {
						uuid := lib.SearchUUID(list[i].Name)
						if uuid == slaveUUID {
							index = i
							break
						} else {
							continue
						}
					}

					// Choose the correct option
					if index == (count-1) {
						ack = nil
						ack = strconv.AppendInt(ack, MostActual, 10)
						send = false
					} else if index < (count-1) && index != -1 {
						snap1 := lib.SearchName(slaveUUID)
						ds1, _ = zfs.GetDataset(snap1)
						ds2, _ = zfs.GetDataset(SnapshotName)
						ack = nil
						ack = strconv.AppendInt(ack, Incremental, 10)
						send = true
					} else {
						ack = nil
						ack = strconv.AppendInt(ack, Zerror, 10)
						send = false
					}

					// New connection with the slave node
					conn2ToSlave, _ := net.Dial("tcp", d.Destination+":7744")
					// Send the flag to destination
					conn2ToSlave.Write(ack)

					if send == true {
						// Send the incremental stream
						zfs.SendSnapshotIncremental(conn2ToSlave, ds1, ds2, true, zfs.IncrementalPackage)
						conn2ToSlave.Close()
					} else {
						conn2ToSlave.Close()
					}

					// Reconnection to get ZFSResponse
					l3, _ := net.Listen("tcp", ":7755")
					defer l3.Close()
					fmt.Println("[Agent:7755] Receiving response from slave...")

					conn3Agent, _ := l3.Accept()
					var r2 ZFSResponseFromSlave
					response2, err := bufio.NewReader(conn3Agent).ReadBytes('\x0A')
					if err != nil {
						w.Err("[ERROR] an error has occurred while reading from the socket.")
						break
					}
					err = json.Unmarshal(response2, &r2)
					if err != nil {
						w.Err("[ERROR] it was impossible to parse the JSON struct from the socket.")
						break
					}
					if r2.IsSuccess == true {
						switch r2.Status {
						case WasWritten:
							w.Info("[INFO] the snapshot '"+SnapshotName+"' has been sent.")
						case NothingToDo:
							w.Info("[INFO] the node '"+d.Destination+"' has a snapshot more actual.")
						}
					} else {
						switch r2.Status {
						case Zerror:
							w.Err(""+r2.Error+"")
						}
					}

				}
			}

		// Case: dataset does not exit on destination or it's empty
//		// *** Use -R option ? No option ? ***
		case DatasetFalse:
			// Send snapshot to slave
			ds.SendSnapshot(connToSlave, zfs.ReplicationStream)
			connToSlave.Close()

			// Reconnection to get ZFSResponse
			l2, _ := net.Listen("tcp", ":7733")
			defer l2.Close()
			fmt.Println("[AGENT:7733] Receiving response from slave...")

			conn2Agent, _ := l2.Accept()

			var r ZFSResponseFromSlave
			response, err := bufio.NewReader(conn2Agent).ReadBytes('\x0A')
			if err != nil {
				w.Err("[ERROR] an error has occurred while reading from the socket.")
				break
			}
			err = json.Unmarshal(response, &r)
			if err != nil {
				w.Err("[ERROR] it was impossible to parse the JSON struct from the socket.")
				break
			}
			if r.IsSuccess == true {
				switch r.Status {
				// Snapshot written
				case WasWritten:
					w.Info("[INFO] the snapshot '"+SnapshotName+"' has been sent.")
				}
			} else {
				switch r.Status {
				// ZFS error
				case Zerror:
					w.Err(""+r.Error+"")
				}
			}

		// Network error
		default:
			w.Err("[ERROR] it was not possible to receive any response from '"+d.Destination+"'.")
			break
		}

	// Destroy snapshot
	case "destroy_snapshot":
		// Check if the uuid of snapshot has been sent
		if d.SnapshotUUID == "" || d.SnapshotName == "" {
			w.Err("[ERROR] inconsistant data structure in ZFS order.")
			break
		}
		// Search the snapshot name from its uuid
		SnapshotName := lib.SearchName(d.SnapshotUUID)

		// Check if something was written
		snap, _ := zfs.GetDataset(SnapshotName)
		written := snap.Written

		if d.SkipIfNotWritten == false || d.SkipIfNotWritten == true && written > 0 {
			// Check if the snapshot was renamed
			if d.SkipIfRenamed == true && d.SnapshotName != SnapshotName {
				w.Info("[INFO] the snapshot '"+d.SnapshotName+"' was renamed to '"+SnapshotName+"'.")
			} else {
				// Read JSON configuration file
				j, _, _ := config.JSON()

				// Call to function CommandOrder for create the snapshot
				destroy, clone := lib.DestroyOrder(j, SnapshotName)

				// Print the name of snapshot destroyed (using its uuid)
				if destroy == true && d.SnapshotName != SnapshotName {
					w.Info("[INFO] the snapshot '"+d.SnapshotName+"' (renamed as "+SnapshotName+") has been destroyed.")
				} else if destroy == true && d.SnapshotName == SnapshotName {
					w.Info("[INFO] the snapshot '"+d.SnapshotName+"' has been destroyed.")
				} else {
					w.Info("[INFO] the snapshot '"+d.SnapshotName+"' has dependent clones: '"+clone+"'.")
				}
			}
		}

	default:
		w.Err("[ERROR] the action '"+d.Action+"' is not supported.")
		break
	}
	stop := false
	return stop
}
