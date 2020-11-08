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
      --expiry-interval duration   Port to run the webserver on (default 10m0s)
      --expiry-memory int          Max memory usage in MB (default 128)
      --expiry-shards int          Cache shard count (default 64)
      --expiry-time duration       Lifetime of a cache entry (default 1h0m0s)
  -h, --help                       help for yeet
      --host string                Hostname to bind the webserver (default "0.0.0.0")
      --index-file string          The directory default index file (default "index.html")
      --log string                 The log level to output (default "info")
      --paths strings              Paths to serve on the webserver (default [./www])
      --port int                   Port to run the webserver on (default 8080)
      --warmup                     Load all files into memory on startup

Use "yeet [command] --help" for more information about a command.
```

## Docker

```
docker run -v /path/to/data:/www -p 8080:8080 ghcr.io/vilsol/yeet:latest
```

## Benchmarks (GOMAXPROCS=1)

### Without cache expiry

```
BenchmarkServerGet1ReqPerConn                            8684911              1345 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConn                           11419969              1046 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConn                          15521960               763 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10KReqPerConn                         17596045               690 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet1ReqPerConn10KClients                  8395502              1386 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet2ReqPerConn10KClients                 11061466              1053 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet10ReqPerConn10KClients                15229926               771 ns/op               0 B/op          0 allocs/op
BenchmarkServerGet100ReqPerConn10KClients               16757032               699 ns/op               0 B/op          0 allocs/op
```

### With cache expiry

```
BenchmarkServerGet1ReqPerConnExpiry                      2419442              4780 ns/op            6176 B/op          2 allocs/op
BenchmarkServerGet2ReqPerConnExpiry                      2912630              4074 ns/op            6176 B/op          2 allocs/op
BenchmarkServerGet10ReqPerConnExpiry                     3479211              3468 ns/op            6176 B/op          2 allocs/op
BenchmarkServerGet10KReqPerConnExpiry                    4027263              2952 ns/op            6176 B/op          2 allocs/op
BenchmarkServerGet1ReqPerConn10KClientsExpiry            1645986              7244 ns/op            6411 B/op          2 allocs/op
BenchmarkServerGet2ReqPerConn10KClientsExpiry            1818912              7026 ns/op            6177 B/op          2 allocs/op
BenchmarkServerGet10ReqPerConn10KClientsExpiry           1960964              6217 ns/op            6177 B/op          2 allocs/op
BenchmarkServerGet100ReqPerConn10KClientsExpiry          2766102              4114 ns/op            6223 B/op          2 allocs/op
```