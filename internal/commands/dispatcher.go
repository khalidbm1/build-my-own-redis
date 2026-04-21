// internal/commands/dispatcher.go
//
// Dispatcher routes commands to handlers.
// Dispatcher يوجه الأوامر للمعالجات.

package commands

import (
	"errors"
	"strconv"
	"strings"
	"time"

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
	"SET":       handleSet,
	"GET":       handleGet,
	"DEL":       handleDel,
	"INCR":      handleIncr,
	"DECR":      handleDecr,
	"EXISTS":    handleExists,
	"KEYS":      handleKeys,
	"PING":      handlePing,
	"ECHO":      handleEcho,
	"LPUSH":     handleLPush,
	"RPUSH":     handleRPush,
	"LPOP":      handleLPop,
	"RPOP":      handleRPop,
	"LLEN":      handleLLen,
	"LRANGE":    handleLRange,
	"SADD":      handleSAdd,
	"SREM":      handleSRem,
	"SMEMBERS":  handleSMembers,
	"SISMEMBER": handleSIsMember,
	"SCARD":     handleSCard,
	"EXPIRE":    handleExpire,
	"PEXPIRE":   handlePExpire,
	"TTL":       handleTTL,
	"PERSIST":   handlePersist,
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

	var ttl *time.Duration
	for i := 2; i < len(args); i++ {
		if args[i].Type != protocol.BulkString {
			return protocol.NewError("ERR syntax error")
		}
		opt := strings.ToUpper(args[i].String)
		switch opt {
		case "EX", "PX":
			if i+1 >= len(args) || args[i+1].Type != protocol.BulkString {
				return protocol.NewError("ERR syntax error")
			}
			n, err := strconv.ParseInt(args[i+1].String, 10, 64)
			if err != nil || n <= 0 {
				return protocol.NewError("ERR value is not an integer or out of range")
			}
			unit := time.Second
			if opt == "PX" {
				unit = time.Millisecond
			}
			d := time.Duration(n) * unit
			ttl = &d
			i++
		default:
			return protocol.NewError("ERR syntax error")
		}
	}

	store.Set(key, value, ttl)
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
	value, exists, err := store.Get(key)

	if err != nil {
		if errors.Is(err, storage.ErrWrongType) {
			return protocol.NewError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return protocol.NewError("ERR " + err.Error())
	}

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
	value, exists, err := store.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrWrongType) {
			return protocol.NewError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return protocol.NewError("ERR " + err.Error())
	}

	var num int
	if !exists {
		num = 0
	} else {
		n, convErr := strconv.Atoi(value)
		if convErr != nil {
			return protocol.NewError("ERR value is not an integer or out of range")
		}
		num = n
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
	value, exists, err := store.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrWrongType) {
			return protocol.NewError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return protocol.NewError("ERR " + err.Error())
	}

	var num int
	if !exists {
		num = 0
	} else {
		n, convErr := strconv.Atoi(value)
		if convErr != nil {
			return protocol.NewError("ERR value is not an integer or out of range")
		}
		num = n
	}

	num--
	store.Set(key, strconv.Itoa(num), nil)
	return protocol.NewInteger(num)
}

