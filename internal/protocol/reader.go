package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(r),
	}
}

func (r *Reader) Read() (Value, error) {
	typ, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch Type(typ) {
	case SimpleString:
		return r.readSimpleString()

	case ErrorType:
		return r.readError()

	case Integer:
		return r.readInteger()

	case BulkString:
		return r.readBulkString()

	case Array:
		return r.readArray()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %c", typ)
	}
}
func (r *Reader) readLine() ([]byte, error) {
	line, err := r.reader.ReadBytes('\n')

	if err != nil {
		return nil, err
	}

	if len(line) >= 2 && line[len(line)-2] == '\r' {
		return line[:len(line)-2], nil
	}

	return line[:len(line)-1], nil
}

func (r *Reader) readSimpleString() (Value, error) {
	line, err := r.readLine()

	if err != nil {
		return Value{}, err
	}

	return NewSimpleString(string(line)), nil
}

func (r *Reader) readError() (Value, error) {
	line, err := r.readLine()

	if err != nil {
		return Value{}, err
	}
	return NewError(string(line)), nil
}

func (r *Reader) readInteger() (Value, error) {
	line, err := r.readLine()

	if err != nil {
		return Value{}, err
	}

	num, err := strconv.Atoi(string(line))

	if err != nil {
		return Value{}, fmt.Errorf("Invalid integer: %v", err)
	}

	return NewInteger(num), nil
}

func (r *Reader) readBulkString() (Value, error) {
	line, err := r.readLine()

	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(string(line))

	if err != nil {
		return Value{}, fmt.Errorf("Invalid bulk string length: %v", err)
	}

	if length == -1 {
		return NewNull(), nil
	}

	buf := make([]byte, length+2)
	_, err = io.ReadFull(r.reader, buf)

	if err != nil {
		return Value{}, err
	}

	return NewBulkString(string(buf[:length])), nil
}

func (r *Reader) readArray() (Value, error) {
	line, err := r.readLine()

	if err != nil {
		return Value{}, err
	}

	count, err := strconv.Atoi(string(line))

	if err != nil {
		return Value{}, fmt.Errorf("Invalid array count: %v", err)
	}

	if count == -1 {
		return NewNullArray(), nil
	}

	array := make([]Value, count)
	for i := 0; i < count; i++ {
		val, err := r.Read()

		if err != nil {
			return Value{}, err
		}
		array[i] = val
	}
	return NewArray(array), nil
}
