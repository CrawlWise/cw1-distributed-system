package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"

	"example.edu/6cc545/cw1/internal/config"
)

const MaxMessage = 4096
const Version = 1

var ErrUnimplemented = errors.New("TODO Week 4: implement IPC")

type Message struct {
	Version   int    `json:"version"`
	RequestID string `json:"request_id"`
	NodeID    string `json:"node_id"`
	Text      string `json:"text"`
}
type Ack struct {
	Version   int    `json:"version"`
	RequestID string `json:"request_id"`
	Receiver  string `json:"receiver"`
	Accepted  bool   `json:"accepted"`
}

func (m Message) Validate() error {
	if m.Version != Version {
		return errors.New("unsupported message version")
	}

	if !config.ValidID(m.RequestID) {
		return errors.New("invalid request ID")
	}

	if !config.ValidID(m.NodeID) {
		return errors.New("invalid node ID")
	}

	if !utf8.ValidString(m.Text) {
		return errors.New("message text is not valid UTF-8")
	}

	textLength := len([]byte(m.Text))
	if textLength < 1 || textLength > 512 {
		return errors.New("message text must contain 1 to 512 bytes")
	}

	return nil
}

func Encode(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	if len(b) == 0 || len(b) > MaxMessage {
		return nil, errors.New("message exceeds limit")
	}
	return b, nil
}
func Decode(data []byte, v any) error {
	if len(data) == 0 || len(data) > MaxMessage {
		return errors.New("invalid message size")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}
		return err
	}

	return nil
}

func WriteFrame(w io.Writer, data []byte) error {
	if len(data) == 0 || len(data) > MaxMessage {
		return errors.New("invalid frame length")
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))

	frame := append(header[:], data...)

	for len(frame) > 0 {
		n, err := w.Write(frame)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		frame = frame[n:]
	}

	return nil
}

func ReadFrame(r io.Reader) ([]byte, error) {
	var header [4]byte

	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header[:])

	if length == 0 || length > MaxMessage {
		return nil, errors.New("invalid frame length")
	}

	data := make([]byte, int(length))

	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return data, nil
}

var _ = binary.BigEndian
var _ = bytes.NewReader
var _ = utf8.ValidString
var _ = config.ValidID
