package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"
)

var (
	pending       = make(chan *Job, 100)
	finished      = make(chan *Job, 100)
	done          = make(chan bool)
	workersCount  = 32
	dnsServer     = "8.8.8.8"
	dnsPort       = "53"
	dnsServerPort = ""
)

func main() {
	flag.StringVar(&dnsServer, "server", dnsServer, "The resolver to ask")
	flag.IntVar(&workersCount, "workers", workersCount, "Number of worker routines per CPU core")
	timeout := flag.Int("timeout", 5, "Timeout for a query in seconds")
	flag.Parse()

	dnsServerPort = net.JoinHostPort(dnsServer, dnsPort)

	dnsClient.ReadTimeout = time.Duration(*timeout) * time.Second

	// Use all cores
	cpuCount := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuCount)
	log.Println("Using", cpuCount, "threads")
	log.Println("Starting", workersCount, "workers")

	// Start result writer
	go resultWriter()

	// Start workers
	for i := 0; i < workersCount; i++ {
		go worker()
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

func worker() {
	for {
		job := <-pending
		if job != nil {
			executeJob(job)
			finished <- job
		} else {
			// no more jobs to do
			finished <- nil
			return
		}

	}
}

func executeJob(job *Job) {
	results, duration, err := lookup(fmt.Sprintf("%s.", job.Domain))
	job.Duration = int(duration / time.Millisecond)
	if err == nil {
		job.Results = results
	} else {
		job.Error = err.Error()
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
