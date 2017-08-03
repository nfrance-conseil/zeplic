// Package director contains: actions.go - agent.go - director.go - slave.go
//
// Director sends an order to the agent
// Make orders from synchronisation between nodes
//
package director

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/nfrance-conseil/zeplic/tools"
	"github.com/pborman/uuid"
)

var (
	w = config.LogBook()
)

// Status for DestDataset
const (
	DatasetTrue    = iota + 1 // Dataset not empty
	DatasetFalse		  // Dataset does not exist or empty
	DatasetDisable		  // Dataset disabled
	DatasetDocker             // Dataset docker
	DatasetNotConf		  // Dataset not configured
)

// Status for response
const (
	WasRenamed = iota + 1 // The snapshot sent was renamed on destination
	WasWritten	      // The snapshot sent was written on destination
	NothingToDo	      // The snapshot sent already existed on destination
	Zerror		      // Any error string
	NotEmpty	      // Need an incremental stream
	Incremental	      // Ready to send an incremental stream
	MostActual	      // The last snapshot on destination is the most actual
)

// ZFSDirectorsOrder is the struct for the director's orders
type ZFSDirectorsOrder struct {
	OrderUUID        string `json:"OrderUUID"`	  // Mandatory
	Action           string `json:"Action"`		  // take_snapshot, send_snapshot, destroy_snapshot
	Destination      string `json:"Destination"`	  // Hostname or IP for send
	SnapshotUUID   []string `json:"SnapshotUUID"`	  // List of snapshots
	SnapshotName     string `json:"SnapshotName"`	  // Name of snapshot for take_snapshot order
	DestDataset      string `json:"DestDataset"`	  // Dataset for receive
	RollbackIfNeeded bool   `json:"RollbackIfNeeded"` // Should I rollback if written is true on destination
	SkipIfRenamed    bool   `json:"SkipIfRenamed"`	  // Should I do the stuff if a snapshot has been renamed
	SkipIfNotWritten bool   `json:"SkipIfNotWritten"` // Should I take a snapshot if nothing is written
	SkipIfCloned     bool   `json:"SkipIfCloned"`	  // Should I delete a snapshot if it was cloned
}

