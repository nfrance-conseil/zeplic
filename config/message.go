// format for messages : ZFSOrder and ZFSResponse

package config

import (
	"io/ioutil"
)

type ZFSOrder struct {
	OrderUUID	 string	// mandatory
	Action		 string // take_snapshot, send_snapshot, destroy_snapshot
	Dataset		 string // tank/plop
	SnapshotName	 string // Snapshot is dataset@snapshotname
	SnapshotUUID	 string // mandatory
	Destination	 string // hostname or ip for send
	DestDataset	 string // dataset for recieve
	SkipIfNotWritten bool   // should I take a snapshot if nothing is written
	RollbackIfNeeded bool   // should I rollback if written is true on destination
	SkipIfRenamed	 bool   // should I do the stuff if a snapshot has been renamed
}

// Status for response
const (
	WAS_RENAMED   = 1
	WAS_WRITTEN   = 2
	NOTHING_TO_DO = 3
	ZFS_ERROR     = 4
	NETWORK_ERROR = 5
)

type ZFSResponse struct {
	OrderUUID    string  // reference to a valid order
	IsSuccess    bool    // true or false
	Status	     int64   // 
	Error	     string  // error string if needed
}

// Read and convert json from file descriptor
func fromJSON (r io.Reader)  (ZFSOrder, error) {
	var z ZFSOrder
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err := json.Unmarshal(data, &z)
	if err != nil {
		return nil, err
	}
	if ( z.OrderUUID == nil || z.SnapshotUUID == nil)  {
		return nil, errors.New("Inconsistant data structure")
	}
}
