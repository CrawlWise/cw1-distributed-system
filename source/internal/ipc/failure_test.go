package ipc

import (
	"context"
	"encoding/binary"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestWeek04_ServerRejectsOversizeTCP(t *testing.T) {
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer ln.Close()
	var calls atomic.Int32
	done := make(chan error, 1)
	go func() { done <- ServeTCP(ctx, ln, "node-b", 100*time.Millisecond, func(Message) { calls.Add(1) }) }()
	c, e := net.DialTimeout("tcp", ln.Addr().String(), 200*time.Millisecond)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(300 * time.Millisecond))
	var h [4]byte
	binary.BigEndian.PutUint32(h[:], MaxMessage+1)
	_, _ = c.Write(h[:])
	var b [1]byte
	n, e := c.Read(b[:])
	if e == nil || n != 0 {
		t.Error("oversize TCP frame was not rejected")
	}
	if calls.Load() != 0 {
		t.Error("rejected frame mutated inbox")
	}
	cancel()
	select {
	case e := <-done:
		if e != nil {
			t.Errorf("serve failed: %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown hung")
	}
}
func TestWeek04_UDPDropAndDeadline(t *testing.T) {
	c, e := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer c.Close()
	done := make(chan error, 1)
	var calls atomic.Int32
	go func() { done <- ServeUDP(ctx, c, "node-b", 100*time.Millisecond, func(Message) { calls.Add(1) }) }()
	remote, e := net.DialTimeout("udp", c.LocalAddr().String(), time.Second)
	if e != nil {
		t.Fatal(e)
	}
	defer remote.Close()
	_ = remote.SetDeadline(time.Now().Add(150 * time.Millisecond))
	_, _ = remote.Write([]byte(`{malformed}`))
	var b [128]byte
	n, e := remote.Read(b[:])
	if e == nil || n != 0 {
		t.Error("malformed UDP request should be dropped")
	}
	if calls.Load() != 0 {
		t.Error("rejected datagram changed inbox")
	}
	cancel()
	select {
	case e := <-done:
		if e != nil {
			t.Errorf("serve: %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("UDP shutdown hung")
	}
}

func TestWeek04_TCPIncompleteFrameTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- ServeTCP(
			ctx,
			ln,
			"node-b",
			100*time.Millisecond,
			nil,
		)
	}()

	conn, err := net.DialTimeout(
		"tcp",
		ln.Addr().String(),
		200*time.Millisecond,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Advertise a 10-byte message but send only 5 bytes.
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], 10)

	if _, err := conn.Write(
		append(header[:], []byte("hello")...),
	); err != nil {
		t.Fatal(err)
	}

	// Server should close the incomplete connection.
	if err := conn.SetReadDeadline(
		time.Now().Add(500 * time.Millisecond),
	); err != nil {
		t.Fatal(err)
	}

	var response [1]byte
	n, err := conn.Read(response[:])

	if err == nil || n != 0 {
		t.Fatalf(
			"expected connection closure, got n=%d err=%v",
			n,
			err,
		)
	}

	// Verify the server closed the connection before
	// the client-side deadline expired.
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		t.Fatal("server failed to enforce read timeout")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("server shutdown: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("TCP server did not shut down")
	}
}

func TestWeek04_TCPServerRecoversAfterMalformedFrame(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var received atomic.Int32

	done := make(chan error, 1)

	go func() {
		done <- ServeTCP(
			ctx,
			ln,
			"node-b",
			300*time.Millisecond,
			func(Message) {
				received.Add(1)
			},
		)
	}()

	// Send an invalid frame.
	conn, err := net.DialTimeout(
		"tcp",
		ln.Addr().String(),
		200*time.Millisecond,
	)
	if err != nil {
		t.Fatal(err)
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], MaxMessage+1)

	if _, err := conn.Write(header[:]); err != nil {
		t.Fatal(err)
	}

	conn.Close()

	// Send a valid message after the malformed frame.
	message := Message{
		Version:   Version,
		RequestID: "recovery-1",
		NodeID:    "node-a",
		Text:      "Server recovery test",
	}

	ack, err := SendTCP(
		ctx,
		ln.Addr().String(),
		message,
		500*time.Millisecond,
	)

	if err != nil {
		t.Fatalf("server did not recover: %v", err)
	}

	if ack.Receiver != "node-b" || !ack.Accepted {
		t.Fatalf("unexpected acknowledgement: %+v", ack)
	}

	if received.Load() != 1 {
		t.Fatalf(
			"expected one accepted message, got %d",
			received.Load(),
		)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server shutdown timed out")
	}
}
