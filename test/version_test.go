package test

import (
	"strings"

	"github.com/nfrance-conseil/zeplic/config"
	"testing"
)

func TestShowVersion(t *testing.T) {
	version := config.ShowVersion()
	if !strings.Contains(version, "zeplic preliminar version:") {
		t.Errorf("ShowVersion() test failed!")
	}
}
