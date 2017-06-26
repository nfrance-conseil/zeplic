// Package lib contains: commands.go - origin.go - snapshot.go - uuid.go - written.go
//
// Written finds the size of data written after take a snapshot
//
package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os/exec"
)

// HumanReadable returns a value in bytes to digital units
func HumanReadable(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	if s < 1024 {
		return fmt.Sprintf("%d bytes", s)
	}
	logn := math.Log(float64(s)) / math.Log(1024)
	e := math.Floor(logn)
	suffix := sizes[int(e)]
	value := math.Floor(float64(s)/math.Pow(1024, e)*10+0.5) / 10
	var str string
	if value == math.Trunc(value) {
		str = fmt.Sprintf("%.0f %s", value, suffix)
	} else {
		str = fmt.Sprintf("%.1f %s", value, suffix)
	}
	return str
}

// Written returns the amount of data written in dataset
func Written(uuid string) int64 {
	SnapshotName := SearchName(uuid)
	size := fmt.Sprintf("zfs get -rHp -t snapshot -o name,value written | awk '{if ($1 == \"%s\") print $2}'", SnapshotName)
	cmd, _ := exec.Command("sh", "-c", size).Output()
	out := bytes.Trim(cmd, "\x0A")
	written, _ := binary.Varint(out)
	return written
}
