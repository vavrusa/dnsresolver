package main

import (
	"errors"
	"time"

	"github.com/miekg/dns"
)

// Returns
func lookup(job *Job, dnsType uint16, conn **dns.Conn) (err error) {
	//result, err := unboundCtx.Resolve(job.Domain, dns.TypeA, dns.ClassINET)
	m := &dns.Msg{}
	m.RecursionDesired = true
	m.SetQuestion(job.Domain, dnsType)

	// execute the query
	var result *dns.Msg
	start := time.Now()
	result, _, err = exchange(m, *conn)
	job.Duration += int(time.Since(start) / time.Millisecond)

	// error or NXDomain rcode?
	if err != nil || result.Rcode == dns.RcodeNameError {
		if *conn != nil {
			// reconnect
			c, _ := dnsClient.Dial(dnsServerPort)
			*conn = c
			// retry without connection
			result, _, err = exchange(m, nil)
		}
		if err != nil {
			return	
		}
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
		case *dns.NS:
			job.Results = append(job.Results, record.Ns)
		}
	}

	return
}

func exchange(msg *dns.Msg, conn *dns.Conn) (*dns.Msg, time.Duration, error) {
	if conn == nil {
		return dnsClient.Exchange(msg, dnsServerPort)
	} else {
		return dnsClient.ExchangeWithConn(msg, conn)
	}
}