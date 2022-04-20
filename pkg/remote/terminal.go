package remote

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"sync"
)

type Terminal interface {
	ChangeWindow(ws *WinSize) error
	SetSignal(chan error) error
	Close() error
	Stdout() (io.Reader, error)
	Stdin() (io.WriteCloser, error)
	Stderr() (io.Reader, error)
	Wait() error
}

type WinSize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

var pool sync.Map

func GetTerminal(id string) Terminal {
	val, ok := pool.Load(id)
	if !ok {
		return nil
	}
	return val.(Terminal)
}

func SetTerminal(id string, terminal Terminal) {
	pool.Store(id, terminal)
}

func DelTerminal(id string) {
	pool.Delete(id)
}

func GenerateTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}
