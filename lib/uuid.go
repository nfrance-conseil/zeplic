// Package lib contains: clones.go - commands.go - snapshot.go - uuid.go
//
// UUID sets an uuid to the snapshot
// Search snapshot name from its uuid and vice versa
//
package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/nfrance-conseil/zeplic/utils"
	"github.com/IgnacioCarbajoVallejo/go-zfs"
	"github.com/pborman/uuid"
)

// UUID asigns a new uuid
func UUID(ds *zfs.Dataset) error {
	id := uuid.New()
	err := ds.SetProperty(":uuid", id)
	return err
}

// ReceiveUUID asigns an uuid received to snapshot
func ReceiveUUID(id string, SnapshotName string, DestDataset string) error {
	check := utils.Before(SnapshotName, "@")
	var name string
	if check == DestDataset {
		name = SnapshotName
	} else {
		name = strings.Replace(SnapshotName, check, DestDataset, -1)
	}
	ds, err := zfs.GetDataset(name)
	if err != nil {
		w.Err("[ERROR] it was not possible to get the snapshot '"+ds.Name+"'.")
	}
	err = ds.SetProperty(":uuid", id)
	return err
}

// SearchName searchs the name of snapshot from its uuid
func SearchName(uuid string) string {
	search := fmt.Sprintf("zfs get -rHp -t snapshot -o name,value :uuid | awk '{if ($2 == \"%s\") print $1}'", uuid)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	snapshot := string(out)
	return snapshot
}

// SearchUUID searchs the uuid of snapshot from its name
func SearchUUID(ds *zfs.Dataset) string {
	uuid, err := ds.GetProperty(":uuid")
	if err != nil {
		w.Err("[ERROR] it was not possible to find the uuid of '"+ds.Name+"'.")
	}
	return uuid
}
