// Package director contains: agent.go - !director.go - slave.go
//
// Agent executes the orders received from director
//
package director

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/mistifyio/go-zfs"
)

// Struct for ZFS orders from director
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

// Struc for ZFS orders to slave
type ZFSOrderToSlave struct {
	Hostname	string `json:"Source"`
	OrderUUID	string `json:"OrderUUID"`
	SnapshotUUID	string `json:"SnapshotUUID"`
	SnapshotName	string `json:"SnapshotName"`
	DestDataset	string `json:"DestDataset"`
}

// Struct for ZFS response from slave
type ZFSResponseFromSlave struct {
	OrderUUID    string  // reference to a valid order
	IsSuccess    bool    // true or false
	Status	     int64   // 
	Error	     string  // error string if needed
}

// Handle incoming requests from director
func HandleRequestAgent (connAgent net.Conn) bool {
	// Resolve hostname
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("[ERROR] it was not possible to resolve the hostname.\n")
	}

	// Unmarshal orders from director
	var d ZFSOrderFromDirector
	director, err := ioutil.ReadAll(connAgent)
	if err != nil {
		fmt.Printf("[ERROR] an error has occurred while reading from the socket.\n")
	}
	err = json.Unmarshal(director, &d)
	if err != nil {
		fmt.Printf("[ERROR] it was impossible to parse the JSON struct from the socket.\n")
	}
	if d.OrderUUID == "" || d.Action == "" || (d.Action == "send_snapshot" && d.Destination == "") {
		fmt.Printf("[ERROR] inconsistant data structure in ZFS order.\n")
	}

	// Switch for action order
	switch d.Action {

	// Create a new snapshot
	case "take_snapshot":
//		// *** RESOLVE THIS POINT! CONFLICT WITH JSON CONFIG FILE *** DOES THE DATASET EXISTS? IS IT ENABLE? FROM DATASET NAME > TAKE SNAPSHOT NAME? ***
		if d.DestDataset == "" {
			fmt.Printf("[ERROR] inconsistant data structure in ZFS order.\n")
			break
		}
		// Get dataset from d.DestDataset
		ds, err := zfs.GetDataset(d.DestDataset)
		// Create dataset if it does not exist
		if err != nil {
			_, err := zfs.CreateFilesystem(d.DestDataset, nil)
			if err != nil {
				fmt.Printf("[ERROR] it was not possible to create the dataset '%s'\n.", d.DestDataset)
				break
			} else {
				fmt.Printf("[INFO] the dataset '%s' has been created.\n", d.DestDataset)
			}
			ds, _ = zfs.GetDataset(d.DestDataset)
		}
		// Name of snapshot using SnapName function
		SnapshotName := lib.SnapName("DIRECTOR")
		// Create the snapshot
		ds.Snapshot(SnapshotName, false)
		// Get the last snapshot created
		list, _ := zfs.Snapshots(d.DestDataset)
		count := len(list)
		SnapshotCreated := list[count-1].Name
		// Set it an uuid
		go lib.UUID(SnapshotCreated)
		// Print the name of last snapshot created (it has an uuid)
		fmt.Printf("[INFO] the snapshot '%s' has been created.\n", SnapshotCreated)

	// Send snapshot to d.Destination
	case "send_snapshot":
		// Checking required information
		if d.SnapshotUUID == "" || d.Destination == "" || d.DestDataset == "" {
			fmt.Printf("[ERROR] inconsistant data structure in ZFS order.\n")
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
			fmt.Printf("[ERROR] it was impossible to encode the JSON struct.\n")
		}
		connToSlave.Write([]byte(ots))
		connToSlave.Write([]byte("\n"))

		// Read from destinantion if Dataset exists
		buff := bufio.NewReader(connToSlave)
		n, _ := buff.ReadByte()
		dsExist, _ := strconv.Atoi(string(n))

		switch dsExist {

		// Case: dataset exist on destination
		case DATASET_TRUE:
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
				fmt.Printf("[ERROR] an error has occurred while reading from the socket.\n")
				break
			}
			err = json.Unmarshal(response, &r)
			if err != nil {
				fmt.Printf("[ERROR] it was impossible to parse the JSON struct from the socket.\n")
				break
			}
			if r.IsSuccess == true {
				switch r.Status {
				// Snapshot renamed
				case WAS_RENAMED:
					fmt.Printf("[INFO] the snapshot '%s' has been renamed to '%s'.\n", SnapshotName, r.Error)
				// Nothing to do
				case NOTHING_TO_DO:
					fmt.Printf("[INFO] the snapshot '%s' already existed.\n", SnapshotName)
				}
			} else {
				switch r.Status {
				// Slave are snapshots
				case NOT_EMPTY:
					// Take the uuid of last snapshot on destination
					slaveUUID := r.Error
					// Take the dataset name of snapshot to send to slave
					DatasetName := lib.DatasetName(SnapshotName)
					list, _ := zfs.Snapshots(DatasetName)
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
						ack = strconv.AppendInt(ack, MOST_ACTUAL, 10)
						send = false
					} else if index < (count-1) && index != -1 {
						snap1 := lib.SearchName(slaveUUID)
						ds1, _ = zfs.GetDataset(snap1)
						ds2, _ = zfs.GetDataset(SnapshotName)
						ack = nil
						ack = strconv.AppendInt(ack, INCREMENTAL, 10)
						send = true
					} else {
						ack = nil
						ack = strconv.AppendInt(ack, ZFS_ERROR, 10)
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
						fmt.Printf("[ERROR] an error has occurred while reading from the socket.\n")
						break
					}
					err = json.Unmarshal(response2, &r2)
					if err != nil {
						fmt.Printf("[ERROR] it was impossible to parse the JSON struct from the socket.\n")
						break
					}
					if r2.IsSuccess == true {
						switch r2.Status {
						case WAS_WRITTEN:
							fmt.Printf("[INFO] the snapshot '%s' has been sent.\n", SnapshotName)
						case NOTHING_TO_DO:
							fmt.Printf("[INFO] the node '%s' has a snapshot more actual.\n", d.Destination)
						}
					} else {
						switch r2.Status {
						case ZFS_ERROR:
							fmt.Printf("%s\n", r2.Error)
						}
					}

				}
			}

		// Case: dataset does not exit on destination or it's empty
