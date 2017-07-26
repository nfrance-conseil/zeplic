package test

import (
	"strings"
	"time"

	"github.com/nfrance-conseil/zeplic/calendar"
	"testing"
)

func TestLengthMonth(t *testing.T) {
	leap, length := calendar.LengthMonth(2017, time.July)
	if leap == true || !strings.Contains(length, "short") {
		t.Errorf("LengthMonth() test failed!")
	}
}

func TestNameMonth(t *testing.T) {
	nameMonth := calendar.NameMonth("July")
	if nameMonth != time.July {
		t.Errorf("NameMonth() test failed!")
	}
}

func TestNameMonthZero(t *testing.T) {
	nameString := calendar.NameMonthZero("07")
	if nameString != time.July {
		t.Errorf("NameMonthZero() test failed!")
	}
}

func TestNameMonthInt(t *testing.T) {
	nameInt := calendar.NameMonthInt(7)
	if nameInt != time.July {
		t.Errorf("NameMonthInt() test failed!")
	}
}

func TestNameWeek(t *testing.T) {
	nameString := calendar.NameWeek("Monday")
	if nameString != time.Monday {
		t.Errorf("NameWeek() test failed!")
	}
}

func TestNameWeekInt(t *testing.T) {
	nameInt := calendar.NameWeekInt(1)
	if nameInt != time.Monday {
		t.Errorf("NameWeekInt() test failed!")
	}
}

func TestNumberMonth(t *testing.T) {
	snapName := "tank/test@SNAP_2017-July-01_10:30:00"
	snapName = calendar.NumberMonth(snapName)
	if snapName != "tank/test@SNAP_2017-07-01_10:30:00" {
		t.Errorf("NumberMonth() test failed!")
	}
}
