/*
If your Redis server crashes,
all data in memory is lost.
Real Redis uses two persistence mechanisms:
AOF (Append-Only File) records every write command, and RDB creates snapshots.
In this stage, you'll implement AOF — every SET, LPUSH, etc.
gets written to a file so you can replay it on restart.

لو خادم Redis يتعطل، كل البيانات في الذاكرة تروح. Redis الحقيقي يستخدم آليتين:
AOF (ملف Append-Only) يسجل كل أمر كتابة، و RDB ينشئ لقطات. في هالمرحلة، بتطبق AOF —
كل SET, LPUSH, إلخ تُكتب لملف عشان تقدر تكررها عند الإعادة.
*/

package persistence

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/khalidbm1/build-my-own-redis/internal/commands"
	"github.com/khalidbm1/build-my-own-redis/internal/protocol"
	"github.com/khalidbm1/build-my-own-redis/internal/storage"
)

type AOF struct {
	file     *os.File
	writer   *bufio.Writer
	filename string
	mu       sync.Mutex
}

func NewAOF(filename string) *AOF {
	return &AOF{
		filename: filename,
	}
}

func (a *AOF) Open() error {
	file, err := os.OpenFile(a.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	a.file = file
	a.writer = bufio.NewWriter(file)
	return nil
}

// IsWriteCommand reports whether a RESP command mutates state and must be persisted.
func IsWriteCommand(value protocol.Value) bool {
	if value.Type != protocol.Array || len(value.Array) == 0 {
		return false
	}

	if value.Array[0].Type != protocol.BulkString {
		return false
	}

	cmd := value.Array[0].String
	writeCommands := map[string]bool{
		"SET":     true,
		"DEL":     true,
		"INCR":    true,
		"DECR":    true,
		"LPUSH":   true,
		"RPUSH":   true,
		"LPOP":    true,
		"RPOP":    true,
		"SADD":    true,
		"SREM":    true,
		"EXPIRE":  true,
		"PEXPIRE": true,
	}

	return writeCommands[cmd]
}

func (a *AOF) Append(value protocol.Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file == nil {
		if err := a.Open(); err != nil {
			return err
		}
	}

	if !IsWriteCommand(value) {
		return nil
	}

	respWriter := protocol.NewWriter(a.writer)

	if err := respWriter.Write(value); err != nil {
		return err
	}
	return a.writer.Flush()
}

func (a *AOF) Load(store *storage.Store) error {
	file, err := os.Open(a.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	reader := protocol.NewReader(file)
	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		_ = commands.Dispatch(store, value)
	}
	return nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.writer != nil {
		if err := a.writer.Flush(); err != nil {
			return err
		}
	}

	if a.file != nil {
		return a.file.Close()
	}
	return nil
}



