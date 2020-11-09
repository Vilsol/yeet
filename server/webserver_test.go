package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const getRequest = "GET /webserver_test.go HTTP/1.1\r\nHost: google.com\r\n\r\n"

func init() {
	viper.Set("paths", []string{"."})
	log.SetLevel(log.FatalLevel)
}

func benchmarkServerGet(b *testing.B, clientsCount, requestsPerConn int, expiry bool, watch bool) {
	if expiry {
		viper.Set("expiry", true)
		viper.Set("expiry.time", time.Minute*10)
		viper.Set("expiry.interval", time.Minute*60)
	}

	if watch {
		viper.Set("watch", true)
	}

	ws, err := GetWebserver()

	if err != nil {
		panic(err)
	}

	s := &fasthttp.Server{
		Handler:     ws.HandleFastHTTP,
		Concurrency: 16 * clientsCount,
	}

	b.ReportAllocs()
	b.ResetTimer()
	benchmarkServer(b, s, clientsCount, requestsPerConn, getRequest)
}

func benchmarkServer(b *testing.B, s *fasthttp.Server, clientsCount, requestsPerConn int, request string) {
	ln := newFakeListener(b.N, clientsCount, requestsPerConn, request)
	ch := make(chan struct{})

	go func() {
		_ = s.Serve(ln)
		ch <- struct{}{}
	}()

	<-ln.done

	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		b.Fatalf("Server.Serve() didn't stop")
	}
}

type fakeServerConn struct {
	net.TCPConn
	ln            *fakeListener
	requestsCount int
	pos           int
	closed        uint32
}

func (c *fakeServerConn) Read(b []byte) (int, error) {
	nn := 0
	reqLen := len(c.ln.request)
	for len(b) > 0 {
		if c.requestsCount == 0 {
			if nn == 0 {
				return 0, io.EOF
			}
			return nn, nil
		}
		pos := c.pos % reqLen
		n := copy(b, c.ln.request[pos:])
		b = b[n:]
		nn += n
		c.pos += n
		if n+pos == reqLen {
			c.requestsCount--
		}
	}
	return nn, nil
}

func (c *fakeServerConn) Write(b []byte) (int, error) {
	return len(b), nil
}

var fakeAddr = net.TCPAddr{
	IP:   []byte{1, 2, 3, 4},
	Port: 12345,
}

func (c *fakeServerConn) RemoteAddr() net.Addr {
	return &fakeAddr
}

func (c *fakeServerConn) Close() error {
	if atomic.AddUint32(&c.closed, 1) == 1 {
		c.ln.ch <- c
	}
	return nil
}

func (c *fakeServerConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *fakeServerConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type fakeListener struct {
	lock            sync.Mutex
	requestsCount   int
	requestsPerConn int
	request         []byte
	ch              chan *fakeServerConn
	done            chan struct{}
	closed          bool
}

func (ln *fakeListener) Accept() (net.Conn, error) {
	ln.lock.Lock()
	if ln.requestsCount == 0 {
		ln.lock.Unlock()
		for len(ln.ch) < cap(ln.ch) {
			time.Sleep(10 * time.Millisecond)
		}
		ln.lock.Lock()
		if !ln.closed {
			close(ln.done)
			ln.closed = true
		}
		ln.lock.Unlock()
		return nil, io.EOF
	}
	requestsCount := ln.requestsPerConn
	if requestsCount > ln.requestsCount {
		requestsCount = ln.requestsCount
	}
	ln.requestsCount -= requestsCount
	ln.lock.Unlock()

	c := <-ln.ch
	c.requestsCount = requestsCount
	c.closed = 0
	c.pos = 0

	return c, nil
}

func (ln *fakeListener) Close() error {
	return nil
}

func (ln *fakeListener) Addr() net.Addr {
	return &fakeAddr
}

func newFakeListener(requestsCount, clientsCount, requestsPerConn int, request string) *fakeListener {
	ln := &fakeListener{
		requestsCount:   requestsCount,
		requestsPerConn: requestsPerConn,
		request:         []byte(request),
		ch:              make(chan *fakeServerConn, clientsCount),
		done:            make(chan struct{}),
	}
	for i := 0; i < clientsCount; i++ {
		ln.ch <- &fakeServerConn{
			ln: ln,
		}
	}
	return ln
}

// Baseline
func BenchmarkServerGet1ReqPerConn(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 1, false, false)
}

func BenchmarkServerGet2ReqPerConn(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 2, false, false)
}

func BenchmarkServerGet10ReqPerConn(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 10, false, false)
}

func BenchmarkServerGet10KReqPerConn(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 10000, false, false)
}

func BenchmarkServerGet1ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 1, false, false)
}

func BenchmarkServerGet2ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 2, false, false)
}

func BenchmarkServerGet10ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 10, false, false)
}

func BenchmarkServerGet100ReqPerConn10KClients(b *testing.B) {
	benchmarkServerGet(b, 10000, 100, false, false)
}

// With expiry
func BenchmarkServerGet1ReqPerConnExpiry(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 1, true, false)
}

func BenchmarkServerGet2ReqPerConnExpiry(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 2, true, false)
}

func BenchmarkServerGet10ReqPerConnExpiry(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 10, true, false)
}

func BenchmarkServerGet10KReqPerConnExpiry(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 10000, true, false)
}

func BenchmarkServerGet1ReqPerConn10KClientsExpiry(b *testing.B) {
	benchmarkServerGet(b, 10000, 1, true, false)
}

func BenchmarkServerGet2ReqPerConn10KClientsExpiry(b *testing.B) {
	benchmarkServerGet(b, 10000, 2, true, false)
}

func BenchmarkServerGet10ReqPerConn10KClientsExpiry(b *testing.B) {
	benchmarkServerGet(b, 10000, 10, true, false)
}

func BenchmarkServerGet100ReqPerConn10KClientsExpiry(b *testing.B) {
	benchmarkServerGet(b, 10000, 100, true, false)
}

// With watch
func BenchmarkServerGet1ReqPerConnWatch(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 1, false, true)
}

func BenchmarkServerGet2ReqPerConnWatch(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 2, false, true)
}

func BenchmarkServerGet10ReqPerConnWatch(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 10, false, true)
}

func BenchmarkServerGet10KReqPerConnWatch(b *testing.B) {
	benchmarkServerGet(b, runtime.NumCPU(), 10000, false, true)
}

func BenchmarkServerGet1ReqPerConn10KClientsWatch(b *testing.B) {
	benchmarkServerGet(b, 10000, 1, false, true)
}

func BenchmarkServerGet2ReqPerConn10KClientsWatch(b *testing.B) {
	benchmarkServerGet(b, 10000, 2, false, true)
}

func BenchmarkServerGet10ReqPerConn10KClientsWatch(b *testing.B) {
	benchmarkServerGet(b, 10000, 10, false, true)
}

func BenchmarkServerGet100ReqPerConn10KClientsWatch(b *testing.B) {
	benchmarkServerGet(b, 10000, 100, false, true)
}