//		// *** Use -R option ? No option ? ***
		case DATASET_FALSE:
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
				fmt.Printf("[ERROR] an error has occurred while reading from the socket.\n")
				break
			}
			err = json.Unmarshal(response, &r)
			if err != nil {
				fmt.Printf("[ERROR] it was impossible to parse the JSON struct from the socket.\n")
				break
			}
			if r.IsSuccess == true {
				switch r.Status {
				// Snapshot renamed
				case WAS_RENAMED:
					fmt.Printf("[INFO] the snapshot '%s' has been renamed to ''.\n", SnapshotName)
				// Snapshot written
				case WAS_WRITTEN:
					fmt.Printf("[INFO] the snapshot '%s' has been sent.\n", SnapshotName)
				// Nothing to do
				case NOTHING_TO_DO:
					fmt.Printf("[INFO] the snapshot '%s' already existed.\n", SnapshotName)
				}
			} else {
				switch r.Status {
				// ZFS error
				case ZFS_ERROR:
					fmt.Printf("%s\n", r.Error)
				}
			}

		// Network error
		default:
			fmt.Printf("[ERROR] it was not possible to receive any response from '%s'.\n", d.Destination)
			break
		}

	// Destroy snapshot
	case "destroy_snapshot":
		// Check if the uuid of snapshot has been sent
		if d.SnapshotUUID == "" || d.SnapshotName == "" {
			fmt.Printf("[ERROR] inconsistant data structure in ZFS order.\n")
			break
		}
		// Search the snapshot name from its uuid
		SnapshotName := lib.SearchName(d.SnapshotUUID)

		// Check if the snapshot was renamed
		if d.SkipIfRenamed == true && d.SnapshotName != SnapshotName {
			fmt.Printf("[INFO] the snapshot '%s' was renamed to '%s'.\n", d.SnapshotName, SnapshotName)
		} else {
			// Take the snapshot...
			ds, _ := zfs.GetDataset(SnapshotName)
			// ... and destroy it
			ds.Destroy(zfs.DestroyDefault)
			// Print the name of snapshot destroyed (using its uuid)
			if d.SnapshotName != SnapshotName {
				fmt.Printf("[INFO] the snapshot '%s' (and renamed to '%s') has been destroyed.\n", d.SnapshotName, SnapshotName) 
			} else {
				fmt.Printf("[INFO] the snapshot '%s' has been destroyed.\n", d.SnapshotName)
			}
		}
	default:
		fmt.Printf("[ERROR] the action '%s' is not supported.\n", d.Action)
		break
	}
	stop := false
	return stop
}
