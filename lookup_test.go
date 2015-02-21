package main

import (
	"testing"
)

func TestCheckMX(t *testing.T) {
	dnsServerPort = "8.8.8.8:53"

	records, duration, err := lookup("ripe.net.")

	if err != nil {
		t.Fatal(err)
		return
	}

	if duration == 0 {
		t.Error("invalid duration")
		return
	}

	if len(records) == 0 {
		t.Error("no records returned")
		return
	}

	if records[0] != "koko.ripe.net." {
		t.Errorf("unexpected record: %s", records[0])
	}
}

func TestCheckNoMX(t *testing.T) {
	dnsServerPort = "8.8.8.8:53"

	records, duration, err := lookup("example.com.")

	if err != nil {
		t.Fatal(err)
		return
	}

	if duration == 0 {
		t.Error("invalid duration")
		return
	}

	if len(records) > 0 {
		t.Error("no records expected")
	}
}
