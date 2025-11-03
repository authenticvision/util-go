package httpmw

import (
	"context"
	"net"
	"net/http"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_clientConnDied_EPIPE(t *testing.T) {
	r := require.New(t)

	err, req := synthesizeEPIPE(t, r)
	r.True(clientConnDied(req, err))

	// same EPIPE but not matching the "server's" local address
	ctx := context.WithValue(context.Background(), http.LocalAddrContextKey, &fakeAddr{})
	req = req.WithContext(ctx)
	r.False(clientConnDied(req, err))
}

type fakeAddr struct{}

func (f fakeAddr) Network() string {
	return "fake"
}

func (f fakeAddr) String() string {
	return "fake:port"
}

func synthesizeEPIPE(t *testing.T, r *require.Assertions) (error, *http.Request) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	r.NoError(err)
	defer func() { _ = ln.Close() }()

	serverDone := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
		close(serverDone)
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	r.NoError(err)
	defer func() { _ = conn.Close() }()

	ctx := context.WithValue(context.Background(), http.LocalAddrContextKey, conn.LocalAddr())
	req := (&http.Request{}).WithContext(ctx)

	<-serverDone

	n, err := conn.Write([]byte("\x00"))
	if err == nil {
		// this is always the case, no idea why, but it doesn't matter here
		r.Equal(1, n)
		t.Log("no error on first write, trying again")
		_, err = conn.Write([]byte("\x00"))
	}
	r.ErrorIs(err, syscall.EPIPE)
	return err, req
}