// Director reads the server config file and all KV pairs
// Then it creates the orders
func Director() {
	// Get the data of server file
	values := config.Server()

	// KV write options
	keyfix := fmt.Sprintf("zeplic/")
	datacenter := values.Datacenter

	// Get KV pairs
	pairs := lib.ListKV(keyfix, datacenter)
	if len(pairs) < 1 {
		for i := 0; i < len(values.Director); i++ {
			hostname := values.Director[i].Hostname
			dataset  := values.Director[i].Dataset
			go lib.Sync(hostname, datacenter, dataset, i)
			continue
		}
		time.Sleep(10 * time.Second)
	}

	// Resynchronization?
	hour, min, _ := time.Now().Clock()
	if hour == 18 && min > 49 && min < 58 {
		for i := 0; i < len(values.Director); i++ {
			hostname   := values.Director[i].Hostname
			dataset	   := values.Director[i].Dataset

			// Resync
			OrderUUID	 := "zRESYNC"
			Action		 := "kv_resync"
			Destination	 := datacenter
			SnapshotUUID	 := []string{""}
			SnapshotName	 := ""
			DestDataset	 := dataset
			RollbackIfNeeded := false
			SkipIfRenamed    := false
			SkipIfNotWritten := false
			SkipIfCloned     := false

			// Create order to agent
			OrderToResync := ZFSDirectorsOrder{OrderUUID,Action,Destination,SnapshotUUID,SnapshotName,DestDataset,RollbackIfNeeded,SkipIfRenamed,SkipIfNotWritten,SkipIfCloned}

			// Marshal response to agent
			otr, err := json.Marshal(OrderToResync)
			if err != nil {
				w.Err("[ERROR > director/director.go:104] it was not possible to encode the JSON struct.")
			} else {
				// Send order to agent
				ConnToResync, err := net.Dial("tcp", hostname+":7711")
				if err != nil {
					w.Err("[ERROR > director/director.go:109] it was not possible to connect with '"+hostname+"'.")
				} else {
					ConnToResync.Write([]byte(otr))
					ConnToResync.Write([]byte("\n"))
					ConnToResync.Close()
				}
				time.Sleep(10 * time.Second)
			}
		}
	} else {
		for i := 0; i < len(values.Director); i++ {
			// Get KV pairs for same datacenter
			hostname   := values.Director[i].Hostname
			dataset	   := values.Director[i].Dataset

			// KV write options
			keyfix = fmt.Sprintf("zeplic/%s/", hostname)

			// Get KV pairs
			pairs := lib.ListKV(keyfix, datacenter)
			var PairsList []string
			for j := 0; j < len(pairs); j++ {
				pair := fmt.Sprintf("%s:%s", pairs[j].Key, string(pairs[j].Value[:]))
				PairsList = append(PairsList, pair)
			}

			// Extract all information of server JSON file
			BackupCreation	    := values.Director[i].Backup.Creation
			BackupPrefix	    := values.Director[i].Backup.Prefix
			BackupSyncOn	    := values.Director[i].Backup.SyncOn
			BackupSyncDataset   := values.Director[i].Backup.SyncDataset
			BackupSyncPolicy    := values.Director[i].Backup.SyncPolicy
			BackupRetention     := values.Director[i].Backup.Retention
			SyncCreation	    := values.Director[i].Sync.Creation
			SyncPrefix	    := values.Director[i].Sync.Prefix
			SyncSyncOn	    := values.Director[i].Sync.SyncOn
			SyncSyncDataset	    := values.Director[i].Sync.SyncDataset
			SyncSyncPolicy	    := values.Director[i].Sync.SyncPolicy
			SyncRetention	    := values.Director[i].Sync.Retention
			rollback	    := values.Director[i].RollbackIfNeeded
			renamed		    := values.Director[i].SkipIfRenamed
			NotWritten	    := values.Director[i].SkipIfNotWritten
			cloned		    := values.Director[i].SkipIfCloned

			// Define variables of order struct
			var OrderUUID	   string
			var Action	   string
			var Destination    string
			var SnapshotUUID []string
			var SnapshotName   string
			var DestDataset    string
			var RollbackIfNeeded bool
			var SkipIfRenamed    bool
			var SkipIfNotWritten bool
			var SkipIfCloned     bool

			// Add only the snapshot name, uuid and the flag to new snapshots list
			var SnapshotsList []string
			for k := 0; k < len(PairsList); k++ {
				snapString := tools.After(PairsList[k], keyfix)
				SnapshotsList = append(SnapshotsList, snapString)
			}

			// Remove a snapshot if it contains the flag #deleted or if the dataset is not correct
			for m := 0; m < len(SnapshotsList); m++ {
				_, snapName, flag := lib.InfoKV(SnapshotsList[m])
				if strings.Contains(flag, "#deleted") {
					SnapshotsList = append(SnapshotsList[:m], SnapshotsList[m+1:]...)
					m--
					continue
				} else {
					snapDataset := lib.DatasetName(snapName)
					if snapDataset != dataset {
						SnapshotsList = append(SnapshotsList[:m], SnapshotsList[m+1:]...)
						m--
					}
					continue
				}
			}

			// Take snapshot?
			var take bool
			if BackupCreation == "" && SyncCreation == "" {
				take = false
			} else {
				// Should I send a take_snapshot order?
				take, SnapshotName = NewSnapshot(SnapshotsList, BackupCreation, BackupPrefix)
				if take == true {
					Action = "take_snapshot"
				} else {
					take, SnapshotName = NewSnapshot(SnapshotsList, SyncCreation, SyncPrefix)
					if take == true {
						Action = "take_snapshot"
					}
				}
			}

			// Send snapshot?
			var sent bool
			var uuidList []string
			if take == false {
				// Should I send a send_snapshot order?
				sent, uuid := Send(dataset, SnapshotsList, BackupSyncPolicy, BackupPrefix)
				if sent == true {
					Action = "send_snapshot"
					Destination = BackupSyncOn
					uuidList = append(uuidList, uuid)
					DestDataset = BackupSyncDataset
				} else {
					sent, uuid = Send(dataset, SnapshotsList, SyncSyncPolicy, SyncPrefix)
					if sent == true {
						Action = "send_snapshot"
						Destination = SyncSyncOn
						uuidList = append(uuidList, uuid)
						DestDataset = SyncSyncDataset
					}
				}
			}

			// Destroy snapshot?
			if sent == false {
				// Should I send a destroy_snapshot order?
				destroy, list := Delete(dataset, SnapshotsList, BackupPrefix, BackupRetention)
				if destroy == true {
					Action = "destroy_snapshot"
					SnapshotUUID = list
				} else {
					destroy, list := Delete(dataset, SnapshotsList, SyncPrefix, SyncRetention)
					if destroy == true {
						Action = "destroy_snapshot"
						SnapshotUUID = list
					}
				}
			}

			// Actions...
			switch Action {

			// Take a new snapshot
			case "take_snapshot":
				Destination	 = ""
				SnapshotUUID	 = append(SnapshotUUID, "")
				DestDataset	 = dataset
				RollbackIfNeeded = rollback
				SkipIfRenamed    = renamed
				SkipIfNotWritten = NotWritten
				SkipIfCloned     = cloned

			// Send a snapshot
			case "send_snapshot":
				SnapshotUUID	 = uuidList
				SnapshotName	 = ""
				RollbackIfNeeded = rollback
				SkipIfRenamed    = renamed
				SkipIfNotWritten = NotWritten
				SkipIfCloned     = cloned

			// Destroy a snapshot
			case "destroy_snapshot":
				Destination	 = ""
				SnapshotName	 = ""
				DestDataset	 = ""
				RollbackIfNeeded = rollback
				SkipIfRenamed    = renamed
				SkipIfNotWritten = NotWritten
				SkipIfCloned     = cloned

			// No action
			default:
				continue
			}

			if Action != "" {
				// New OrderUUID
				OrderUUID = uuid.New()

				// Create order to agent
				OrderToAgent := ZFSDirectorsOrder{OrderUUID,Action,Destination,SnapshotUUID,SnapshotName,DestDataset,RollbackIfNeeded,SkipIfRenamed,SkipIfNotWritten,SkipIfCloned}

				// Marshal response to agent
				ota, err := json.Marshal(OrderToAgent)
				if err != nil {
					w.Err("[ERROR > director/director.go:291] it was not possible to encode the JSON struct.")
				} else {
					// Send order to agent
					ConnToAgent, err := net.Dial("tcp", hostname+":7711")
					if err != nil {
						w.Err("[ERROR > director/director.go:296] it was not possible to connect with '"+hostname+"'.")
					} else {
						ConnToAgent.Write([]byte(ota))
						ConnToAgent.Write([]byte("\n"))
						ConnToAgent.Close()
					}
				}
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(10 * time.Second)
				continue
			}
		}
	}
}
