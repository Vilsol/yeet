package server

import (
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"net"
	"testing"
)

var listener *fasthttputil.InmemoryListener
var requestURI string

func init() {
	viper.Set("paths", []string{"."})

	cache := make(map[string]*CachedInstance)

	if err := IndexCache(cache); err != nil {
		panic(err)
	}

	for key := range cache {
		requestURI = "http://foo.bar" + key
		break
	}

	ws := &Webserver{
		Cache: cache,
	}

	server := fasthttp.Server{
		Handler: ws.HandleFastHTTP,
	}

	listener = fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()
}

func BenchmarkWebserver(b *testing.B) {
	c := &fasthttp.HostClient{
		Dial: func(addr string) (net.Conn, error) {
			return listener.Dial()
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var req fasthttp.Request
			req.SetRequestURI(requestURI)
			var resp fasthttp.Response

			if err := c.Do(&req, &resp); err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
		}
	})
}
