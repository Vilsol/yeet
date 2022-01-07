# Yeet ![GitHub Workflow Status](https://img.shields.io/github/workflow/status/vilsol/yeet/build) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vilsol/yeet) ![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/vilsol/yeet)

CLI Usage: [Docs](./docs/yeet.md)

## Features

* Fast
* 0 setup
* Local fs support
* S3 support
* Redis-backed S3 support
* Cuckoo filter for Redis-S3
* File watching (local only)

## Docker

```
docker run -v /path/to/data:/www -p 8080:8080 ghcr.io/vilsol/yeet:latest
```

## Benchmarks (GOMAXPROCS=1)

### Baseline

```
BenchmarkServerGet1ReqPerConn                            8512062              1379 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConn                           11406890              1057 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConn                          15189015               775 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10KReqPerConn                         17068996               698 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet1ReqPerConn10KClients                  8310056              1409 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConn10KClients                 10608926              1058 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConn10KClients                15363962               773 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet100ReqPerConn10KClients               16854955               707 ns/op               0 B/op          0 allocs/op
```

### With cache expiry

```
BenchmarkServerGet1ReqPerConnExpiry                      8677137              1375 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConnExpiry                     11386528              1053 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConnExpiry                    15480867               773 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10KReqPerConnExpiry                   16949194               707 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet1ReqPerConn10KClientsExpiry            8515335              1388 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConn10KClientsExpiry           11266317              1060 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConn10KClientsExpiry          15184057               776 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet100ReqPerConn10KClientsExpiry         16339011               714 ns/op               0 B/op          0 allocs/op
```
