package main

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/miekg/unbound"
	"strconv"
	"time"
)

var unboundCtx = unbound.New()

// Returns
func lookup(job *Job, dnsType uint16) (err error) {
	// execute the query
	start := time.Now()
	result, err := unboundCtx.Resolve(job.Domain, dnsType, dns.ClassINET)
	job.Duration += int(time.Since(start) / time.Millisecond)

	if result.Bogus {
		job.Security = fmt.Sprintf("bogus: %s", result.WhyBogus)
	} else if result.Secure {
		job.Security = "secure"
	} else {
		job.Security = "insecure"
	}

	// error or NXDomain rcode?
	if err != nil || result.NxDomain {
		return
	}

	// Other erroneous rcode?
	if result.Rcode != dns.RcodeSuccess {
		err = errors.New(dns.RcodeToString[result.Rcode])
		return
	}

	for i, _ := range result.Data {
		switch record := result.Rr[i].(type) {
		case *dns.MX:
			job.Results = append(job.Results, record.Mx)
		case *dns.A:
			job.Results = append(job.Results, record.A.String())
		case *dns.AAAA:
			job.Results = append(job.Results, record.AAAA.String())
		case *dns.TLSA:
			job.Results = append(job.Results, strconv.Itoa(int(record.Usage))+
				" "+strconv.Itoa(int(record.Selector))+
				" "+strconv.Itoa(int(record.MatchingType))+
				" "+record.Certificate)
		}
	}

	return
}
