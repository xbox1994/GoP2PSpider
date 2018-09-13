package bep

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xbox1994/bencode"
	"io"
	"net"
	"time"
)

const (
	perBlock        = 16384
	maxMetadataSize = perBlock * 1024
	extended        = 20
	extHandshake    = 0
)

var (
	ErrExtHeader    = errors.New("metawire: invalid extention header response")
	ErrInvalidPiece = errors.New("metawire: invalid piece response")
	ErrTimeout      = errors.New("metawire: time out")
)

func randomPeerID() string {
	b := make([]byte, 20)
	rand.Read(b)
	return string(b)
}

type Wire struct {
	infohash     string
	from         string
	peerID       string
	conn         *net.TCPConn
	timeout      time.Duration
	metadataSize int
	utMetadata   int
	numOfPieces  int
	pieces       [][]byte
	err          error
}

type option func(w *Wire)

type meta struct {
	data []byte
	err  error
}

func Timeout(t time.Duration) option {
	return func(w *Wire) {
		w.timeout = t
	}
}

func New(infohash string, from string, options ...option) *Wire {
	w := &Wire{
		infohash: infohash,
		from:     from,
		peerID:   randomPeerID(),
		timeout:  5 * time.Second,
	}
	for _, option := range options {
		option(w)
	}
	return w
}

func (w *Wire) Fetch() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
	defer cancel()
	return w.fetch(ctx)
}

func (w *Wire) connect(ctx context.Context) {
	select {
	case <-ctx.Done():
		w.err = ErrExtHeader
		return
	default:
	}
	conn, err := net.DialTimeout("tcp", w.from, w.timeout)
	if err != nil {
		w.err = fmt.Errorf("metawire: connect to remote peer failed: %v", err)
		return
	}
	w.conn = conn.(*net.TCPConn)
}

func (w *Wire) fetch(ctx context.Context) ([]byte, error) {
	w.connect(ctx)
	w.handshake(ctx)
	w.onHandshake(ctx)
	w.extHandshake(ctx)
	if w.err != nil {
		if w.conn != nil {
			w.conn.Close()
		}
		return nil, w.err
	}
	for {
		data, err := w.next(ctx)
		if err != nil {
			return nil, err
		}
		if data[0] != extended {
			continue
		}
		if err := w.onExtended(ctx, data[1], data[2:]); err != nil {
			return nil, err
		}
		if !w.checkDone() {
			continue
		}
		m := bytes.Join(w.pieces, []byte(""))
		sum := sha1.Sum(m)
		if bytes.Equal(sum[:], []byte(w.infohash)) {
			return m, nil
		}
		return nil, errors.New("metawire: metadata checksum mismatch")
	}
}

func (w *Wire) handshake(ctx context.Context) {
	if w.err != nil {
		return
	}
	select {
	case <-ctx.Done():
		w.err = ErrTimeout
		return
	default:
	}
	buf := bytes.NewBuffer(nil)
	buf.Write(w.preHeader())
	buf.WriteString(w.infohash)
	buf.WriteString(w.peerID)
	_, w.err = w.conn.Write(buf.Bytes())
}

func (w *Wire) onHandshake(ctx context.Context) {
	if w.err != nil {
		return
	}
	select {
	case <-ctx.Done():
		w.err = ErrTimeout
		return
	default:
	}
	res, err := w.read(ctx, 68)
	if err != nil {
		w.err = err
		return
	}
	if !bytes.Equal(res[:20], w.preHeader()[:20]) {
		w.err = errors.New("metawire: remote peer not supporting bittorrent protocol")
		return
	}
	if res[25]&0x10 != 0x10 {
		w.err = errors.New("metawire: remote peer not supporting extention protocol")
		return
	}
	if !bytes.Equal(res[28:48], []byte(w.infohash)) {
		w.err = errors.New("metawire: invalid bittorrent header response")
		return
	}
}

func (w *Wire) extHandshake(ctx context.Context) {
	if w.err != nil {
		return
	}
	select {
	case <-ctx.Done():
		w.err = ErrTimeout
		return
	default:
	}
	data := append([]byte{extended, extHandshake}, bencode.Encode(map[string]interface{}{
		"m": map[string]interface{}{
			"ut_metadata": 1,
		},
	})...)
	if err := w.send(ctx, data); err != nil {
		w.err = err
		return
	}
}

