package ipc

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

type fragmentReader struct{ r io.Reader }

func (f fragmentReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return f.r.Read(p)
}

type shortWriter struct{ b bytes.Buffer }

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) > 2 {
		p = p[:2]
	}
	return w.b.Write(p)
}
func TestWeek04_FrameFragmentation(t *testing.T) {
	var w shortWriter
	want := []byte(`{"text":"fragmented"}`)
	if e := WriteFrame(&w, want); e != nil {
		t.Fatal(e)
	}
	got, e := ReadFrame(fragmentReader{&w.b})
	if e != nil || !bytes.Equal(got, want) {
		t.Fatalf("frame lost bytes: %q %v", got, e)
	}
}
func TestWeek04_FrameLimits(t *testing.T) {
	for _, n := range []uint32{0, MaxMessage + 1} {
		var h [4]byte
		binary.BigEndian.PutUint32(h[:], n)
		if _, e := ReadFrame(bytes.NewReader(h[:])); e == nil {
			t.Fatalf("length %d accepted", n)
		}
	}
	if e := WriteFrame(io.Discard, bytes.Repeat([]byte{'x'}, MaxMessage+1)); e == nil {
		t.Fatal("oversized write accepted")
	}
}
func TestWeek04_InvalidMessage(t *testing.T) {
	valid := Message{1, "r1", "node-a", "hello"}
	if e := valid.Validate(); e != nil {
		t.Fatal(e)
	}
	valid.Version = 2
	if valid.Validate() == nil {
		t.Fatal("version 2 accepted")
	}
	for _, s := range []string{`{bad}`, `{"surprise":1}`, `{}{}`, strings.Repeat("x", MaxMessage+1)} {
		var m Message
		if Decode([]byte(s), &m) == nil {
			t.Fatalf("invalid data accepted: %.50s", s)
		}
	}
}

func TestWeek04_MessageValidationBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "empty text",
			message: Message{
				Version:   1,
				RequestID: "r1",
				NodeID:    "node-a",
				Text:      "",
			},
		},
		{
			name: "invalid request ID",
			message: Message{
				Version:   1,
				RequestID: "invalid id!",
				NodeID:    "node-a",
				Text:      "hello",
			},
		},
		{
			name: "text exceeds 512 bytes",
			message: Message{
				Version:   1,
				RequestID: "r1",
				NodeID:    "node-a",
				Text:      strings.Repeat("a", 513),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.message.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestWeek04_IncompleteFrames(t *testing.T) {
	t.Run("incomplete header", func(t *testing.T) {
		_, err := ReadFrame(bytes.NewReader([]byte{0, 0}))
		if err == nil {
			t.Fatal("expected error for incomplete header")
		}
	})

	t.Run("incomplete body", func(t *testing.T) {
		var header [4]byte
		binary.BigEndian.PutUint32(header[:], 10)

		frame := append(header[:], []byte("hello")...)

		_, err := ReadFrame(bytes.NewReader(frame))
		if err == nil {
			t.Fatal("expected error for incomplete body")
		}
	})
}