func handleLPush(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) < 2 {
		return protocol.NewError("ERR wrong number of arguments for 'lpush' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	values := make([]string, len(args)-1)

	for i, arg := range args[1:] {
		if arg.Type != protocol.BulkString {
			return protocol.NewError("ERR invalid value type")
		}
		values[i] = arg.String
	}

	length := store.LPush(key, values...)

	if length == -1 {
		return protocol.NewError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return protocol.NewInteger(length)
}

func handleRPush(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) < 2 {
		return protocol.NewError("ERR wrong number of arguments for 'rpush' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	values := make([]string, len(args)-1)
	for i, arg := range args[1:] {
		if arg.Type != protocol.BulkString {
			return protocol.NewError("ERR invalid value type")
		}
		values[i] = arg.String
	}

	length := store.RPush(key, values...)
	if length == -1 {
		return protocol.NewError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return protocol.NewInteger(length)
}

func handleLPop(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'lpop' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	value, exists := store.LPop(key)

	if !exists {
		return protocol.NewNull()
	}

	return protocol.NewBulkString(value)
}

func handleRPop(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'rpop' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	value, exists := store.RPop(key)

	if !exists {
		return protocol.NewNull()
	}

	return protocol.NewBulkString(value)
}

func handleLLen(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'llen' command")
	}

	key := args[0].String
	length := store.LLen(key)
	return protocol.NewInteger(length)
}

func handleLRange(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 3 {
		return protocol.NewError("ERR wrong number of arguments for 'lrange' command")
	}

	key := args[0].String
	start, err1 := strconv.Atoi(args[1].String)
	stop, err2 := strconv.Atoi(args[2].String)

	if err1 != nil || err2 != nil {
		return protocol.NewError("ERR invalid index type")
	}

	values := store.LRange(key, start, stop)
	response := make([]protocol.Value, len(values))

	for i, v := range values {
		response[i] = protocol.NewBulkString(v)
	}

	return protocol.NewArray(response)
}

func handleSAdd(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) < 2 {
		return protocol.NewError("ERR wrong number of arguments for 'sadd' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	members := make([]string, len(args)-1)

	for i, arg := range args[1:] {
		if arg.Type != protocol.BulkString {
			return protocol.NewError("ERR invalid member type")
		}
		members[i] = arg.String
	}

	added := store.SAdd(key, members...)

	if added == -1 {
		return protocol.NewError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return protocol.NewInteger(added)
}

func handleSRem(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) < 2 {
		return protocol.NewError("ERR wrong number of arguments for 'srem' command")
	}

	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}

	key := args[0].String
	members := make([]string, len(args)-1)

	for i, arg := range args[1:] {
		if arg.Type != protocol.BulkString {
			return protocol.NewError("ERR invalid member type")
		}
		members[i] = arg.String
	}
	removed := store.SRem(key, members...)
	return protocol.NewInteger(removed)
}

func handleSMembers(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'smembers' command")
	}

	key := args[0].String
	members := store.SMembers(key)

	response := make([]protocol.Value, len(members))
	for i, member := range members {
		response[i] = protocol.NewBulkString(member)
	}

	return protocol.NewArray(response)
}

func handleSIsMember(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 2 {
		return protocol.NewError("ERR wrong number of arguments for 'sismember' command")
	}

	key := args[0].String
	member := args[1].String

	if store.SIsMember(key, member) {
		return protocol.NewInteger(1)
	}

	return protocol.NewInteger(0)
}

func handleSCard(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'scard' command")
	}

	key := args[0].String
	return protocol.NewInteger(store.SCard(key))
}

func handleExpire(store *storage.Store, args []protocol.Value) protocol.Value {
	return setExpire(store, args, time.Second, "expire")
}

func handlePExpire(store *storage.Store, args []protocol.Value) protocol.Value {
	return setExpire(store, args, time.Millisecond, "pexpire")
}

func setExpire(store *storage.Store, args []protocol.Value, unit time.Duration, name string) protocol.Value {
	if len(args) != 2 {
		return protocol.NewError("ERR wrong number of arguments for '" + name + "' command")
	}
	if args[0].Type != protocol.BulkString || args[1].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid argument type")
	}

	n, err := strconv.ParseInt(args[1].String, 10, 64)
	if err != nil {
		return protocol.NewError("ERR value is not an integer or out of range")
	}

	ttl := time.Duration(n) * unit
	if store.Expire(args[0].String, &ttl) {
		return protocol.NewInteger(1)
	}
	return protocol.NewInteger(0)
}

func handleTTL(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'ttl' command")
	}
	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}
	return protocol.NewInteger(store.TTL(args[0].String))
}

func handlePersist(store *storage.Store, args []protocol.Value) protocol.Value {
	if len(args) != 1 {
		return protocol.NewError("ERR wrong number of arguments for 'persist' command")
	}
	if args[0].Type != protocol.BulkString {
		return protocol.NewError("ERR invalid key type")
	}
	if store.Expire(args[0].String, nil) {
		return protocol.NewInteger(1)
	}
	return protocol.NewInteger(0)
}
