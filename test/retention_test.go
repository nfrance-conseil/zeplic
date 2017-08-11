package test

import (
	"github.com/nfrance-conseil/zeplic/lib"
	"testing"
)

func TestRetention(t *testing.T) {
	retention := []string{"24 in last day", "3/day in last week", "2/week in last month", "1/month in last year"}
	D, W, M, Y := lib.Retention(retention)
	if D != 24 || W != 3 || M != 2 || Y != 1 {
		t.Errorf("Retention() test failed!")
	}
}
