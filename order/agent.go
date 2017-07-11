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
	SkipIfCloned	 bool	// should I delete a snapshot if it was cloned
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

// ZFSListUUIDsFromSlave is the struct to receive the list of uuids in DestDataset
type ZFSListUUIDsFromSlave struct {
	MapUUID		 []string // map of uuids
}

// HandleRequestAgent incoming requests from director
func HandleRequestAgent (connAgent net.Conn) bool {
	// Start syslog system service
	w := config.LogBook()

	// Resolve hostname
	hostname, err := os.Hostname()
	if err != nil {
		w.Err("[ERROR > order/agent.go:63] it was not possible to resolve the hostname.")
	}

	// Unmarshal orders from director
	var d ZFSOrderFromDirector
	director, err := ioutil.ReadAll(connAgent)
	if err != nil {
		w.Err("[ERROR > order/agent.go:70] an error has occurred while reading from the socket.")
	}
	err = json.Unmarshal(director, &d)
	if err != nil {
		w.Err("[ERROR > order/agent.go:74] it was not possible to parse the JSON struct from the socket.")
	}
	if d.OrderUUID == "" || d.Action == "" || (d.Action == "send_snapshot" && d.Destination == "") {
		w.Err("[ERROR > order/agent.go:78] inconsistant data structure in ZFS order.")
	}

	// Switch for action order
	switch d.Action {

	// Create a new snapshot
	case "take_snapshot":
		// Check if the DestDataset exists and it is enable
		if d.DestDataset == "" {
			w.Err("[ERROR > order/agent.go:88] inconsistant data structure in ZFS order.")
			break
		}

		// Read JSON configuration file
		j, _, _ := config.JSON()

		// Call to function CommandOrder for create the snapshot
		lib.TakeOrder(j, d.DestDataset, d.SkipIfNotWritten)

	// Send snapshot to d.Destination
	case "send_snapshot":
		// Checking required information
		if d.SnapshotUUID == "" || d.Destination == "" || d.DestDataset == "" {
			w.Err("[ERROR > order/agent.go:102] inconsistant data structure in ZFS order.")
			break
		}
		// Search the snapshot name from its uuid
		SnapshotName := lib.SearchName(d.SnapshotUUID)

		// Check if something was written
		snap, err := zfs.GetDataset(SnapshotName)
		if err != nil {
			w.Err("[ERROR > order/agent.go:110] it was not possible to get the snapshot '"+SnapshotName+"'.")
		}
		written := snap.Written

		if d.SkipIfNotWritten == false || (d.SkipIfNotWritten == true && written > 0) {

			// Take the snapshot
			ds, err := zfs.GetDataset(SnapshotName)
			if err != nil {
				w.Err("[ERROR > order/agent.go:119] it was not possible to get the snapshot '"+SnapshotName+"'.")
			}

			// Create a new connection with the destination
			connToSlave, err := net.Dial("tcp", d.Destination+":7722")
			if err != nil {
				w.Err("[ERROR > order/agent.go:125] it was not possible to connect with '"+d.Destination+"'.")
			}

			// Struct for ZFS orders to slave
			ZFSOrderToSlave := ZFSOrderToSlave{hostname,d.OrderUUID,d.SnapshotUUID,SnapshotName,d.DestDataset}
			ots, err := json.Marshal(ZFSOrderToSlave)
			if err != nil {
				w.Err("[ERROR > order/agent.go:132] it was impossible to encode the JSON struct.")
			}
			connToSlave.Write([]byte(ots))
			connToSlave.Write([]byte("\n"))

			// Read from destinantion if Dataset exists
			buff := bufio.NewReader(connToSlave)
			n, err := buff.ReadByte()
			if err != nil {
				w.Err("[ERROR > order/agent.go:141] it was not possible to read the 'dataset byte'.")
			}
			m := string(n)
			dsExist, err := strconv.Atoi(m)
			if err != nil {
				w.Err("[ERROR > order/agent.go:145] it was not possible to convert the string '"+m+"' to integer.")
			}

			switch dsExist {

			// Case: dataset exist on destination
			case DatasetTrue:
				// Check uuid of last snapshot on destination
				connToSlave.Close()

				// Reconnection to get ZFSResponse
				l2, err := net.Listen("tcp", ":7733")
				if err != nil {
					w.Err("[ERROR > order/agent.go:159] it was not possible to listen on port '7733'.")
				}
				defer l2.Close()
				fmt.Println("[AGENT:7733] Receiving response from slave...")

				conn2Agent, err := l2.Accept()
				if err != nil {
					w.Err("[ERROR > order/agent.go:166] it was not possible to accept the connection.")
				}

				var r ZFSResponseFromSlave
				response, err := bufio.NewReader(conn2Agent).ReadBytes('\x0A')
				if err != nil {
					w.Err("[ERROR > order/agent.go:172] an error has occurred while reading from the socket.")
					break
				}
				err = json.Unmarshal(response, &r)
				if err != nil {
					w.Err("[ERROR > order/agent.go:177] it was impossible to parse the JSON struct from the socket.")
					break
				}
				if r.IsSuccess == true {
					switch r.Status {
					// Slave are snapshots
					case NotEmpty:
						// Reconnection to get ZFSResponse
						l3, err := net.Listen("tcp", ":7744")
						if err != nil {
							w.Err("[ERROR > order/agent.go:187] it was not possible to listen on port '7744'.")
						}
						defer l3.Close()
						fmt.Println("[Agent:7744] Receiving list of uuids in DestDataset...")
						conn3Agent, err := l3.Accept()
						if err != nil {
							w.Err("[ERROR > order/agent.go:193] it was not possible to accept the connection.")
						}
						var r2 ZFSListUUIDsFromSlave
						response2, err := bufio.NewReader(conn3Agent).ReadBytes('\x0A')
						if err != nil {
							w.Err("[ERROR > order/agent.go:198] an error has occurred while reading from the socket.")
							break
						}
						err = json.Unmarshal(response2, &r2)
						if err != nil {
							w.Err("[ERROR > order/agent.go:203] it was impossible to parse the JSON struct from the socket.")
							break
						}

						// Search the snapshots in source dataset
						slice := len(r2.MapUUID)
						var found string
						index := -1
						for i := slice-1; i > -1; i-- {
							name := lib.SearchName(r2.MapUUID[i])
							_, err := zfs.GetDataset(name)
							if err != nil {
								continue
							} else {
								found = name
								index = i
								break
							}
						}

						// Define all variables
						var send bool
						var incremental bool
						var ds1 *zfs.Dataset
						var ds2 *zfs.Dataset

						// Struct for the flag
						ack := make([]byte, 0)

						// Choose the correct option
						if found == "" {
							ack = nil
							ack = strconv.AppendInt(ack, Zerror, 10)
							send = true
							incremental = false
						} else if found == SnapshotName {
							ack = nil
							ack = strconv.AppendInt(ack, NothingToDo, 10)
							send = false
							incremental = false
						} else if found != "" && found != SnapshotName {
							dataset := lib.DatasetName(SnapshotName)
							ds, err := zfs.GetDataset(dataset)
							if err != nil {
								w.Err("[ERROR > order/agent.go:247] it was not possible to get the dataset '"+dataset+"'.")
							}
							list, err := ds.Snapshots()
							if err != nil {
								w.Err("[ERROR > order/agent.go:251] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
							}
							count := len(list)

							// Search the index of snapshot to send
							_, amount := lib.RealList(count, list, dataset)
							var number int
							for i := 0; i < amount; i++ {
								take := list[i].Name
								if take == SnapshotName {
									number = i
									break
								} else {
									continue
								}
							}
							if index < number {
								ds1, err = zfs.GetDataset(found)
								if err != nil {
									w.Err("[ERROR > order/agent.go:270] it was not possible to get the snapshot '"+found+"'.")
								}
								ds2, err = zfs.GetDataset(SnapshotName)
								if err != nil {
									w.Err("[ERROR > order/agent.go:274] it was not possible to get the snapshot '"+SnapshotName+"'.")
								}
								ack = nil
								ack = strconv.AppendInt(ack, Incremental, 10)
								send = true
								incremental = true
							} else {
								ack = nil
								ack = strconv.AppendInt(ack, MostActual, 10)
								send = false
								incremental = false
							}
						}

						// New connection with the slave node
						conn2ToSlave, err := net.Dial("tcp", d.Destination+":7755")
						if err != nil {
							w.Err("[ERROR > order/agent.go:291] it was not possible to listen on port '7755'.")
						}
						// Send the flag to destination
						conn2ToSlave.Write(ack)

						if send == true && incremental == false {
							// Send the snapshot
							ds, err := zfs.GetDataset(SnapshotName)
							if err != nil {
								w.Err("[ERROR > order/agent.go:300] it was not possible to get the snapshot '"+SnapshotName+"'.")
							}
							ds.SendSnapshot(conn2ToSlave, zfs.ReplicationStream)
							conn2ToSlave.Close()
						} else if send == true && incremental == true {
							// Send the incremental stream
							zfs.SendSnapshotIncremental(conn2ToSlave, ds1, ds2, true, zfs.IncrementalPackage)
							conn2ToSlave.Close()
						} else {
							conn2ToSlave.Close()
						}

						// Reconnection to get ZFSResponse
						l4, err := net.Listen("tcp", ":7766")
						if err != nil {
							w.Err("[ERROR > order/agent.go:315] it was not possible to listen on port '7766'.")
						}
						defer l4.Close()
						fmt.Println("[Agent:7766] Receiving response from slave...")

						conn4Agent, _ := l4.Accept()
						var r3 ZFSResponseFromSlave
						response3, err := bufio.NewReader(conn4Agent).ReadBytes('\x0A')
						if err != nil {
							w.Err("[ERROR > order/agent.go:324] an error has occurred while reading from the socket.")
							break
						}
						err = json.Unmarshal(response3, &r3)
						if err != nil {
							w.Err("[ERROR > order/agent.go:329] it was impossible to parse the JSON struct from the socket.")
							break
						}
						if r3.IsSuccess == true {
							switch r3.Status {
							case WasWritten:
								w.Info("[INFO] the snapshot '"+SnapshotName+"' has been sent.")
							case MostActual:
								w.Info("[INFO] the node '"+d.Destination+"' has a snapshot more actual.")
							case NothingToDo:
								w.Info("[INFO] the snapshot '"+SnapshotName+"' already exists on '"+d.Destination+"'.")
							case WasRenamed:
								w.Info("[INFO] the snapshot '"+SnapshotName+"' was renamed on '"+d.Destination+"'.")
							}
						} else {
							switch r3.Status {
							case Zerror:
								w.Err(""+r3.Error+"")
							}
						}
					}
				}

			// Case: dataset does not exit on destination or it's empty
			case DatasetFalse:
				// Send snapshot to slave
				ds.SendSnapshot(connToSlave, zfs.ReplicationStream)
				connToSlave.Close()

				// Reconnection to get ZFSResponse
				l2, err := net.Listen("tcp", ":7733")
				if err != nil {
					w.Err("[ERROR > order/agent.go:361] it was not possible to listen on port '7733'.")
				}
				defer l2.Close()
				fmt.Println("[AGENT:7733] Receiving response from slave...")

				conn2Agent, err := l2.Accept()
				if err != nil {
					w.Err("[ERROR > order/agent.go:368] it was not possible to accept the connection.")
				}

				var r ZFSResponseFromSlave
				response, err := bufio.NewReader(conn2Agent).ReadBytes('\x0A')
				if err != nil {
					w.Err("[ERROR > order/agent.go:374] an error has occurred while reading from the socket.")
					break
				}
				err = json.Unmarshal(response, &r)
				if err != nil {
					w.Err("[ERROR > order/agent.go:379] it was impossible to parse the JSON struct from the socket.")
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

			// Case: dataset is disabled on destination
			case DatasetDisable:
				w.Notice("[NOTICE] the dataset '"+d.DestDataset+"' on destination is disabled.")

			// Case: dataset is not a docker dataset
			case DatasetDocker:
				w.Notice("[NOTICE] the dataset '"+d.DestDataset+"' is not a docker dataset.")

			// Case: dataset is not configured on destination
			case DatasetNotConf:
				w.Notice("[NOTICE] the dataset '"+d.DestDataset+"' on destination is not configured.")

			// Network error
			default:
				w.Err("[ERROR > order/agent.go:411] it was not possible to receive any response from '"+d.Destination+"'.")
				break
			}
		} else {
			// Create a new connection with the destination
			connToSlave, err := net.Dial("tcp", d.Destination+":7722")
			if err != nil {
				w.Err("[ERROR > order/agent.go:417] it was not possible to connect with '"+d.Destination+"'.")
			}

			// Struct for ZFS orders to slave
			ZFSOrderToSlave := ZFSOrderToSlave{hostname,"NotWritten","","",""}
			ots, err := json.Marshal(ZFSOrderToSlave)
			if err != nil {
				w.Err("[ERROR > order/agent.go:424] it was impossible to encode the JSON struct.")
			}
			connToSlave.Write([]byte(ots))
			connToSlave.Write([]byte("\n"))
			connToSlave.Close()
		}

	// Destroy snapshot
	case "destroy_snapshot":
		// Checking required information
		if d.SnapshotUUID == "" || d.SnapshotName == "" {
			w.Err("[ERROR > order/agent.go:436] inconsistant data structure in ZFS order.")
			break
		}
		// Search the snapshot name from its uuid
		SnapshotName := lib.SearchName(d.SnapshotUUID)

		// Check if something was written
		snap, err := zfs.GetDataset(SnapshotName)
		if err != nil {
			w.Err("[ERROR > order/agent.go:444] it was not possible to get the snapshot '"+SnapshotName+"'.")
		}
		written := snap.Written

		if d.SkipIfNotWritten == false || d.SkipIfNotWritten == true && written > 0 {
			// Check if the snapshot was renamed
			if d.SkipIfRenamed == true && d.SnapshotName != SnapshotName {
				w.Info("[INFO] the snapshot '"+d.SnapshotName+"' was renamed to '"+SnapshotName+"'.")
			} else {
				// Call to function DestroyOrder for destroy the snapshot
				destroy, clone := lib.DestroyOrder(SnapshotName, d.SkipIfCloned)

				// Print the name of snapshot destroyed (using its uuid)
				if destroy == true && d.SnapshotName != SnapshotName {
					w.Info("[INFO] the snapshot '"+d.SnapshotName+"' (renamed as "+SnapshotName+") has been destroyed.")
				} else if destroy == true && d.SnapshotName == SnapshotName {
					w.Info("[INFO] the snapshot '"+d.SnapshotName+"' has been destroyed.")
				} else if destroy == false && clone != "" {
					w.Info("[INFO] the snapshot '"+d.SnapshotName+"' has dependent clones: '"+clone+"'.")
				}
			}
		}

	default:
		w.Err("[ERROR > order/agent.go:469] the action '"+d.Action+"' is not supported.")
		break
	}
	stop := false
	return stop
}
