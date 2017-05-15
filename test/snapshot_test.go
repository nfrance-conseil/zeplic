package test

import (
	"fmt"
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/api"
	"testing"
)

func TestSnapName(t *testing.T) {
	name := api.SnapName("SNAP")
	year, month, day := time.Now().Date()
	get := fmt.Sprintf("%s_%d-%s-%02d", "SNAP", year, month, day)
	if strings.Contains(name, get) == false {
		t.Errorf("SnapName() test failed!")
	}
}

func TestSnapBackup(t *testing.T) {
	backup := api.SnapBackup()
	year, month, day := time.Now().Date()
	get := fmt.Sprintf("%s_%d-%s-%02d", "BACKUP", year, month, day)
	if strings.Contains(backup, get) == false {
		t.Errorf("SnapBackup() test failed!")
	}
}
