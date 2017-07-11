// Package order contains: agent.go - !director.go - slave.go
//
// Slave receives a snapshot from agent
//
package order

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
)

// ZFSOrderFromAgent is the struct for ZFS orders from agent
type ZFSOrderFromAgent struct {
	Source		string // hostname of agent
	OrderUUID	string // mandatory
	SnapshotUUID	string // uuid of snapshot received
	SnapshotName	string // name of snapshot received
	DestDataset	string // dataset for receive
}

// ZFSResponseToAgent is the struct for ZFS response to agent
type ZFSResponseToAgent struct {
	OrderUUID	string	`json:"OrderUUID"`
	IsSuccess	bool	`json:"IsSuccess"`
	Status		int64	`json:"Status"`
	Error		string	`json:"Error"`
}

// ZFSListUUIDsToAgent is the struct to send the list of uuids in dataset
type ZFSListUUIDsToAgent struct {
	MapUUID		[]string `json:"MapUUID"`
}

// HandleRequestSlave incoming requests from agent
func HandleRequestSlave (connSlave net.Conn) bool {
	// Start syslog system service
	w := config.LogBook()

	// Resolve hostname
	hostname, err := os.Hostname()
	if err != nil {
		w.Err("[ERROR > order/slave.go:48] it was not possible to resolve the hostname.")
	}

	// Unmarshal orders from agent
	var a ZFSOrderFromAgent
	agent, err := bufio.NewReader(connSlave).ReadBytes('\x0A')
	if err != nil {
		w.Err("[ERROR > order/slave.go:55] an error has occurred while reading from the socket.")
	}
	err = json.Unmarshal(agent, &a)
	if err != nil {
		w.Err("[ERROR > order/slave.go:59] it was not possible to parse the JSON struct from the socket.")
	}

	if a.OrderUUID != "NotWritten" {
		// Struct for Status constant
		ack := make([]byte, 0)
		// Variable to continue the transmission
		var more bool
		// Variable to receive an incremental stream
		var stream bool
		// Variable to execute the runner
		var runner bool
		// Define list and count
		var list []*zfs.Dataset
		var count int

		// Struct for response
		ResponseToAgent := ZFSResponseToAgent{}

		// Read the JSON configuration file
		j, _, _ := config.JSON()

		// Check if dataset is configured
		index := -1
		for i := 0; i < j; i++ {
			pieces	:= config.Extract(i)
			dataset := pieces[4].(string)

			if dataset == a.DestDataset {
				index = i
				break
			} else {
				continue
			}
		}
		if index > -1 {
			// Extract data of dataset
			pieces  := config.Extract(index)
			enable  := pieces[0].(bool)
			docker  := pieces[1].(bool)
			dataset := pieces[4].(string)

			if dataset == a.DestDataset && enable == true && docker == true {
				// Check if the dataset received exists
				ds, err := zfs.GetDataset(a.DestDataset)

				if err != nil {
					// Status for DestDataset
					ack = nil
					ack = strconv.AppendInt(ack, DatasetFalse, 10)
					connSlave.Write(ack)

					// Receive the snapshot
					_, err := zfs.ReceiveSnapshotRollback(connSlave, a.DestDataset, false)

					// Check for response to agent
					if err != nil {
						Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
						ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
						w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
						more = true
					} else {
						ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
						w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
						more = true
						runner = true
					}
				} else {
					// Get the last snapshot in DestDataset
					list, err := ds.Snapshots()
					if err != nil {
						w.Err("[ERROR > order/slave.go:130] it was not possible to access of snapshots list in dataset '"+a.DestDataset+"'.")
					}
					count = len(list)

					// Get the correct number of snapshots in dataset
					backup, amount := lib.RealList(count, list, a.DestDataset)

					if amount == 0 {
						// Status for DestDataset
						ack = nil
						ack = strconv.AppendInt(ack, DatasetFalse, 10)
						connSlave.Write(ack)

						if backup != amount {
							list, err := ds.Snapshots()
							if err != nil {
								w.Err("[ERROR > order/slave.go:146] it was not possible to access of snapshots list in dataset '"+a.DestDataset+"'.")
							}
							backup := list[0].Name
							snap, err := zfs.GetDataset(backup)
							if err != nil {
								w.Err("[ERROR > order/slave.go:151] it was not possible to get the snapshot '"+backup+"'.")
							}
							snap.Destroy(zfs.DestroyDefault)
							w.Info("[INFO] the backup snapshot '"+backup+"' has been destroyed.")
						}

						// Receive the snapshot
						_, err := zfs.ReceiveSnapshotRollback(connSlave, a.DestDataset, true)

						// Check for response to agent
						if err != nil {
							Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
							ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
							w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
							more = true
						} else {
							ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
							w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
							more = true
							runner = true
						}
					} else {
						// Status for DestDataset
						ack = nil
						ack = strconv.AppendInt(ack, DatasetTrue, 10)
						connSlave.Write(ack)

						// Information to agent where Error field contains the uuid of last snapshot in slave
						ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NotEmpty,""}
						more = true
						stream = true
					}
				}
			} else if dataset == a.DestDataset && enable == false {
				// Status for DestDataset
				ack = nil
				ack = strconv.AppendInt(ack, DatasetDisable, 10)
				connSlave.Write(ack)
				connSlave.Close()
				w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is disabled.")
			} else if dataset == a.DestDataset && docker == false {
				// Status for DestDataset
				ack = nil
				ack = strconv.AppendInt(ack, DatasetDocker, 10)
				connSlave.Write(ack)
				connSlave.Close()
				w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is not a docker dataset.")
			}
		} else {
			// Status for DestDataset
			ack = nil
			ack = strconv.AppendInt(ack, DatasetNotConf, 10)
			connSlave.Write(ack)
			connSlave.Close()
			w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is not configured.")
		}

		// ResponseToAgent
		var rta []byte
		if more == true {
			// Reconnection to send ZFSResponseToAgent
			connToAgent, err := net.Dial("tcp", a.Source+":7733")
			if err != nil {
				w.Err("[ERROR > order/slave.go:214] it was not possible to connect with '"+a.Source+"'.")
			}

			// Marshal response to agent
			rta, err = json.Marshal(ResponseToAgent)
			if err != nil {
				w.Err("[ERROR > order/slave.go:220] it was not possible to encode the JSON struct.")
			} else {
				connToAgent.Write([]byte(rta))
				connToAgent.Write([]byte("\n"))
				connToAgent.Close()
			}
		}

		if stream == true {
			// MapUUID to save the list of uuids
			var MapUUID []string

			// Get the list of snapshots in DestDataset
			ds, err := zfs.GetDataset(a.DestDataset)
			if err != nil {
				w.Err("[ERROR > order/slave.go:235] it was not possible to get the dataset '"+a.DestDataset+"'.")
			}
			list, err = ds.Snapshots()
			if err != nil {
				w.Err("[ERROR > order/slave.go:239] it was not possible to access of snapshots list in dataset '"+a.DestDataset+"'.")
			}
			count = len(list)

			// Get the correct number of snapshots in dataset
			_, amount := lib.RealList(count, list, a.DestDataset)

			// Get the list of all uuids in DestDataset
			for i := 0; i < amount; i++ {
				take := list[i].Name
				snap, _ := zfs.GetDataset(take)
				uuid := lib.SearchUUID(snap)
				MapUUID = append(MapUUID, uuid)
			}
			ListUUIDsToAgent := ZFSListUUIDsToAgent{MapUUID}

			// Send the list of uuids in DestDataset
			conn2ToAgent, err := net.Dial("tcp", a.Source+":7744")
			if err != nil {
				w.Err("[ERROR > order/slave.go:258] it was not possible to connect with '"+a.Source+"'.")
			}

			// Marshal response to agent
			lta, err := json.Marshal(ListUUIDsToAgent)
			if err != nil {
				w.Err("[ERROR > order/slave.go:264] it was not possible to encode the JSON struct.")
			} else {
				conn2ToAgent.Write([]byte(lta))
				conn2ToAgent.Write([]byte("\n"))
				conn2ToAgent.Close()
			}

			l2, err := net.Listen("tcp", ":7755")
			if err != nil {
				w.Err("[ERROR > order/slave.go:273] it was not possible to listen on port '7755'.")
			}
			defer l2.Close()
			fmt.Println("[SLAVE:7755] Receiving incremental stream from agent...")

			conn2Slave, err := l2.Accept()
			if err != nil {
				w.Err("[ERROR > order/slave.go:280] it was not possible to accept the connection.")
			}

			// Read the status
			buff := bufio.NewReader(conn2Slave)
			n, err := buff.ReadByte()
			if err != nil {
				w.Err("[ERROR > order/slave.go:287] it was not possible to read the 'dataset byte'.")
			}
			m := string(n)
			snapExist, err := strconv.Atoi(m)
			if err != nil {
				w.Err("[ERROR > order/slave.go:292] it was not possible to convert the string '"+m+"' to integer.")
			}

			// Last snapshot in slave node
			LastSnapshotName := list[amount-1].Name

			switch snapExist {
			// Case: receive the snapshot
			case Zerror:
				// Receive the snapshot
				_, err := zfs.ReceiveSnapshotRollback(conn2Slave, a.DestDataset, true)

				// Check for response to agent
				if err != nil {
					Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s': incoherent.", hostname, a.SnapshotName, a.Source)
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
					w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"': incoherent.")
				} else {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
					w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
					runner = true
				}

			// Case: the received snapshot already existed
			case NothingToDo:
				SnapshotName := lib.SearchName(a.SnapshotUUID)
				renamed := lib.Renamed(a.SnapshotName, SnapshotName)
				if renamed == true {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasRenamed,""}
					w.Info("[INFO] the snapshot '"+a.SnapshotName+"' already exists and it was renamed to '"+SnapshotName+"'.")
				} else {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NothingToDo,""}
					w.Info("[INFO] the snapshot '"+a.SnapshotName+"' already exists.")
				}

			// Case: the last snapshot in slave is the most actual
			case MostActual:
				ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,MostActual,""}
				w.Info("[INFO] the snapshot '"+LastSnapshotName+"' is the most actual.")

			// Case: receive incremental stream
			case Incremental:
				// Receive incremental stream
				_, err := zfs.ReceiveSnapshotRollback(conn2Slave,a.DestDataset,true)

				// Check for response to agent
				if err != nil {
					Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
					w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
				} else {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
					w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
					runner = true
				}
			}
			// Send the last ZFSResponseToAgent
			conn3ToAgent, err := net.Dial("tcp", a.Source+":7766")
			if err != nil {
				w.Err("[ERROR > order/slave.go:351] it was not possible to connect with '"+a.Source+"'.")
			}

			// Marshal response to agent
			rta, err := json.Marshal(ResponseToAgent)
			if err != nil {
				w.Err("[ERROR > order/slave.go:356] it was not possible to encode the JSON struct.")
			} else {
				conn3ToAgent.Write([]byte(rta))
				conn3ToAgent.Write([]byte("\n"))
				conn3ToAgent.Close()
			}
			// Close transmission
			stream = false
		}

		if runner == true {
			pieces	  := config.Extract(index)
			enable	  := pieces[0].(bool)
			delClone  := pieces[2].(bool)
			clone	  := pieces[3].(string)
			dataset	  := pieces[4].(string)
			retain	  := pieces[6].(int)
			getBackup := pieces[7].(bool)
			getClone  := pieces[8].(bool)

			if enable == true {
				// Delete an existing clone
				lib.DeleteClone(delClone, clone)

				// Delete the backup snapshot
				ds, err := zfs.GetDataset(dataset)
				if err != nil {
					w.Err("[ERROR > order/slave.go:384] it was not possible to get the dataset '"+dataset+"'.")
				} else {
					list, err := ds.Snapshots()
					if err != nil {
						w.Err("[ERROR > order/slave.go:388] it was not possible to access of snapshots list in dataset '"+dataset+"'.")
					}
					count = len(list)
					backup, amount := lib.RealList(count, list, dataset)
					if backup != amount {
						take := list[backup-1].Name
						snap, err := zfs.GetDataset(take)
						if err != nil {
							w.Err("[ERROR > order/slave.go:396] it was not possible to get the snapshot '"+take+"'.")
						}
						lib.DeleteBackup(dataset, snap)
					}

					// Retain policy
					lib.Policy(dataset, ds, retain)

					// Create a backup snapshot
					lib.Backup(getBackup, dataset, ds)

					// Clone the last snapshot received
					take := list[amount-1].Name
					snap, err := zfs.GetDataset(take)
					if err != nil {
						w.Err("[ERROR > order/slave.go:411] it was not possible to get the snapshot '"+take+"'.")
					}
					lib.Clone(getClone, clone, take, snap)
				}
			}
		}
	}
	stop := false
	return stop
}
