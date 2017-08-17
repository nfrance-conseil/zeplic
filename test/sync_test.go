package test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nfrance-conseil/zeplic/lib"
	"testing"
)

func TestResync(t *testing.T) {
	hour, minute, _ := time.Now().Clock()
	minorH := hour-1
	majorH := hour+1
	minor := strconv.Itoa(minorH)
	major := strconv.Itoa(majorH)
	min := strconv.Itoa(minute)
	minor = fmt.Sprintf("%s:%s", minor, min)
	major = fmt.Sprintf("%s:%s", major, min)
	timezone := []string{minor, major}
	resync := lib.Resync(timezone)
	if resync == false {
		t.Errorf("Resync() test failed!")
	}
}
