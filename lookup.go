package main

import (
	"errors"
	"time"

	"github.com/miekg/dns"
)

// Returns
func lookup(job *Job, dnsType uint16) (err error) {
	//result, err := unboundCtx.Resolve(job.Domain, dns.TypeA, dns.ClassINET)
	m := &dns.Msg{}
	m.RecursionDesired = true
	m.SetQuestion(job.Domain, dnsType)

	// execute the query
	start := time.Now()
	result, _, err := dnsClient.Exchange(m, dnsServerPort)
	job.Duration += int(time.Since(start) / time.Millisecond)

	// error or NXDomain rcode?
	if err != nil || result.Rcode == dns.RcodeNameError {
		return
	}

	// Other erroneous rcode?
	if result.Rcode != dns.RcodeSuccess {
		err = errors.New(dns.RcodeToString[result.Rcode])
		return
	}

	for _, a := range result.Answer {
		switch record := a.(type) {
		case *dns.MX:
			job.Results = append(job.Results, record.Mx)
		case *dns.A:
			job.Results = append(job.Results, record.A.String())
		case *dns.AAAA:
			job.Results = append(job.Results, record.AAAA.String())
		}
	}

	return
}
