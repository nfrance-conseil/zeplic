// Package director contains: actions.go - agent.go - director.go - slave.go
//
// Agent executes the orders received from director
//
package director

import (
	"bufio"
	"encoding/json"
	"net"
	"strconv"

	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

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
func HandleRequestAgent (ConnAgent net.Conn) {
	// Unmarshal orders from director
	var d ZFSDirectorsOrder
	director, err := bufio.NewReader(ConnAgent).ReadBytes('\x0A')
	if err != nil {
		w.Err("[ERROR > director/agent.go:43] an error has occurred while reading from the socket.")
	} else {
		err = json.Unmarshal(director, &d)
		if err != nil {
			w.Err("[ERROR > director/agent.go:47] it was not possible to parse the JSON struct from the socket.")
		} else {
			if d.OrderUUID == "" || d.Action == "" || (d.Action == "send_snapshot" && d.Destination == "") {
				w.Err("[ERROR > director/agent.go:51] inconsistant data structure in ZFS order.")
			} else {

				// Switch for action order
				switch d.Action {

				// Create a new snapshot
				case "take_snapshot":
					// Checking required information
					if d.DestDataset == "" || d.SnapshotName == "" {
						w.Err("[ERROR > director/agent.go:61] inconsistant data structure in ZFS order.")
						break
					} else {
						// Call to function TakeOrder for create the snapshot
						go lib.TakeOrder(d.DestDataset, d.SnapshotName, d.SkipIfNotWritten)
					}

				// Send snapshot to d.Destination
				case "send_snapshot":
					// Checking required information
					if len(d.SnapshotUUID) == 0 || d.Destination == "" || d.DestDataset == "" {
						w.Err("[ERROR > director/agent.go:72] inconsistant data structure in ZFS order.")
						break
					}
					// Search the snapshot name from its uuid
					SnapshotName := lib.SearchName(d.SnapshotUUID[0])

					// Get the snapshot
					snap, err := zfs.GetDataset(SnapshotName)
					if err != nil {
						w.Err("[ERROR > director/agent.go:80] it was not possible to get the snapshot '"+SnapshotName+"'.")
						break
					}

					// Struct for ZFS orders to slave
					ZFSOrderToSlave := ZFSOrderToSlave{lib.Host(),d.OrderUUID,d.SnapshotUUID[0],SnapshotName,d.DestDataset}
					ots, err := json.Marshal(ZFSOrderToSlave)
					if err != nil {
						w.Err("[ERROR > director/agent.go:87] it was impossible to encode the JSON struct.")
						break
					}

					// Create a new connection with the destination
					ConnToSlave, err := net.Dial("tcp", d.Destination+":7722")
					if err != nil {
						w.Err("[ERROR > director/agent.go:95] it was not possible to connect with '"+d.Destination+"'.")
						break
					}
					ConnToSlave.Write([]byte(ots))
					ConnToSlave.Write([]byte("\n"))

					// Read from destinantion if Dataset exists
					buff := bufio.NewReader(ConnToSlave)
					n, err := buff.ReadByte()
					if err != nil {
						w.Err("[ERROR > director/agent.go:105] it was not possible to read the 'dataset byte'.")
						break
					}
					m := string(n)
					dsExist, _ := strconv.Atoi(m)

					switch dsExist {

					// Case: dataset exist on destination
					case DatasetTrue:
						// Check uuid of last snapshot on destination
						ConnToSlave.Close()

						// Reconnection to get ZFSResponse
						l2, err := net.Listen("tcp", ":7733")
						if err != nil {
							w.Err("[ERROR > director/agent.go:121] it was not possible to listen on port '7733'.")
							break
						}
						defer l2.Close()

						Conn2Agent, err := l2.Accept()
						if err != nil {
							w.Err("[ERROR > director/agent.go:128] it was not possible to accept the connection.")
							break
						}

						var r ZFSResponseFromSlave
						response, err := bufio.NewReader(Conn2Agent).ReadBytes('\x0A')
						if err != nil {
							w.Err("[ERROR > director/agent.go:135] an error has occurred while reading from the socket.")
							break
						}
						err = json.Unmarshal(response, &r)
						if err != nil {
							w.Err("[ERROR > director/agent.go:140] it was impossible to parse the JSON struct from the socket.")
							break
						}

						if r.IsSuccess == true {
							switch r.Status {
							// Slave are snapshots
							case NotEmpty:
								// Reconnection to get ZFSResponse
								l3, err := net.Listen("tcp", ":7744")
								if err != nil {
									w.Err("[ERROR > director/agent.go:151] it was not possible to listen on port '7744'.")
									break
								}
								defer l3.Close()

								Conn3Agent, err := l3.Accept()
								if err != nil {
									w.Err("[ERROR > director/agent.go:158] it was not possible to accept the connection.")
									break
								}

								var r2 ZFSListUUIDsFromSlave
								response2, err := bufio.NewReader(Conn3Agent).ReadBytes('\x0A')
								if err != nil {
									w.Err("[ERROR > director/agent.go:165] an error has occurred while reading from the socket.")
									break
								}
								err = json.Unmarshal(response2, &r2)
								if err != nil {
									w.Err("[ERROR > director/agent.go:170] it was impossible to parse the JSON struct from the socket.")
									break
								}

								// Search the snapshots in source dataset
								ack, send, incremental, ds1, ds2 := lib.Delivery(r2.MapUUID, SnapshotName)

								// New connection with the slave node
								Conn2ToSlave, err := net.Dial("tcp", d.Destination+":7755")
								if err != nil {
									w.Err("[ERROR > director/agent.go:180] it was not possible to listen on port '7755'.")
									break
								}

								// Send the flag to destination
								Conn2ToSlave.Write(ack)

								if send == true && incremental == false {
									// Send the snapshot
									snap, err := zfs.GetDataset(SnapshotName)
									if err != nil {
										w.Err("[ERROR > director/agent.go:191] it was not possible to get the snapshot '"+SnapshotName+"'.")
										break
									}
									snap.SendSnapshot(Conn2ToSlave, zfs.ReplicationStream)
									Conn2ToSlave.Close()
								} else if send == true && incremental == true {
									// Send the incremental stream
									zfs.SendSnapshotIncremental(Conn2ToSlave, ds1, ds2, true, zfs.IncrementalPackage)
									Conn2ToSlave.Close()
								} else {
									Conn2ToSlave.Close()
								}

								// Reconnection to get ZFSResponse
								l4, err := net.Listen("tcp", ":7766")
								if err != nil {
									w.Err("[ERROR > director/agent.go:207] it was not possible to listen on port '7766'.")
									break
								}
								defer l4.Close()

								Conn4Agent, _ := l4.Accept()
								var r3 ZFSResponseFromSlave
								response3, err := bufio.NewReader(Conn4Agent).ReadBytes('\x0A')
								if err != nil {
									w.Err("[ERROR > director/agent.go:216] an error has occurred while reading from the socket.")
									break
								}
								err = json.Unmarshal(response3, &r3)
								if err != nil {
									w.Err("[ERROR > director/agent.go:221] it was impossible to parse the JSON struct from the socket.")
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
						snap.SendSnapshot(ConnToSlave, zfs.ReplicationStream)
						ConnToSlave.Close()

						// Reconnection to get ZFSResponse
						l2, err := net.Listen("tcp", ":7733")
						if err != nil {
							w.Err("[ERROR > director/agent.go:255] it was not possible to listen on port '7733'.")
							break
						}
						defer l2.Close()

						Conn2Agent, err := l2.Accept()
						if err != nil {
							w.Err("[ERROR > director/agent.go:262] it was not possible to accept the connection.")
							break
						}

						var r ZFSResponseFromSlave
						response, err := bufio.NewReader(Conn2Agent).ReadBytes('\x0A')
						if err != nil {
							w.Err("[ERROR > director/agent.go:269] an error has occurred while reading from the socket.")
							break
						}
						err = json.Unmarshal(response, &r)
						if err != nil {
							w.Err("[ERROR > director/agent.go:274] it was impossible to parse the JSON struct from the socket.")
							break
						}

						// Snapshot sent?
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
						w.Err("[ERROR > director/agent.go:113] it was not possible to receive any response from '"+d.Destination+"'.")
						break
					}

				// Destroy snapshot
				case "destroy_snapshot":
					// Checking required information
					if len(d.SnapshotUUID) == 0 {
						w.Err("[ERROR > director/agent.go:316] inconsistant data structure in ZFS order.")
						break
					} else {
						// Call to function DestroyOrder
						go lib.DestroyOrder(d.SnapshotUUID, d.SkipIfRenamed, d.SkipIfCloned)
					}

				// Resync
				case "kv_resync":
					go lib.Update(d.Destination, d.DestDataset)

				default:
					w.Err("[ERROR > director/agent.go:56] the action '"+d.Action+"' is not supported.")
					break
				}
			}
		}
	}
}
