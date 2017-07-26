package test

import (
	"strings"

	"github.com/nfrance-conseil/zeplic/utils"
	"testing"
)

func TestAfter(t *testing.T) {
	after := utils.After("testing", "t")
	if !strings.Contains(after, "ing") {
		t.Errorf("After() test failed!")
	}
}

func TestBefore(t *testing.T) {
	before := utils.Before("testing", "st")
	if !strings.Contains(before, "te") {
		t.Errorf("Before() test failed!")
	}
}

func TestBetween(t *testing.T) {
	between := utils.Between("testing", "e", "g")
	if !strings.Contains(between, "stin") {
		t.Errorf("Between() test failed!")
	}
}

func TestReverse(t *testing.T) {
	reverse := utils.Reverse("testing", "t")
	if !strings.Contains(reverse, "esting") {
		t.Errorf("Reverse() test failed!")
	}
}
