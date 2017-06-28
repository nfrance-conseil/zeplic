package test

import (
	"github.com/nfrance-conseil/zeplic/config"
	"testing"
)

func TestPid(t *testing.T) {
	err := config.Pid()
	if err != nil {
		t.Errorf("Pid() test failed!")
	}
}
