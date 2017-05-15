package test

import (
	"github.com/nfrance-conseil/zeplic/config"
	"testing"
)

func TestLogCreate(t *testing.T) {
	err := config.LogCreate()
	if err != nil {
		t.Errorf("LogCreate() test failed!")
	}
}
/*
func TestLogBook(t *testing.T) {
	sysLog, err := config.LogBook()
	if sysLog == nil {
		t.Errorf("LogBook() test failed!")
	}
	if err != nil {
		t.Errorf("LogBook() test failed!")
	}
}
*/