func (w *Wire) onExtHandshake(ctx context.Context, payload []byte) error {
	select {
	case <-ctx.Done():
		return ErrTimeout
	default:
	}
	dict, err := bencode.Decode(bytes.NewBuffer(payload))
	if err != nil {
		return ErrExtHeader
	}
	metadataSize, ok := dict["metadata_size"].(int64)
	if !ok {
		return ErrExtHeader
	}
	if metadataSize > maxMetadataSize {
		return errors.New("metawire: metadata_size too long")
	}
	if metadataSize < 0 {
		return errors.New("metawire: negative metadata_size")
	}
	m, ok := dict["m"].(map[string]interface{})
	if !ok {
		return ErrExtHeader
	}
	utMetadata, ok := m["ut_metadata"].(int64)
	if !ok {
		return ErrExtHeader
	}
	w.metadataSize = int(metadataSize)
	w.utMetadata = int(utMetadata)
	w.numOfPieces = w.metadataSize / perBlock
	if w.metadataSize%perBlock != 0 {
		w.numOfPieces++
	}
	w.pieces = make([][]byte, w.numOfPieces)
	for i := 0; i < w.numOfPieces; i++ {
		w.requestPiece(ctx, i)
	}
	return nil
}

func (w *Wire) requestPiece(ctx context.Context, i int) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(extended))
	buf.WriteByte(byte(w.utMetadata))
	buf.Write(bencode.Encode(map[string]interface{}{
		"msg_type": 0,
		"piece":    i,
	}))
	w.send(ctx, buf.Bytes())
}

func (w *Wire) onExtended(ctx context.Context, ext byte, payload []byte) error {
	if ext == 0 {
		if err := w.onExtHandshake(ctx, payload); err != nil {
			return err
		}
	} else {
		piece, index, err := w.onPiece(ctx, payload)
		if err != nil {
			return err
		}
		w.pieces[index] = piece
	}
	return nil
}

func (w *Wire) onPiece(ctx context.Context, payload []byte) ([]byte, int, error) {
	select {
	case <-ctx.Done():
		return nil, -1, ErrTimeout
	default:
	}
	trailerIndex := bytes.Index(payload, []byte("ee")) + 2
	if trailerIndex == 1 {
		return nil, 0, ErrInvalidPiece
	}
	dict, err := bencode.Decode(bytes.NewBuffer(payload[:trailerIndex]))
	if err != nil {
		return nil, 0, ErrInvalidPiece
	}
	peiceIndex, ok := dict["piece"].(int64)
	if !ok || int(peiceIndex) >= w.numOfPieces {
		return nil, 0, ErrInvalidPiece
	}
	msgType, ok := dict["msg_type"].(int64)
	if !ok || msgType != 1 {
		return nil, 0, ErrInvalidPiece
	}
	return payload[trailerIndex:], int(peiceIndex), nil
}

func (w *Wire) checkDone() bool {
	for _, b := range w.pieces {
		if b == nil {
			return false
		}
	}
	return true
}

func (w *Wire) preHeader() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(19)
	buf.WriteString("BitTorrent protocol")
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x01})
	return buf.Bytes()
}

func (w *Wire) next(ctx context.Context) ([]byte, error) {
	data, err := w.read(ctx, 4)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(data)
	data, err = w.read(ctx, size)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (w *Wire) read(ctx context.Context, size uint32) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ErrTimeout
	default:
	}
	buf := bytes.NewBuffer(nil)
	_, err := io.CopyN(buf, w.conn, int64(size))
	if err != nil {
		return nil, fmt.Errorf("metawire: read %d bytes message failed: %v", size, err)
	}
	return buf.Bytes(), nil
}

func (w *Wire) send(ctx context.Context, data []byte) error {
	select {
	case <-ctx.Done():
		return ErrTimeout
	default:
	}
	buf := bytes.NewBuffer(nil)
	length := int32(len(data))
	binary.Write(buf, binary.BigEndian, length)
	buf.Write(data)
	_, err := w.conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("metawire: send message failed: %v", err)
	}
	return nil
}
