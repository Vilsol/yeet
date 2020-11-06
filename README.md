# Yeet

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
      --log string                 The log level to output (default "info")
      --paths strings              Paths to serve on the webserver (default [./www])
      --port int                   Port to run the webserver on (default 8080)

Use "yeet [command] --help" for more information about a command.
```

## Docker

```
docker run -v /path/to/data:/www -p 8080:8080 ghcr.io/vilsol/yeet:latest
```