package test

import (
	"github.com/nfrance-conseil/zeplic/calendar"
	"testing"
)

func TestRetention(t *testing.T) {
	retention := "24d1w3m1y"
	D, W, M, Y := calendar.Retention(retention)
	if D != 24 || W != 1 || M != 3|| Y != 1 {
		t.Errorf("Retention() test failed!")
	}
}
