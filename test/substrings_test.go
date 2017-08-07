package test

import (
	"strings"

	"github.com/nfrance-conseil/zeplic/tools"
	"testing"
)

func TestArrange(t *testing.T) {
	SnapshotsList := []string{"tank/test@SNAP_2017-July-01_11:00:00", "tank/test@SNAP_2017-June-01_12:00:00"}
	SnapshotsList = tools.Arrange(SnapshotsList)
	if SnapshotsList[0] != "tank/test@SNAP_2017-06-01_12:00:00" || SnapshotsList[1] != "tank/test@SNAP_2017-07-01_11:00:00" {
		t.Errorf("Arrange() test failed!")
	}
}

func TestAfter(t *testing.T) {
	after := tools.After("testing", "t")
	if !strings.Contains(after, "ing") {
		t.Errorf("After() test failed!")
	}
}

func TestBefore(t *testing.T) {
	before := tools.Before("testing", "st")
	if !strings.Contains(before, "te") {
		t.Errorf("Before() test failed!")
	}
}

func TestBetween(t *testing.T) {
	between := tools.Between("testing", "e", "g")
	if !strings.Contains(between, "stin") {
		t.Errorf("Between() test failed!")
	}
}

func TestReverse(t *testing.T) {
	reverse := tools.Reverse("testing", "t")
	if !strings.Contains(reverse, "esting") {
		t.Errorf("Reverse() test failed!")
	}
}
