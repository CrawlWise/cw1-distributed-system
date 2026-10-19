package ipc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type OnMessage func(Message)

func deadline(ctx context.Context, timeout time.Duration) time.Time {
	d := time.Now().Add(timeout)
	if cd, ok := ctx.Deadline(); ok && cd.Before(d) {
		return cd
	}
	return d
}
func makeAck(id string, m Message) Ack {
	return Ack{Version: Version, RequestID: m.RequestID, Receiver: id, Accepted: true}
}
func checkAck(a Ack, m Message) error {
	if a.Version != Version || a.RequestID != m.RequestID || !a.Accepted || a.Receiver == "" {
		return errors.New("invalid or unrelated acknowledgement")
	}
	return nil
}
func ServeTCP(
	ctx context.Context,
	ln net.Listener,
	id string,
	timeout time.Duration,
	on OnMessage,
) error {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	if err := ctx.Err(); err != nil {
		return nil
	}

	// Maximum 32 concurrent connection handlers.
	slots := make(chan struct{}, 32)

	var wg sync.WaitGroup

	// Cancellation must interrupt Accept().
	stop := context.AfterFunc(ctx, func() {
		_ = ln.Close()
	})
	defer stop()

	defer wg.Wait()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("TCP accept: %w", err)
		}

		if ctx.Err() != nil {
			_ = conn.Close()
			return nil
		}

		// Do not allow an unlimited number of active handlers.
		select {
		case slots <- struct{}{}:
		case <-ctx.Done():
			_ = conn.Close()
			return nil
		}

		wg.Add(1)

		go func(c net.Conn) {
			defer wg.Done()
			defer func() {
				<-slots
				_ = c.Close()
			}()

			stopConn := context.AfterFunc(ctx, func() {
				_ = c.Close()
			})
			defer stopConn()

			if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
				return
			}

			data, err := ReadFrame(c)
			if err != nil {
				log.Printf("TCP frame rejected: %v", err)
				return
			}

			var message Message

			if err := Decode(data, &message); err != nil {
				log.Printf("TCP JSON rejected: %v", err)
				return
			}

			if err := message.Validate(); err != nil {
				log.Printf("TCP message rejected: %v", err)
				return
			}

			if ctx.Err() != nil {
				return
			}

			if on != nil {
				on(message)
			}

			ack := makeAck(id, message)

			response, err := Encode(ack)
			if err != nil {
				return
			}

			if err := WriteFrame(c, response); err != nil {
				log.Printf("TCP acknowledgement failed: %v", err)
			}
		}(conn)
	}
}
func SendTCP(
	ctx context.Context,
	address string,
	m Message,
	timeout time.Duration,
) (Ack, error) {
	var ack Ack

	if err := ctx.Err(); err != nil {
		return ack, err
	}

	if err := m.Validate(); err != nil {
		return ack, err
	}

	data, err := Encode(m)
	if err != nil {
		return ack, err
	}

	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	operationDeadline := deadline(ctx, timeout)

	opCtx, cancel := context.WithDeadline(ctx, operationDeadline)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(
		opCtx, "tcp", address,
	)
	if err != nil {
		return ack, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(operationDeadline); err != nil {
		return ack, err
	}

	// Closing the socket interrupts blocked I/O on cancellation.
	stop := context.AfterFunc(opCtx, func() {
		_ = conn.Close()
	})
	defer stop()

	if err := WriteFrame(conn, data); err != nil {
		return ack, err
	}

	response, err := ReadFrame(conn)
	if err != nil {
		return ack, err
	}

	if err := Decode(response, &ack); err != nil {
		return Ack{}, err
	}

	if err := checkAck(ack, m); err != nil {
		return Ack{}, err
	}

	return ack, nil
}

func ServeUDP(
	ctx context.Context,
	c *net.UDPConn,
	id string,
	timeout time.Duration,
	on OnMessage,
) error {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	if err := ctx.Err(); err != nil {
		return nil
	}

	stop := context.AfterFunc(ctx, func() {
		_ = c.Close()
	})
	defer stop()

	buffer := make([]byte, MaxMessage+1)

	for {
		n, sender, err := c.ReadFromUDP(buffer)

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return fmt.Errorf("UDP read: %w", err)
		}

		if n == 0 || n > MaxMessage {
			log.Printf("UDP datagram rejected: invalid size %d", n)
			continue
		}

		var message Message

		if err := Decode(buffer[:n], &message); err != nil {
			log.Printf("UDP JSON rejected: %v", err)
			continue
		}

		if err := message.Validate(); err != nil {
			log.Printf("UDP message rejected: %v", err)
			continue
		}

		if ctx.Err() != nil {
			return nil
		}

		if on != nil {
			on(message)
		}

		ack := makeAck(id, message)

		response, err := Encode(ack)
		if err != nil {
			log.Printf("UDP acknowledgement encoding failed: %v", err)
			continue
		}

		// Limit how long sending the acknowledgement can block.
		if err := c.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}

		nWritten, err := c.WriteToUDP(response, sender)

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			log.Printf("UDP acknowledgement failed: %v", err)
			continue
		}

		if nWritten != len(response) {
			log.Printf("UDP acknowledgement incomplete")
		}
	}
}

func SendUDP(
	ctx context.Context,
	address string,
	m Message,
	timeout time.Duration,
) (Ack, error) {
	var ack Ack

	if err := ctx.Err(); err != nil {
		return ack, err
	}

	if err := m.Validate(); err != nil {
		return ack, err
	}

	data, err := Encode(m)
	if err != nil {
		return ack, err
	}

	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	operationDeadline := deadline(ctx, timeout)

	opCtx, cancel := context.WithDeadline(ctx, operationDeadline)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(
		opCtx, "udp", address,
	)
	if err != nil {
		return ack, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(operationDeadline); err != nil {
		return ack, err
	}

	// Interrupt blocked reads if the caller cancels.
	stop := context.AfterFunc(opCtx, func() {
		_ = conn.Close()
	})
	defer stop()

	// UDP sends one JSON message in one datagram.
	n, err := conn.Write(data)
	if err != nil {
		return ack, err
	}

	if n != len(data) {
		return ack, io.ErrShortWrite
	}

	// Extra byte allows detection of oversized datagrams.
	buffer := make([]byte, MaxMessage+1)

	n, err = conn.Read(buffer)
	if err != nil {
		return ack, err
	}

	if n == 0 || n > MaxMessage {
		return ack, errors.New("invalid UDP response size")
	}

	if err := Decode(buffer[:n], &ack); err != nil {
		return Ack{}, err
	}

	if err := checkAck(ack, m); err != nil {
		return Ack{}, err
	}

	return ack, nil
}

var _ = fmt.Errorf
var _ = log.Printf
var _ sync.WaitGroup
var _ = deadline
var _ = makeAck
var _ = checkAck
