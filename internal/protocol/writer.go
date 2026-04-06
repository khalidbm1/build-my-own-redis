package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Writer struct {
	writer *bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: bufio.NewWriter(w),
	}
}

func (w *Writer) Write(value Value) error {
	switch value.Type {
	case SimpleString:
		return w.writeSimpleString([]byte(value.String))

	case ErrorType:
		return w.writeError([]byte(value.String))

	case Integer:
		return w.writeInteger([]byte(strconv.Itoa(value.Number)))

	case BulkString:
		return w.writeBulkString([]byte(value.String))

	case Array:
		return w.writeArray(value.Array)
	default:
		return fmt.Errorf("unknown type: %c", value.Type)
	}
}

// writeSimpleString writes a simple string to the writer.
// A simple string is a string that is not a bulk string.
// It is prefixed with a "+" character.
func (w *Writer) writeSimpleString(line []byte) error {
	// Write the prefix and the string itself.
	_, err := w.writer.WriteString("+" + string(line) + "\r\n")
	if err != nil {
		return err
	}
	// Make sure the buffer is flushed.
	return w.writer.Flush()
}

func (w *Writer) writeError(line []byte) error {
	_, err := w.writer.WriteString("-" + string(line) + "\r\n")
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

// writeInteger writes an integer to the writer.
// An integer is prefixed with a ":" character.
func (w *Writer) writeInteger(line []byte) error {
	// Write the prefix and the integer itself.
	_, err := w.writer.WriteString(":" + string(line) + "\r\n")
	if err != nil {
		return err
	}
	// Make sure the buffer is flushed.
	return w.writer.Flush()
}

// Integer is a type that represents an integer.
// It is prefixed with a ":" character.

// writeBulkString writes a bulk string to the writer.
// A bulk string is a string that is prefixed with a "$" character.
// The length of the string is written as a decimal number
// followed by a newline and then the string itself.
func (w *Writer) writeBulkString(line []byte) error {
	// Write the prefix and the length of the string.
	_, err := w.writer.WriteString("$" + strconv.Itoa(len(line)) + "\r\n")

	if err != nil {
		return err
	}

	// Write the string itself.
	_, err = w.writer.WriteString(string(line) + "\r\n")
	if err != nil {
		return err
	}

	// Make sure the buffer is flushed.
	return w.writer.Flush()
}

// writeArray writes an array to the writer.
// An array is prefixed with a "*" character followed by the length of the array.
func (w *Writer) writeArray(values []Value) error {
	// Write the prefix and the length of the array.
	_, err := w.writer.WriteString("*" + strconv.Itoa(len(values)) + "\r\n")
	if err != nil {
		return err
	}

	// Write each value in the array.
	for _, val := range values {
		if err := w.Write(val); err != nil {
			return err
		}
	}

	return nil
}
