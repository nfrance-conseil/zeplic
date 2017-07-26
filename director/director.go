// Package director contains: agent.go - director.go - slave.go
//
// Director sends an order to the agent
// Make orders from synchronisation between nodes
//
package director

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/config"
	"github.com/nfrance-conseil/zeplic/lib"
	"github.com/nfrance-conseil/zeplic/utils"
	"github.com/hashicorp/consul/api"
	"github.com/pborman/uuid"
)

var (
	w = config.LogBook()
)

// ServerFilePath returns the path of JSON config file
var ServerFilePath string

// Cold contains the information of backup snapshot
type Cold struct {
	Creation    string `json:"creation"`
	Prefix	    string `json:"prefix"`
	SyncOn      string `json:"sync_on"`
	SyncDataset string `json:"sync_dataset"`
	SyncPolicy  string `json:"sync_policy"`
	Retention   string `json:"retention"`
}

// Hot contains the information of synchronization snapshot
type Hot struct {
	Creation    string `json:"creation"`
	Prefix	    string `json:"prefix"`
	SyncOn      string `json:"sync_on"`
	SyncDataset string `json:"sync_dataset"`
	SyncPolicy  string `json:"sync_policy"`
	Retention   string `json:"retention"`
}

// Actions contains the information of replicate every snapshot
type Actions struct {
	Hostname	 string `json:"hostname"`
	Datacenter	 string `json:"datacenter"`
	Dataset		 string `json:"dataset"`
	Backup		 Cold
	Sync		 Hot
	RollbackIfNeeded bool	`json:"rollback_needed"`
	SkipIfRenamed    bool	`json:"skip_renamed"`
	SkipIfNotWritten bool	`json:"skip_not_written"`
	SkipIfCloned     bool	`json:"skip_cloned"`
}

