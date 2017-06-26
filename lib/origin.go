// Package lib contains: commands.go - origin.go - snapshot.go - uuid.go - written.go
//
// Origin gets the origin of filesystem
//
package lib

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Origin gets the origin property
func Origin(SnapshotName string) string {
	search := fmt.Sprintf("zfs get -rHp -t filesystem -o name,value origin | awk '{if ($2 == \"%s\") print $1}'", SnapshotName)
	cmd, _ := exec.Command("sh", "-c", search).Output()
	out := bytes.Trim(cmd, "\x0A")
	origin := string(out)
	return origin
}
