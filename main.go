package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"os"
	"runtime"
)

var (
	pending      = make(chan *Job, 100)
	finished     = make(chan *Job, 100)
	done         = make(chan bool)
	workersCount = 32
	debug        = 0
	tafile       = ""
	dnsTypes     = map[string]uint16{
		"MX":   dns.TypeMX,
		"A":    dns.TypeA,
		"AAAA": dns.TypeAAAA,
	}
)

func main() {
	flag.IntVar(&workersCount, "workers", workersCount, "Number of worker routines")
	flag.StringVar(&tafile, "tafile", tafile, "Path to trusted anchor file for DNSSEC")
	flag.IntVar(&debug, "debug", debug, "Debug level for libunbound")
	// timeout := flag.Int("timeout", 5, "Timeout for a query in seconds")
	flag.Parse()

	queryTypes := []uint16{}
	for _, arg := range flag.Args() {
		if t, ok := dnsTypes[arg]; ok {
			log.Println("Query for", arg, "records")
			queryTypes = append(queryTypes, t)
		} else {
			fmt.Fprintln(os.Stderr, "invalid query type:", arg)
			os.Exit(2)
		}
	}

	if len(queryTypes) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] TYPE\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nType can be A, AAAA, MX\n")
		fmt.Fprintf(os.Stderr, "Example: echo example.com. | %s A AAAA\n", os.Args[0])
		os.Exit(1)
	}

	// Set debug level
	unboundCtx.DebugLevel(debug)

	// Set trust anchor
	if tafile != "" {
		if err := unboundCtx.AddTaFile(tafile); err != nil {
			log.Fatalf("error %s\n", err.Error())
			os.Exit(3)
		}
	}

	// Use all cores
	cpuCount := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuCount)
	log.Println("Using", cpuCount, "threads")
	log.Println("Starting", workersCount, "workers")

	unboundCtx.SetOption("num-threads", string(cpuCount))

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
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		pending <- &Job{Domain: scanner.Text()}
	}
	close(pending)
}

func worker(queryTypes []uint16) {
	for {
		job := <-pending
		if job != nil {
			executeJob(job, queryTypes)
			finished <- job
		} else {
			// no more jobs to do
			finished <- nil
			return
		}

	}
}

func executeJob(job *Job, queryTypes []uint16) {
	for _, q := range queryTypes {
		if err := lookup(job, q); err != nil {
			log.Printf("%s: %s", job.Domain, err)
			job.Error = err.Error()
			return
		}
	}
}

func resultWriter() {
	doneCount := 0
	for doneCount < workersCount {
		job := <-finished
		if job == nil {
			doneCount++
		} else {
			// Serialize and print result
			b, err := json.Marshal(job)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(b))
		}

	}
	done <- true
}
