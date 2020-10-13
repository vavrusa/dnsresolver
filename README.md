DNS Resolver
============

This is a bulk DNS resolver written in Go.
It is based on [miekg/dns](https://github.com/miekg/dns).

## Usage

```
dnsresolver [options] TYPE

Options:
  -server="8.8.8.8": The resolver to ask
  -pps=100:   Query rate
  -tcp:       Use TCP instead of UDP
  -json:      Return results in JSON instead. 
  -timeout=5: Timeout for a query in seconds
  -workers=32: Number of worker routines
  -append-dot=true: Append missing dot to domains

Type can be A, AAAA, MX, NS
```

### Example

```
$ echo "example.com\ngoogle.de" | dnsresolver A AAAA
2015/02/23 03:32:27 Query for A records
2015/02/23 03:32:27 Query for AAAA records
2015/02/23 03:32:27 Using 8 threads
2015/02/23 03:32:27 Starting 32 workers
{id: 1, "domain":"example.com.","results":["93.184.216.34","2606:2800:220:1:248:1893:25c8:1946"],"duration":60,"error":""}
{id: 2, "domain":"google.com.","results":["173.194.113.161","173.194.113.166","173.194.113.165","173.194.113.167","173.194.113.163","173.194.113.164","173.194.113.174","173.194.113.169","173.194.113.162","173.194.113.168","173.194.113.160","2a00:1450:4005:809::1001"],"duration":87,"error":""}
```

