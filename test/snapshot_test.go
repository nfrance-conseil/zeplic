package test

import (
	"fmt"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/lib"
	"testing"
)

func TestCreateTime(t *testing.T) {
	SnapshotName := "tank/test@SNAP_2017-July-01_10:30:00"
	year, month, day, hour, min, sec := lib.CreateTime(SnapshotName)
	if year != 2017 || month != time.July || day != 1 || hour != 10 || min != 30 || sec != 0 {
		t.Errorf("CreateTime() test failed!")
	}
}

func TestDatasetName(t *testing.T) {
	name := lib.DatasetName("tank/test@SNAP")
	if name != "tank/test" {
		t.Errorf("DatasetName() test failed!")
	}
}

func TestInfoKV(t *testing.T) {
	snapPair := "1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e:tank/test@SNAP_2017-July-01_10:30:00#sent"
	uuid, name, flag := lib.InfoKV(snapPair)
	if uuid != "1a2b3c4d-5678-efgh-6789-0z1a2b3c4d5e" || name != "tank/test@SNAP_2017-July-01_10:30:00" || flag != "#sent" {
		t.Errorf("InfoKV() test failed!")
	}
}

func TestPrefix(t *testing.T) {
	SnapshotName := "tank/test@SNAP_2017-July-01_10:30:00"
	snapPrefix := lib.Prefix(SnapshotName)
	if snapPrefix != "SNAP" {
		t.Errorf("Prefix() test failed!")
	}
}

func TestSnapName(t *testing.T) {
	name := lib.SnapName("SNAP")
	year, month, day := time.Now().Date()
	get := fmt.Sprintf("%s_%d-%s-%02d", "SNAP", year, month, day)
	if !strings.Contains(name, get) {
		t.Errorf("SnapName() test failed!")
	}
}

func TestSnapRenamed(t *testing.T) {
	renamed := lib.SnapRenamed("tank/test@SNAP1", "tank/test@SNAP2")
	if renamed == false {
		t.Errorf("SnapRenamed() test failed!")
	}
}
