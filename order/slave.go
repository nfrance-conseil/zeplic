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
		w.Err("[ERROR] it was not possible to resolve the hostname.")
	}

	// Unmarshal orders from agent
	var a ZFSOrderFromAgent
	agent, err := bufio.NewReader(connSlave).ReadBytes('\x0A')
	if err != nil {
		w.Err("[ERROR] an error has occurred while reading from the socket.")
	}
	err = json.Unmarshal(agent, &a)
	if err != nil {
		w.Err("[ERROR] it was not possible to parse the JSON struct from the socket.")
	}

	// Struct for Status constant
	ack := make([]byte, 0)
	// Variable to receive an incremental stream
	stream := false

	// Check if the dataset received exists
	ds, err := zfs.GetDataset(a.DestDataset)

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
		dataset := pieces[3].(string)

		if dataset == a.DestDataset {
			index = i
			break
		} else {
			continue
		}
	}

	// Dataset does not exit
	if err != nil {
		if index > -1 {
			// Extract data of dataset
			pieces := config.Extract(index)
			enable := pieces[0].(bool)
			dataset := pieces[3].(string)

			if dataset == a.DestDataset && enable == true {
				// Status for DestDataset
				ack = nil
				ack = strconv.AppendInt(ack, DatasetFalse, 10)
				connSlave.Write(ack)

				// Receive the snapshot
				_, err := zfs.ReceiveSnapshotRollback(connSlave, a.DestDataset, false)

				// Take the snapshot received
				ds, _ := zfs.GetDataset(dataset)
				list, _ := ds.Snapshots()
				SnapshotName := list[0].Name
				s, _ := zfs.GetDataset(SnapshotName)

				// Apply configuration
				clone	  := pieces[2].(string)
				getBackup := pieces[6].(bool)
				getClone  := pieces[7].(bool)
				lib.Backup(getBackup, dataset, ds)
				lib.Clone(getClone, clone, SnapshotName, s)

				// Check for response to agent
				if err != nil {
					Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
					w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
				} else {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
					w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
				}
			} else if dataset == a.DestDataset && enable == false {
				// Status for DestDataset
				ack = nil
				ack = strconv.AppendInt(ack, DatasetDisable, 10)
				connSlave.Write(ack)
				connSlave.Close()
				w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is disabled.")
			}
		} else {
			// Status for DestDataset
			ack = nil
			ack = strconv.AppendInt(ack, DatasetNotConf, 10)
			connSlave.Write(ack)
			connSlave.Close()
			w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is not configured.")
		}
	} else {
		// Get the last snapshot in DestDataset
		list, _ = ds.Snapshots()
		count = len(list)

		// Get the correct number of snapshots in dataset
		_, amount := lib.RealList(count, list, a.DestDataset)

		// Dataset is empty
		if amount == 0 {
			if index > -1 {
				// Extract data of dataset
				pieces := config.Extract(index)
				enable := pieces[0].(bool)
				dataset := pieces[3].(string)

				if dataset == a.DestDataset && enable == true {
					// Status for DestDataset
					ack = nil
					ack = strconv.AppendInt(ack, DatasetFalse, 10)
					connSlave.Write(ack)

					// Receive the snapshot
					_, err := zfs.ReceiveSnapshotRollback(connSlave, a.DestDataset, false)

					// Take the snapshot received
					ds, _ := zfs.GetDataset(dataset)
					list, _ := ds.Snapshots()
					SnapshotName := list[0].Name
					s, _ := zfs.GetDataset(SnapshotName)

					// Apply configuration
					clone	  := pieces[2].(string)
					getBackup := pieces[6].(bool)
					getClone  := pieces[7].(bool)
					lib.Backup(getBackup, dataset, ds)
					lib.Clone(getClone, clone, SnapshotName, s)

					// Check for response to agent
					if err != nil {
						Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
						ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
						w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
					} else {
						ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
						w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
					}
				} else if dataset == a.DestDataset && enable == false {
					// Status for DestDataset
					ack = nil
					ack = strconv.AppendInt(ack, DatasetDisable, 10)
					connSlave.Write(ack)
					connSlave.Close()
					w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is disabled.")
				}
			} else {
				// Status for DestDataset
				ack = nil
				ack = strconv.AppendInt(ack, DatasetNotConf, 10)
				connSlave.Write(ack)
				connSlave.Close()
				w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is not configured.")
			}
		} else {
			// Status for DestDataset
			ack = nil
			ack = strconv.AppendInt(ack, DatasetTrue, 10)
			connSlave.Write(ack)

			// Get the last snapshot in DestDataset
			LastSnapshotName := list[count-1].Name
			// Get its uuid
			snap, err := zfs.GetDataset(LastSnapshotName)
			if err != nil {
				w.Err("[ERROR] it was not possible to get the snapshot '"+snap.Name+"'.")
			}
			LastSnapshotUUID := lib.SearchUUID(snap)

			// Check if the snapshot was renamed
			renamed := lib.Renamed(a.SnapshotName, LastSnapshotName)
			if LastSnapshotUUID == a.SnapshotUUID {
				if renamed == true {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasRenamed,""}
					w.Info("[INFO] the snapshot '"+a.SnapshotName+"' already existed but it was renamed to '"+LastSnapshotName+"'.")
				} else {
					ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NothingToDo,""}
					w.Info("[INFO] the snapshot '"+LastSnapshotName+"' already existed.")
				}
			} else {
				// Information to agent where Error field contains the uuid of last snapshot in slave
				ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,NotEmpty,""}
				stream = true
			}
		}
	}

	// Reconnection to send ZFSResponseToAgent
	connToAgent, err := net.Dial("tcp", a.Source+":7733")

	// Marshal response to agent
	rta, err := json.Marshal(ResponseToAgent)
	if err != nil {
		w.Err("[ERROR] it was not possible to encode the JSON struct.")
	} else {
		connToAgent.Write([]byte(rta))
		connToAgent.Write([]byte("\n"))
		connToAgent.Close()
	}

	// MapUUID to save the list of uuids
	var MapUUID []string

	// Get the last snapshot in DestDataset
	list, _ = ds.Snapshots()
	count = len(list)

	// Get the correct number of snapshots in dataset
	_, amount := lib.RealList(count, list, a.DestDataset)

	if stream == true {
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

		// Marshal response to agent
		lta, err := json.Marshal(ListUUIDsToAgent)
		if err != nil {
			w.Err("[ERROR] it was not possible to encode the JSON struct.")
		} else {
			conn2ToAgent.Write([]byte(lta))
			conn2ToAgent.Write([]byte("\n"))
			conn2ToAgent.Close()
		}

		l2, _ := net.Listen("tcp", ":7755")
		defer l2.Close()
		fmt.Println("[SLAVE:7755] Receiving incremental stream from agent...")

		conn2Slave, _ := l2.Accept()

		// Read the status
		buff := bufio.NewReader(conn2Slave)
		n, _ := buff.ReadByte()
		snapExist, _ := strconv.Atoi(string(n))

		// Last snapshot in slave node
		LastSnapshotName := list[amount-1].Name

		switch snapExist {
		// Case: receive the snapshot
		case Zerror:
			// Receive the snapshot
			_, err := zfs.ReceiveSnapshotRollback(conn2Slave, a.DestDataset, true)

			// Check for response to agent
			if err != nil {
				Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
				ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
				w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
			} else {
				ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
				w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
			}

		// Case: the received snapshot already existed
		case NothingToDo:
			ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NothingToDo,""}
			w.Info("[INFO] the snapshot '"+a.SnapshotName+"' is already existed.")

		// Case: the last snapshot in slave is the most actual
		case MostActual:
			ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NothingToDo,""}
			w.Info("[INFO] the snapshot '"+LastSnapshotName+"' is the most actual.")

		// Case: receive incremental stream
		case Incremental:
			// Receive incremental stream
			zfs.ReceiveSnapshotRollback(conn2Slave,a.DestDataset,true)

			// Check for response to agent
			if err != nil {
				Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", hostname, a.SnapshotName, a.Source)
				ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
				w.Err("[ERROR] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
			} else {
				ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}
				w.Info("[INFO] the snapshot '"+a.SnapshotName+"' has been received.")
			}
		}
		// Send the last ZFSResponseToAgent
		conn3ToAgent, err := net.Dial("tcp", a.Source+":7766")

		// Marshal response to agent
		rta, err = json.Marshal(ResponseToAgent)
		if err != nil {
			w.Err("[ERROR] it was not possible to encode the JSON struct.")
		} else {
			conn3ToAgent.Write([]byte(rta))
			conn3ToAgent.Write([]byte("\n"))
			conn3ToAgent.Close()
		}
		// Close transmission
		stream = false
	}

	pieces	  := config.Extract(index)
	enable	  := pieces[0].(bool)
	delClone  := pieces[1].(bool)
	clone	  := pieces[2].(string)
	dataset	  := pieces[3].(string)
	retain	  := pieces[5].(int)
	getBackup := pieces[6].(bool)
	getClone  := pieces[7].(bool)

	if enable == true {
		// Delete an existing clone
		lib.DeleteClone(delClone, clone)

		// Delete the backup snapshot
		ds, _ := zfs.GetDataset(dataset)
		list, _ := ds.Snapshots()
		count = len(list)
		backup, amount := lib.RealList(count, list, dataset)
		if backup != amount {
			take := list[backup-1].Name
			snap, _ := zfs.GetDataset(take)
			lib.DeleteBackup(dataset, snap)
		}

		// Retain policy
		lib.Policy(dataset, ds, retain)

		// Create a backup snapshot
		lib.Backup(getBackup, dataset, ds)

		// Clone the last snapshot received
		take := list[amount-1].Name
		snap, _ := zfs.GetDataset(take)
		lib.Clone(getClone, clone, take, snap)
	}
	stop := false
	return stop
}
