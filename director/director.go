// Package director contains: agent.go - !director.go - slave.go
//
// Director sends an order to the agent
// Make orders from synchronisation between nodes
//
package director

// Status for DestDataset
const (
	DATASET_TRUE  = iota + 1 // Dataset not empty
	DATASET_FALSE		 // Dataset does not exist or empty
)

// Status for response
const (
	WAS_RENAMED = iota + 1 // The snapshot sent was renamed on destination
	WAS_WRITTEN	       // The snapshot sent was written on destination
	NOTHING_TO_DO	       // The snapshot sent already existed on destination
	ZFS_ERROR	       // Any error string
	NOT_EMPTY	       // Need an incremental stream
	INCREMENTAL	       // Ready to send an incremental stream
	MOST_ACTUAL	       // The last snapshot on destination is the most actual
)
