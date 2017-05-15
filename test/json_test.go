package test

import (
	"github.com/nfrance-conseil/zeplic/config"
	"testing"
)

func TestJson(t *testing.T) {
	count, path, err := config.Json()
	if count <= 0 {
		t.Errorf("Json() test failed!")
	}
	if path != "/usr/local/etc/zeplic.d/config.json" {
		t.Errorf("Json() test failed!")
	}
	if err != nil {
		t.Errorf("Json() test failed!")
	}
}

func TestExtract(t *testing.T) {
	pieces := config.Extract(0)
	if pieces == nil {
		t.Errorf("Extract() test failed!")
	}
}
