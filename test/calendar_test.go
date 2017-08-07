package test

import (
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/tools"
	"testing"
)

func TestLengthMonth(t *testing.T) {
	leap, length := tools.LengthMonth(2017, time.July)
	if leap == true || !strings.Contains(length, "short") {
		t.Errorf("LengthMonth() test failed!")
	}
}

func TestNameMonth(t *testing.T) {
	nameMonth := tools.NameMonth("July")
	if nameMonth != time.July {
		t.Errorf("NameMonth() test failed!")
	}
}

func TestNameMonthZero(t *testing.T) {
	nameString := tools.NameMonthZero("07")
	if nameString != time.July {
		t.Errorf("NameMonthZero() test failed!")
	}
}

func TestNameMonthInt(t *testing.T) {
	nameInt := tools.NameMonthInt(7)
	if nameInt != time.July {
		t.Errorf("NameMonthInt() test failed!")
	}
}

func TestNameWeek(t *testing.T) {
	nameString := tools.NameWeek("Monday")
	if nameString != time.Monday {
		t.Errorf("NameWeek() test failed!")
	}
}

func TestNameWeekInt(t *testing.T) {
	nameInt := tools.NameWeekInt(1)
	if nameInt != time.Monday {
		t.Errorf("NameWeekInt() test failed!")
	}
}

func TestNumberMonth(t *testing.T) {
	snapName := "tank/test@SNAP_2017-July-01_10:30:00"
	snapName = tools.NumberMonth(snapName)
	if snapName != "tank/test@SNAP_2017-07-01_10:30:00" {
		t.Errorf("NumberMonth() test failed!")
	}
}

func TestNumberMonthReverse(t *testing.T) {
	snapName := "tank/test@SNAP_2017-07-01_10:30:00"
	snapName = tools.NumberMonthReverse(snapName)
	if snapName != "tank/test@SNAP_2017-July-01_10:30:00" {
		t.Errorf("NumberMonthReverse() test failed!")
	}
}
