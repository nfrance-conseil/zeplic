package test

import (
	"fmt"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/lib"
	"testing"
)

func TestDatasetName(t *testing.T) {
	name := lib.DatasetName("tank/test@SNAP")
	if name != "tank/test" {
		t.Errorf("DatasetName() test failed!")
	}
}

func TestSnapName(t *testing.T) {
	name := lib.SnapName("SNAP")
	year, month, day := time.Now().Date()
	get := fmt.Sprintf("%s_%d-%s-%02d", "SNAP", year, month, day)
	if strings.Contains(name, get) == false {
		t.Errorf("SnapName() test failed!")
	}
}

func TestSnapBackup(t *testing.T) {
	backup := lib.SnapBackup()
	year, month, day := time.Now().Date()
	get := fmt.Sprintf("%s_%d-%s-%02d", "BACKUP", year, month, day)
	if strings.Contains(backup, get) == false {
		t.Errorf("SnapBackup() test failed!")
	}
}

func TestRenamed(t *testing.T) {
	renamed := lib.Renamed("tank/test@SNAP1", "tank/test@SNAP2")
	if renamed == false {
		t.Errorf("Renamed() test failed!")
	}
}
