// internal/commands/dispatcher.go
//
// Dispatcher routes commands to handlers.
// Dispatcher يوجه الأوامر للمعالجات.

package commands

import (
	"strconv"
	"strings"

	"github.com/khalidbm1/build-my-own-redis/internal/protocol"
	"github.com/khalidbm1/build-my-own-redis/internal/storage"
)

// Handler is a function that handles a Redis command.
// It receives the store and the command arguments,
// and returns a RESP response.
//
// Handler هي دالة تتعامل مع أمر Redis.
// تستقبل المخزن وأرجومنتات الأمر،
// وترجع استجابة RESP.
type Handler func(store *storage.Store, args []protocol.Value) protocol.Value

// handlers maps command names to handler functions.
// هالخريطة تربط أسماء الأوامر مع دوال المعالجات.
var handlers = map[string]Handler{
	"SET":    handleSet,
	"GET":    handleGet,
	"DEL":    handleDel,
	"INCR":   handleIncr,
	"DECR":   handleDecr,
	"EXISTS": handleExists,
	"KEYS":   handleKeys,
	"PING":   handlePing,
	"ECHO":   handleEcho,
	//"EXPIRE":    handleExpire,
	//"SCAN":      handleScan,
	// "SCANCOUNT": handleScanCount,
	// "TYPE":      handleType,
	// "FLUSHDB":   handleFlushDB,
	// "FLUSHALL":  handleFlushAll,
	// "SAVE":      handleSave,
	// "BGSAVE":    handleBGSave,
	// "INFO":      handleInfo,
	// "LASTSAVE":  handleLastSave,
	//"QUIT":      handleQuit,
}

// Delete removes key-value pairs from the store.
// Delete يحذف مفتاح-قيمة معينة من المخزن.

// Dispatch routes a command to its handler.
// If the handler exists, it executes and returns the response.
// If not, it returns an error.
//
// Dispatch يوجه أمر لـ معالجه.
// لو المعالج موجود، يشتغل ويرجع الاستجابة.
// لو لا، يرجع خطأ.
func Dispatch(store *storage.Store, value protocol.Value) protocol.Value {
	// Commands must be arrays with at least one element
	// الأوامر لازم تكون قوائم مع عنصر واحد على الأقل
	if value.Type != protocol.Array || len(value.Array) == 0 {
		return protocol.NewError("ERR invalid command format")
	}

	// The first element is the command name
	// العنصر الأول هو اسم الأمر
	if value.Array[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid command")
	}

	// Convert command name to uppercase (Redis is case-insensitive)
	// حول اسم الأمر لأحرف كبيرة (Redis غير حساسة للحالة)
	cmdName := strings.ToUpper(value.Array[0].String)

	// Look up the handler
	// ابحث عن المعالج
	handler, exists := handlers[cmdName]
	if !exists {
		return protocol.NewError("ERR unknown command '" + cmdName + "'")
	}

	// Get the arguments (everything after the command name)
	// احصل على الأرجومنتات (كل شي بعد اسم الأمر)
	args := value.Array[1:]

	// Call the handler and return its response
	// استدعِ المعالج وارجع استجابته
	return handler(store, args)
}

// handlePing responds with PONG.
// PING is the simplest Redis command — used to test connectivity.
//
// handlePing يرد مع PONG.
// PING هو أبسط أمر Redis — يُستخدم لاختبار الاتصال.
func handlePing(store *storage.Store, args []protocol.Value) protocol.Value {
	// PING can have an optional argument to echo back
	// PING يقدر يكون عندها أرجومنت اختياري لإرجاع الصدى
	if len(args) == 0 {
		return protocol.NewSimpleString("PONG")
	}

	// If an argument was provided, echo it back
	// لو أرجومنت اتعطى، ارجع الصدى
	if args[0].Type == protocol.BulkString {
		return protocol.NewBulkString(args[0].String)
	}

	return protocol.NewSimpleString("PONG")
}

// handleEcho echoes back the provided argument.
// ECHO "hello" → "hello"
//
// handleEcho يرجع الصدى للأرجومنت المعطى.
func handleEcho(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'echo' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid argument type")
	}

	return protocol.NewBulkString(args[0].String)
}

// handleSet stores a value in memory.
// SET key value → OK
//
// handleSet يخزن قيمة في الذاكرة.
func handleSet(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) < 2 {
		return protocol.NewError("ERR wrong number of arguments for 'set' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}
	if args[1].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid value type")
	}

	key := args[0].String
	value := args[1].String

	// TODO: In Stage 5, we'll add support for options like EX, PX, NX, XX
	// الحين، نخزن بدون خيارات. في المرحلة 5 بنضيف EX, PX, NX, XX

	store.Set(key, value, nil)
	return protocol.NewSimpleString("OK")
}

// handleGet retrieves a value from memory.
// GET key → value (or nil if not found)
//
// handleGet يسترجع قيمة من الذاكرة.
func handleGet(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'get' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	value, exists := store.Get(key)

	if !exists {
		return protocol.NewNull()
	}

	return protocol.NewBulkString(value)
}

// handleDel deletes keys from memory.
// DEL key1 key2 key3 → number of keys deleted
//
// handleDel يحذف مفاتيح من الذاكرة.
func handleDel(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) == 0 {
		return protocol.NewError("ERR wrong number of arguments for 'del' command")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		if arg.Type != protocol.BulkString {
			return protocol.NewError("ERR invalid key type")
		}
		keys[i] = arg.String
	}

	count := store.Delete(keys)
	return protocol.NewInteger(count)
}

// handleExists checks if keys exist.
// EXISTS key1 key2 key3 → number of keys that exist
//
// handleExists يشيك إذا المفاتيح موجودة.
func handleExists(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) == 0 {
		return protocol.NewError("ERR wrong number of arguments for 'exists' command")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		if arg.Type != protocol.BulkString {
			return protocol.NewError("ERR invalid key type")
		}
		keys[i] = arg.String
	}

	count := store.Exists(keys)
	return protocol.NewInteger(count)
}

// handleKeys returns keys matching a pattern.
// KEYS * → all keys
//
// handleKeys يرجع مفاتيح تطابق pattern.
func handleKeys(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'keys' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid pattern type")
	}

	pattern := args[0].String
	keys := store.Keys(pattern)

	// Convert keys to RESP array of bulk strings
	// حول المفاتيح لقائمة RESP من نصوص بكتلة
	values := make([]protocol.Value, len(keys))
	for i, key := range keys {
		values[i] = protocol.NewBulkString(key)
	}

	return protocol.NewArray(values)
}

// handleIncr increments an integer value.
// INCR key → new value (or error if not an integer)
//
// handleIncr يزيد قيمة عدد صحيح.
// Better implementation for handleIncr and handleDecr
func handleIncr(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'incr' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	value, exists := store.Get(key)

	var num int
	if !exists {
		num = 0
	} else {
		var err error
		num, err = strconv.Atoi(value)
		if err != nil {
			return protocol.NewError("ERR value is not an integer or out of range")
		}
	}

	num++
	store.Set(key, strconv.Itoa(num), nil)
	return protocol.NewInteger(num)
}

func handleDecr(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'decr' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	value, exists := store.Get(key)

	var num int
	if !exists {
		num = 0
	} else {
		var err error
		num, err = strconv.Atoi(value)
		if err != nil {
			return protocol.NewError("ERR value is not an integer or out of range")
		}
	}

	num--
	store.Set(key, strconv.Itoa(num), nil)
	return protocol.NewInteger(num)
}
