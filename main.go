package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"runtime"
	"time"
	"sync"
	"strings"
)

var (
	pending       = make(chan *Job, 100)
	finished      = make(chan *Job, 100)
	done          = make(chan bool)
	workersCount  = 32
	appendDot     = true
	packetsPerSecond = 100
	sendingDelay  = 10 * time.Millisecond
	sendLock      = sync.Mutex{}
	dnsServer     = "1.1.1.1"
	dnsPort       = "53"
	useTcp        = false       
	formatJson    = false
	dnsServerPort = ""
	dnsClient     = &dns.Client{}
	dnsTypes      = map[string]uint16{
		"MX":   dns.TypeMX,
		"A":    dns.TypeA,
		"AAAA": dns.TypeAAAA,
		"NS": dns.TypeNS,
	}
)

func main() {
	flag.StringVar(&dnsServer, "server", dnsServer, "The resolver to ask")
	flag.IntVar(&workersCount, "workers", workersCount, "Number of worker routines")
	flag.BoolVar(&appendDot, "append-dot", appendDot, "Append missing dot to domains")
	flag.IntVar(&packetsPerSecond, "pps", 100,
		"Send up to PPS DNS queries per second")
	flag.BoolVar(&formatJson, "json", formatJson, "Return as JSON instead")
	flag.BoolVar(&useTcp, "tcp", useTcp, "Use TCP instead")
	timeout := flag.Int("timeout", 5, "Timeout for a query in seconds")
	flag.Parse()

	queryTypes := []uint16{}
	for _, arg := range flag.Args() {
		if t, ok := dnsTypes[arg]; ok {
			log.Println("Query for", arg, "records")
			queryTypes = append(queryTypes, t)
		} else {
			fmt.Fprintln(os.Stderr, "invalid query type:", arg)
			os.Exit(1)
		}
	}

	if len(queryTypes) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] TYPE\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nType can be A, AAAA, MX, NS\n")
		fmt.Fprintf(os.Stderr, "Example: echo example.com. | %s A AAAA\n", os.Args[0])
		os.Exit(1)
	}

	sendingDelay  = time.Duration(1000000000/packetsPerSecond) * time.Nanosecond
	dnsServerPort = net.JoinHostPort(dnsServer, dnsPort)
	if useTcp {
		dnsClient.ReadTimeout = 30 * time.Second
	} else {
		dnsClient.ReadTimeout = time.Duration(*timeout) * time.Second
	}
	

	// Use all cores
	cpuCount := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuCount)
	fmt.Fprintln(os.Stderr, "Hi Sergi, GO LAKERS!")
	fmt.Fprintln(os.Stderr, "Using", cpuCount, "threads")
	fmt.Fprintln(os.Stderr, "Starting", workersCount, "workers")

	// Start result writer
	go resultWriter()

	// Start workers
	for i := 0; i < workersCount; i++ {
		go worker(queryTypes)
	}

	createJobs()

	// wait for resultWriter to finish
	<-done
}

func createJobs() {
	id := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		domain := scanner.Text()
		if appendDot && !strings.HasSuffix(domain, ".") {
			domain = fmt.Sprintf("%s.", domain)
		}
		id += 1
		pending <- &Job{Id: id, Domain: domain}
	}
	close(pending)
}

func worker(queryTypes []uint16) {
	var conn *dns.Conn
	if useTcp {
		c, err := dnsClient.Dial(dnsServerPort)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to open connection", err)
		}

		conn = c 
	}

	for {
		job := <-pending
		if job != nil {
			executeJob(job, queryTypes, &conn)
			finished <- job
		} else {
			// no more jobs to do
			finished <- nil
			return
		}

	}
}

func executeJob(job *Job, queryTypes []uint16, conn **dns.Conn) {
	for _, q := range queryTypes {
		ratelimit()
		if err := lookup(job, q, conn); err != nil {
			log.Printf("%s: %s\n", job.Domain, err)
			job.Error = err.Error()
			return
		}
	}
}

func ratelimit() {
	sendLock.Lock()
	time.Sleep(sendingDelay)
	sendLock.Unlock()
}

func resultWriter() {
	doneCount := 0
	for doneCount < workersCount {
		job := <-finished
		if job == nil {
			doneCount++
		} else {
			// Serialize and print result
			if formatJson {
				b, err := json.Marshal(job)
				if err != nil {
					panic(err)
				}
				fmt.Println(string(b))
			} else {
				fmt.Printf("%d,%s,%s\n", job.Id, job.Domain, strings.Join(job.Results, ","))
			}
		}

	}
	done <- true
}
