package dataTypes

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type DataType interface {
	Type() int
	Serialize(encoder Encoder) error
	Deserialization(decoder io.Reader) (DataType, error)
}

const (
	INT = iota
	STRING
	BULK_STRING
	ARRAYS
	ERROR
)

var CRLF = []byte("\r\n")

type Encoder struct {
	w io.Writer

	buf [10]byte
}

func (e *Encoder) WriteByte(b byte) error {
	e.buf[0] = b
	if n, err := e.w.Write(e.buf[:1]); err != nil {
		return err
	} else if n != 1 {
		return fmt.Errorf("write failure")
	}
	return nil
}

func (e *Encoder) WriteUInt(i uint) error {
	s := strconv.FormatUint(uint64(i), 10)
	if n, err := e.w.Write([]byte(s)); err != nil {
		return err
	} else if n != len(s) {
		return fmt.Errorf("part write")
	}
	return nil
}

func (e *Encoder) WriteString(s string) error {
	if n, err := e.w.Write([]byte(s)); err != nil {
		return err
	} else if n != len(s) {
		return fmt.Errorf("part write")
	}
	return nil
}

type Decoder struct {
	r bufio.Reader
}

func (d *Decoder) ReadByte() (byte, error) {
	return d.r.ReadByte()
}

func (d *Decoder) ReadUint() (byte, error) {

}
