package mail

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestCancellationStopsStalledSMTP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- (SMTP{Addr: listener.Addr().String(), From: "test@example.test"}).Send(ctx, "fixture@example.test", "12345678")
	}()
	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(time.Second):
		t.Fatal("SMTP did not connect")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled SMTP succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("SMTP ignored cancellation while awaiting greeting")
	}
}
