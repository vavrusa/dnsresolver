DNS Resolver
============

This is a bulk DNS resolver written in Go that supports DNSSEC.
It is based on [miekg/dns](https://github.com/miekg/unbound) and requires [libunbound](https://unbound.net/documentation/libunbound.html).

## Usage

```
dnsresolver [options] TYPE

Options:
  -workers=32: Number of worker routines
  -tafile="": Path to trusted anchor file for DNSSEC
  -debug=0: Debug level for libunbound

Type can be A, AAAA, MX
```

If you want the resolver to use DNSSEC validation, please provide the root trust anchor (tafile).
You get it with `unbound-anchor -a ./root.key` or from [https://www.iana.org/domains/root/files](IANA).

### Example

```
$ echo "example.com\ngoogle.de" | dnsresolver A AAAA
2015/02/23 03:32:27 Query for A records
2015/02/23 03:32:27 Query for AAAA records
2015/02/23 03:32:27 Using 8 threads
2015/02/23 03:32:27 Starting 32 workers
{"domain":"google.com","results":["74.125.136.139","74.125.136.138","74.125.136.101","74.125.136.102","74.125.136.113","74.125.136.100","2a00:1450:4013:c01::65"],"duration":289,"error":"","security":"insecure"}
{"domain":"example.com","results":["93.184.216.34","2606:2800:220:1:248:1893:25c8:1946"],"duration":503,"error":"","security":"insecure"}
```