// Config extracts the interface of JSON server file
type Config struct {
	Director []Actions `json:"datasets"`
}

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
	jsonFile := ServerFilePath
	serverFile, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("[INFO] The file '%s' does not exist! Please, check your configuration...\n\n", jsonFile)
		os.Exit(1)
	}
	var values Config
	err = json.Unmarshal(serverFile, &values)
	if err != nil {
		w.Err("[ERROR > director/director.go:113] it was not possible to parse the JSON configuration file.")
	}
	list := len(values.Director)

	for i := 0; i < list; i++ {
		// Get KV pairs for same datacenter
		hostname   := values.Director[i].Datacenter
		datacenter := values.Director[i].Datacenter
		dataset	   := values.Director[i].Dataset

		// Create a new client
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			w.Err("[ERROR > director/director.go:126]@[CONSUL] it was not possible to create a new client.")
		}
		kv := client.KV()

		// KV write options
		key := fmt.Sprintf("zeplic/%s/", hostname)
		q := &api.QueryOptions{Datacenter: datacenter}

		// Get KV pairs
		pairs, _, err := kv.Keys(key, "", q)
		if err != nil {
			w.Err("[ERROR > director/director.go:137]@[CONSUL] it was not possible to get the KV pairs.")
		}
		if len(pairs) < 1 {
			go lib.Sync(hostname, datacenter, datacenter)
			time.Sleep(10 * time.Second)
		}

		// Extract all information of server JSON file
		backup_creation	    := values.Director[i].Backup.Creation
		backup_prefix	    := values.Director[i].Backup.Prefix
		backup_sync_on	    := values.Director[i].Backup.SyncOn
		backup_sync_dataset := values.Director[i].Backup.SyncDataset
		backup_sync_policy  := values.Director[i].Backup.SyncPolicy
		backup_retention    := values.Director[i].Backup.Retention
		sync_creation	    := values.Director[i].Sync.Creation
		sync_prefix	    := values.Director[i].Sync.Prefix
		sync_sync_on	    := values.Director[i].Sync.SyncOn
		sync_sync_dataset   := values.Director[i].Sync.SyncDataset
		sync_sync_policy    := values.Director[i].Sync.SyncPolicy
		sync_retention      := values.Director[i].Sync.Retention
		rollback	    := values.Director[i].RollbackIfNeeded
		renamed		    := values.Director[i].SkipIfRenamed
		not_written	    := values.Director[i].SkipIfNotWritten
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
		var snapshotsList []string
		for j := 0; j < len(pairs); j++ {
			snapString := utils.After(pairs[j], key)
			snapshotsList = append(snapshotsList, snapString)
		}

		// Remove a snapshot if it contains the flag #deleted or if the dataset is not correct
		for k := 0; k < len(snapshotsList); k++ {
			_, snapName, flag := lib.InfoKV(snapshotsList[k])
			if strings.Contains(flag, "#deleted") {
				snapshotsList = append(snapshotsList[:k], snapshotsList[k+1:]...)
				continue
			} else {
				snapDataset := lib.DatasetName(snapName)
				if snapDataset != dataset {
					snapshotsList = append(snapshotsList[:k], snapshotsList[k+1:]...)
				}
				continue
			}
		}

		// Should I send a take_snapshot order?
		take, SnapshotName := lib.NewSnapshot(snapshotsList, backup_creation, backup_prefix)
		if take == true {
			Action = "take_snapshot"
		} else {
			take, SnapshotName = lib.NewSnapshot(snapshotsList, sync_creation, sync_prefix)
			if take == true {
				Action = "take_snapshot"
			}
		}

		var sent bool
		var uuidList []string
		if take == false {
			// Should I send a send_snapshot order?
			sent, uuid := lib.Send(dataset, snapshotsList, backup_sync_policy, backup_prefix)
			if sent == true {
				Action = "send_snapshot"
				Destination = backup_sync_on
				uuidList = append(uuidList, uuid)
				DestDataset = backup_sync_dataset
			} else {
				sent, uuid = lib.Send(dataset, snapshotsList, sync_sync_policy, sync_prefix)
				if sent == true {
					Action = "send_snapshot"
					Destination = sync_sync_on
					uuidList = append(uuidList, uuid)
					DestDataset = sync_sync_dataset
				}
			}
		}

		if sent == false {
			// Should I send a destroy_snapshot order?
			destroy, list := lib.Delete(dataset, snapshotsList, backup_prefix, backup_retention)
			if destroy == true {
				Action = "destroy_snapshot"
				SnapshotUUID = list
			} else {
				destroy, list := lib.Delete(dataset, snapshotsList, sync_prefix, sync_retention)
				if destroy == true {
					Action = "destroy_snapshot"
					SnapshotUUID = list
				}
			}
		}

		switch Action {

		// Take a new snapshot
		case "take_snapshot":
			Destination	 = ""
			SnapshotUUID	 = append(SnapshotUUID, "")
			DestDataset	 = dataset
			RollbackIfNeeded = rollback
			SkipIfRenamed    = renamed
			SkipIfNotWritten = not_written
			SkipIfCloned     = cloned

		// Send a snapshot
		case "send_snapshot":
			SnapshotUUID	 = uuidList
			SnapshotName	 = ""
			RollbackIfNeeded = rollback
			SkipIfRenamed    = renamed
			SkipIfNotWritten = not_written
			SkipIfCloned     = cloned

		// Destroy a snapshot
		case "destroy_snapshot":
			Destination	 = ""
			SnapshotName	 = ""
			DestDataset	 = ""
			RollbackIfNeeded = rollback
			SkipIfRenamed    = renamed
			SkipIfNotWritten = not_written
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

			// Send order to agent
			connToAgent, err := net.Dial("tcp", hostname+":7711")
			if err != nil {
				w.Err("[ERROR > director/director.go:289] it was not possible to connect with '"+hostname+"'.")
			}

			// Marshal response to agent
			ota, err := json.Marshal(OrderToAgent)
			if err != nil {
				w.Err("[ERROR > director/director.go:295] it was not possible to encode the JSON struct.")
			} else {
				connToAgent.Write([]byte(ota))
				connToAgent.Write([]byte("\n"))
				connToAgent.Close()
			}
		} else {
			continue
		}
	}
}
