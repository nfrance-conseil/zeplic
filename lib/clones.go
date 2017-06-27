// Package lib contains: clones.go - comnands.go - snapshot.go - uuid.go
//
// Clones searchs the dataset where the snapshot was cloned
//
package lib

import (
	"bytes"
	"fmt"
	"os/exec"
)

// SearchClone searchs the name of the dataset where a snapshot was cloned
func SearchClone(snapshot string) string {
	search := fmt.Sprintf("zfs get -rHp -o name,value clones %s | awk '{if ($1 == \"%s\") print $2}'", snapshot, snapshot)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	clone := string(out)
	return clone
}
