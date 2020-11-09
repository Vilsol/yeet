# Yeet

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/vilsol/yeet/build)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vilsol/yeet)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/vilsol/yeet)

```
Usage:
  yeet [command]

Available Commands:
  help        Help about any command
  serve       Run the webserver

Flags:
      --colors                     Force output with colors
      --expiry                     Use cache expiry
      --expiry-interval duration   Interval between cache GC's (default 10m0s)
      --expiry-time duration       Lifetime of a cache entry (default 1h0m0s)
  -h, --help                       help for yeet
      --host string                Hostname to bind the webserver (default "0.0.0.0")
      --index-file string          The directory default index file (default "index.html")
      --log string                 The log level to output (default "info")
      --paths strings              Paths to serve on the webserver (default [./www])
      --port int                   Port to run the webserver on (default 8080)
      --warmup                     Load all files into memory on startup
      --watch                      Watch filesystem for changes

Use "yeet [command] --help" for more information about a command.
```

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

### With file watching

```
BenchmarkServerGet1ReqPerConnWatch                       8282121              1430 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConnWatch                      11046026              1080 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConnWatch                     15087688               791 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10KReqPerConnWatch                    16619936               733 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet1ReqPerConn10KClientsWatch             8312827              1460 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConn10KClientsWatch            10938738              1097 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConn10KClientsWatch           14515534               811 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet100ReqPerConn10KClientsWatch          16010731               717 ns/op               0 B/op          0 allocs/op
```