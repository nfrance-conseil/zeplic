// Package lib contains: commands.go - snapshot.go - uuid.go
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

	"github.com/pborman/uuid"
)

// Asign a new uuid
func UUID(SnapshotName string) error {
	id := uuid.New()
	args := make([]string, 1, 4)
	args[0] = "zfs"
	args = append(args, "set")
	id = strings.Join([]string{":uuid=", id}, "")
	args = append(args, id)
	args = append(args, SnapshotName)
	idset := strings.Join(args, " ")
	err := exec.Command("sh", "-c", idset).Run()
	return err
}

// Asign an uuid received to snapshot
func ReceiveUUID(id string, SnapshotName string, DestDataset string) {
	check := Before(SnapshotName, "@")
	if check == DestDataset {
		SnapshotName = SnapshotName
	} else {
		SnapshotName = strings.Replace(SnapshotName, check, DestDataset, -1)
	}
	args := make([]string, 1, 4)
	args[0] = "zfs"
	args = append(args, "set")
	id = strings.Join([]string{":uuid=", id}, "")
	args = append(args, id)
	args = append(args, SnapshotName)
	idset := strings.Join(args, " ")
	exec.Command("sh", "-c", idset).Run()
}

// Search the name of snapshot from its uuid
func SearchName(uuid string) string {
	search := fmt.Sprintf("zfs get -rHp -t snapshot -o name,value :uuid | awk '{if ($2 == \"%s\") print $1}'", uuid)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	snapshot := string(out)
	return snapshot
}

// Search the uuid of snapshot from its name
func SearchUUID(SnapshotName string) string {
	search := fmt.Sprintf("zfs get -rHp -t snapshot -o name,value :uuid | awk '{if ($1 == \"%s\") print $2}'", SnapshotName)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	uuid := string(out)
	return uuid
}
