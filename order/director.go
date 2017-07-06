// Package order contains: agent.go - !director.go - slave.go
//
// Director sends an order to the agent
// Make orders from synchronisation between nodes
//
package order

// Status for DestDataset
const (
	DatasetTrue    = iota + 1 // Dataset not empty
	DatasetFalse		  // Dataset does not exist or empty
	DatasetDisable		  // Dataset disabled
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
