package main

import (
	"errors"
	"github.com/miekg/dns"
	"time"
)

var dnsClient = &dns.Client{}

// Returns
func lookup(domain string) (records []string, duration time.Duration, err error) {
	m := &dns.Msg{}
	m.RecursionDesired = true
	m.SetQuestion(domain, dns.TypeMX)

	result := &dns.Msg{}

	// execute the query
	start := time.Now()
	result, _, err = dnsClient.Exchange(m, dnsServerPort)
	duration = time.Since(start)

	// error or NXDomain rcode?
	if err != nil || result.Rcode == dns.RcodeNameError {
		return
	}

	// Other erroneous rcode?
	if result.Rcode != dns.RcodeSuccess {
		err = errors.New(dns.RcodeToString[result.Rcode])
		return
	}

	// Add addresses to result
	for _, a := range result.Answer {
		if record, ok := a.(*dns.MX); ok {
			records = append(records, record.Mx)
		}
	}

	return
}
