// Package director contains: agent.go - consul.go - director.go - server.go - slave.go
//
// Slave receives a snapshot from agent
//
package director

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
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
func HandleRequestSlave (ConnSlave net.Conn) {
	// Unmarshal orders from agent
	var a ZFSOrderFromAgent
	agent, err := bufio.NewReader(ConnSlave).ReadBytes('\x0A')
	if err != nil {
		w.Err("[ERROR > director/slave.go:45] an error has occurred while reading from the socket.")
	} else {
		err = json.Unmarshal(agent, &a)
		if err != nil {
			w.Err("[ERROR > director/slave.go:49] it was not possible to parse the JSON struct from the socket.")
		} else {
			// Receive a snapshot
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
				values := config.Local()

				// Check if dataset is configured
				index := -1
				for i := 0; i < len(values.Dataset); i++ {
					dataset := values.Dataset[i].Name
					if dataset == a.DestDataset {
						index = i
						break
					} else {
						continue
					}
				}

				if index > -1 {
					// Extract data of dataset
					enable	   := values.Dataset[index].Enable
					docker	   := values.Dataset[index].Docker
					dataset	   := values.Dataset[index].Name
					datacenter := values.Dataset[index].Consul.Datacenter

					if dataset == a.DestDataset && enable == true && docker == true {
						// Check if the dataset received exists
						ds, err := zfs.GetDataset(a.DestDataset)
						if err != nil {
							// Status for DestDataset
							ack = nil
							ack = strconv.AppendInt(ack, DatasetFalse, 10)
							ConnSlave.Write(ack)

							// Receive the snapshot
							_, err := zfs.ReceiveSnapshotRollback(ConnSlave, a.DestDataset, false)

							// Check for response to agent
							if err != nil {
								Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", lib.Host(), a.SnapshotName, a.Source)
								ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
								w.Err("[ERROR > director/slave.go:102] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
								more = true
							} else {
								ds, err := zfs.GetDataset(a.DestDataset)
								if err != nil {
									w.Err("[ERROR > director/slave.go:111] it was not possible to get the dataset '"+a.DestDataset+"'.")
								} else {
									list, err := ds.Snapshots()
									if err != nil {
										w.Err("[ERROR > director/slave.go:115] it was not possible to access of snapshots list.")
									} else {
										count = len(list)
										_, amount := lib.RealList(count, list, a.DestDataset)
										ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}

										for i := 0; i < amount; i++ {
											w.Info("[INFO] the snapshot '"+list[i].Name+"' has been received.")

											// KV write options
											key := fmt.Sprintf("%s/%s/%s", "zeplic", a.Source, a.SnapshotUUID)
											value := fmt.Sprintf("%s#%s", a.SnapshotName, "sent")

											// Edit KV pair
											go lib.PutKV(key, value, datacenter)
										}
										more = true
										runner = true
									}
								}
							}
						} else {
							// Get the last snapshot in DestDataset
							list, err := ds.Snapshots()
							if err != nil {
								w.Err("[ERROR > director/slave.go:140] it was not possible to access of snapshots list in dataset '"+a.DestDataset+"'.")
							} else {
								count = len(list)

								// Get the correct number of snapshots in dataset
								backup, amount := lib.RealList(count, list, a.DestDataset)

								if amount == 0 {
									// Status for DestDataset
									ack = nil
									ack = strconv.AppendInt(ack, DatasetFalse, 10)
									ConnSlave.Write(ack)

									if backup != amount {
										list, err := ds.Snapshots()
										if err != nil {
											w.Err("[ERROR > director/slave.go:156] it was not possible to access of snapshots list in dataset '"+a.DestDataset+"'.")
										} else {
											backup := list[0].Name
											snap, err := zfs.GetDataset(backup)
											if err != nil {
												w.Err("[ERROR > director/slave.go:161] it was not possible to get the snapshot '"+backup+"'.")
											} else {
												err := snap.Destroy(zfs.DestroyDefault)
												if err != nil {
													w.Err("[ERROR > director/slave.go:165] it was not possible to destroy the backup snapshot '"+backup+"'.")
												} else {
													w.Info("[INFO] the backup snapshot '"+backup+"' has been destroyed.")
												}
											}
										}
									}

									// Receive the snapshot
									_, err := zfs.ReceiveSnapshotRollback(ConnSlave, a.DestDataset, true)

									// Check for response to agent
									if err != nil {
										Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", lib.Host(), a.SnapshotName, a.Source)
										ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
										w.Err("[ERROR > director/slave.go:176] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
										more = true
									} else {
										ds, err := zfs.GetDataset(a.DestDataset)
										if err != nil {
											w.Err("[ERROR > director/slave.go:185] it was not possible to get the dataset '"+a.DestDataset+"'.")
										} else {
											list, err := ds.Snapshots()
											if err != nil {
												w.Err("[ERROR > director/slave.go:189] it was not possible to access of snapshots list.")
											} else {
												count = len(list)
												_, amount := lib.RealList(count, list, a.DestDataset)
												ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}

												for i := 0; i < amount; i++ {
													w.Info("[INFO] the snapshot '"+list[i].Name+"' has been received.")

													// KV write options
													key := fmt.Sprintf("%s/%s/%s", "zeplic", a.Source, a.SnapshotUUID)
													value := fmt.Sprintf("%s#%s", a.SnapshotName, "sent")

													// Edit KV pair
													go lib.PutKV(key, value, datacenter)
												}
												more = true
												runner = true
											}
										}
									}
								} else {
									// Status for DestDataset
									ack = nil
									ack = strconv.AppendInt(ack, DatasetTrue, 10)
									ConnSlave.Write(ack)

									// Information to agent where Error field contains the uuid of last snapshot in slave
									ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NotEmpty,""}
									more = true
									stream = true
								}
							}
						}
					} else if dataset == a.DestDataset && enable == false {
						// Status for DestDataset
						ack = nil
						ack = strconv.AppendInt(ack, DatasetDisable, 10)
						ConnSlave.Write(ack)
						ConnSlave.Close()
						w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is disabled.")
					} else if dataset == a.DestDataset && docker == false {
						// Status for DestDataset
						ack = nil
						ack = strconv.AppendInt(ack, DatasetDocker, 10)
						ConnSlave.Write(ack)
						ConnSlave.Close()
						w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is not a docker dataset.")
					}
				} else {
					// Status for DestDataset
					ack = nil
					ack = strconv.AppendInt(ack, DatasetNotConf, 10)
					ConnSlave.Write(ack)
					ConnSlave.Close()
					w.Notice("[NOTICE] impossible to receive: the dataset '"+a.DestDataset+"' is not configured.")
				}

				// ResponseToAgent
				var rta []byte
				if more == true {
					// Marshal response to agent
					rta, err = json.Marshal(ResponseToAgent)
					if err != nil {
						w.Err("[ERROR > director/slave.go:253] it was not possible to encode the JSON struct.")
					} else {
						// Reconnection to send ZFSResponseToAgent
						ConnToAgent, err := net.Dial("tcp", a.Source+":7733")
						if err != nil {
							w.Err("[ERROR > director/slave.go:258] it was not possible to connect with '"+a.Source+"'.")
						} else {
							ConnToAgent.Write([]byte(rta))
							ConnToAgent.Write([]byte("\n"))
							ConnToAgent.Close()
						}
					}
				}

				if stream == true {
					// MapUUID to save the list of uuids
					var MapUUID []string

					// Get the list of snapshots in DestDataset
					ds, err := zfs.GetDataset(a.DestDataset)
					if err != nil {
						w.Err("[ERROR > director/slave.go:274] it was not possible to get the dataset '"+a.DestDataset+"'.")
					} else {
						list, err = ds.Snapshots()
						if err != nil {
							w.Err("[ERROR > director/slave.go:278] it was not possible to access of snapshots list in dataset '"+a.DestDataset+"'.")
						} else {
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

							// Marshal response to agent
							lta, err := json.Marshal(ListUUIDsToAgent)
							if err != nil {
								w.Err("[ERROR > director/slave.go:297] it was not possible to encode the JSON struct.")
							} else {
								// Send the list of uuids in DestDataset
								Conn2ToAgent, err := net.Dial("tcp", a.Source+":7744")
								if err != nil {
									w.Err("[ERROR > director/slave.go:302] it was not possible to connect with '"+a.Source+"'.")
								} else {
									Conn2ToAgent.Write([]byte(lta))
									Conn2ToAgent.Write([]byte("\n"))
									Conn2ToAgent.Close()
								}

								l2, err := net.Listen("tcp", ":7755")
								if err != nil {
									w.Err("[ERROR > director/slave.go:311] it was not possible to listen on port '7755'.")
								} else {
									defer l2.Close()
									Conn2Slave, err := l2.Accept()

									if err != nil {
										w.Err("[ERROR > director/slave.go:317] it was not possible to accept the connection.")
									} else {
										// Read the status
										buff := bufio.NewReader(Conn2Slave)
										n, err := buff.ReadByte()
										if err != nil {
											w.Err("[ERROR > director/slave.go:323] it was not possible to read the 'dataset byte'.")
										} else {
											m := string(n)
											snapExist, _ := strconv.Atoi(m)

											// Last snapshot in slave node
											LastSnapshotName := list[amount-1].Name

											switch snapExist {
											// Case: receive the snapshot
											case Zerror:
												// Receive the snapshot
												_, err := zfs.ReceiveSnapshotRollback(Conn2Slave, a.DestDataset, true)

												// Check for response to agent
												if err != nil {
													Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s': incoherent.", lib.Host(), a.SnapshotName, a.Source)
													ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
													w.Err("[ERROR > director/slave.go:337] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"': incoherent.")
													break
												}
												ds, err := zfs.GetDataset(a.DestDataset)
												if err != nil {
													w.Err("[ERROR > director/slave.go:346] it was not possible to get the dataset '"+a.DestDataset+"'.")
													break
												}
												list, err := ds.Snapshots()
												if err != nil {
													w.Err("[ERROR > director/slave.go:351] it was not possible to access of snapshots list.")
													break
												}
												count = len(list)
												_, newAmount := lib.RealList(count, list, a.DestDataset)
												ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}

												for i := amount; i < newAmount-1; i++ {
													w.Info("[INFO] the snapshot '"+list[i].Name+"' has been received.")

													// KV write options
													key := fmt.Sprintf("%s/%s/%s", "zeplic", a.Source, a.SnapshotUUID)
													value := fmt.Sprintf("%s#%s", a.SnapshotName, "sent")

													// Edit KV pair
													go lib.PutKV(key, value, values.Dataset[index].Consul.Datacenter)
												}
												runner = true

											// Case: the received snapshot already existed
											case NothingToDo:
												SnapshotName := lib.SearchName(a.SnapshotUUID)
												renamed := lib.SnapRenamed(a.SnapshotName, SnapshotName)

												if renamed == true {
													ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasRenamed,""}
													w.Info("[INFO] the snapshot '"+a.SnapshotName+"' already exists and it was renamed to '"+SnapshotName+"'.")
													break
												}
												ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,NothingToDo,""}
												w.Info("[INFO] the snapshot '"+a.SnapshotName+"' already exists.")

											// Case: the last snapshot in slave is the most actual
											case MostActual:
												ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,MostActual,""}
												w.Info("[INFO] the snapshot '"+LastSnapshotName+"' is the most actual.")

											// Case: receive incremental stream
											case Incremental:
												// Receive incremental stream
												_, err := zfs.ReceiveSnapshotRollback(Conn2Slave,a.DestDataset,true)

												// Check for response to agent
												if err != nil {
													Error := fmt.Sprintf("[ERROR from '%s'] it was not possible to receive the snapshot '%s' from '%s'.", lib.Host(), a.SnapshotName, a.Source)
													ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,false,Zerror,Error}
													w.Err("[ERROR > director/slave.go:393] it was not possible to receive the snapshot '"+a.SnapshotName+"' from '"+a.Source+"'.")
													break
												}
												ds, err := zfs.GetDataset(a.DestDataset)
												if err != nil {
													w.Err("[ERROR > director/slave.go:402] it was not possible to get the dataset '"+a.DestDataset+"'.")
													break
												}
												list, err := ds.Snapshots()
												if err != nil {
													w.Err("[ERROR > director/slave.go:407] it was not possible to access of snapshots list.")
													break
												}
												count = len(list)
												_, newAmount := lib.RealList(count, list, a.DestDataset)
												ResponseToAgent = ZFSResponseToAgent{a.OrderUUID,true,WasWritten,""}

												for i := amount; i < newAmount-1; i++ {
													w.Info("[INFO] the snapshot '"+list[i].Name+"' has been received.")

													// KV write options
													key := fmt.Sprintf("%s/%s/%s", "zeplic", a.Source, a.SnapshotUUID)
													value := fmt.Sprintf("%s#%s", a.SnapshotName, "sent")

													// Edit KV pair
													go lib.PutKV(key, value, values.Dataset[index].Consul.Datacenter)
												}
												runner = true
											}
										}
										// Marshal response to agent
										rta, err := json.Marshal(ResponseToAgent)
										if err != nil {
											w.Err("[ERROR > director/slave.go:430] it was not possible to encode the JSON struct.")
										} else {
											// Send the last ZFSResponseToAgent
											Conn3ToAgent, err := net.Dial("tcp", a.Source+":7766")
											if err != nil {
												w.Err("[ERROR > director/slave.go:435] it was not possible to connect with '"+a.Source+"'.")
											} else {
												Conn3ToAgent.Write([]byte(rta))
												Conn3ToAgent.Write([]byte("\n"))
												Conn3ToAgent.Close()
											}
										}
									}
								}
							}
						}
					}
					// Close transmission
					stream = false
				}

				if runner == true {
					// Extract dataset information
					enable	   := values.Dataset[index].Enable
					dataset	   := values.Dataset[index].Name
					getBackup  := values.Dataset[index].Backup
					getClone   := values.Dataset[index].Clone.Enable
					clone	   := values.Dataset[index].Clone.Name
					delClone   := values.Dataset[index].Clone.Delete

					if enable == true {
						// Delete the backup snapshot
						ds, err := zfs.GetDataset(dataset)
						if err != nil {
							w.Err("[ERROR > director/slave.go:464] it was not possible to get the dataset '"+dataset+"'.")
						} else {
							// Delete an existing clone?
							if delClone == true {
								go lib.DeleteClone(clone)
							}

							// Delete backup snapshot
							go lib.DeleteBackup(dataset, ds)

							// Create a backup snapshot?
							if getBackup == true {
								go lib.Backup(dataset, ds)
							}

							// Clone the last snapshot received?
							if getClone == true {
								list, err := ds.Snapshots()
								if err != nil {
									w.Err("[ERROR > director/slave.go:483] it was not possible to access of snapshots list.")
								} else {
									count := len(list)
									_, amount := lib.RealList(count, list, dataset)
									LastSnapshot := list[amount-1].Name
									snap, err := zfs.GetDataset(LastSnapshot)
									if err != nil {
										w.Err("[ERROR > director/slave.go:490] it was not possible to get the snapshot '"+snap.Name+"'.")
									} else {
										go lib.Clone(clone, snap)
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
