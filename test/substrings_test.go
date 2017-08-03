package test

import (
	"strings"

	"github.com/nfrance-conseil/zeplic/tools"
	"testing"
)

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
