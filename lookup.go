package main

import (
	"errors"
	"github.com/miekg/dns"
	"time"
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
		// TODO make this more DRY

		if record, ok := a.(*dns.MX); ok {
			job.Results = append(job.Results, record.Mx)
		}
		if record, ok := a.(*dns.A); ok {
			job.Results = append(job.Results, record.A.String())
		}
		if record, ok := a.(*dns.AAAA); ok {
			job.Results = append(job.Results, record.AAAA.String())
		}
	}

	return
}
