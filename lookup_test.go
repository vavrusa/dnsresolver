package main

import (
	"testing"
)

func TestCheckMX(t *testing.T) {
	dnsServerPort = "8.8.8.8:53"

	job := &Job{Domain: "ripe.net."}
	err := lookup(job, dnsTypes["MX"])

	if err != nil {
		t.Fatal(err)
		return
	}

	if len(job.Results) != 2 {
		t.Error("expected 2 results")
		return
	}

	if job.Duration == 0 {
		t.Error("invalid duration")
		return
	}

	if job.Results[0] != "koko.ripe.net." && job.Results[1] != "koko.ripe.net." {
		t.Errorf("unexpected record: %s", job.Results[0])
	}
}

func TestCheckNoMX(t *testing.T) {
	dnsServerPort = "8.8.8.8:53"

	job := &Job{Domain: "example.com."}
	err := lookup(job, dnsTypes["MX"])

	if err != nil {
		t.Fatal(err)
		return
	}

	if job.Duration == 0 {
		t.Error("invalid duration")
		return
	}

	if len(job.Results) > 0 {
		t.Error("no records expected")
	}
}

func TestCheckA(t *testing.T) {
	dnsServerPort = "8.8.8.8:53"

	job := &Job{Domain: "example.com."}
	err := lookup(job, dnsTypes["A"])

	if err != nil {
		t.Fatal(err)
		return
	}

	if len(job.Results) == 0 {
		t.Error("no records returned")
		return
	}

	if job.Results[0] != "93.184.216.34" {
		t.Error("invalid address returned")
	}
}
