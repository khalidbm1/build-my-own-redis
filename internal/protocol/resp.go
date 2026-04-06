package protocol

type Type byte

const (
	SimpleString Type = '+'
	ErrorType    Type = '-'
	Integer      Type = ':'
	BulkString   Type = '$'
	Array        Type = '*'
)

type Value struct {
	Type   Type
	String string
	Number int
	Array  []Value
	IsNull bool
}

func NewSimpleString(s string) Value {
	return Value{Type: SimpleString, String: s}
}

func NewError(s string) Value {
	return Value{Type: ErrorType, String: s}
}

func NewInteger(n int) Value {
	return Value{Type: Integer, Number: n}
}

func NewBulkString(s string) Value {
	return Value{Type: BulkString, String: s}
}

func NewArray(values []Value) Value {
	return Value{Type: Array, Array: values}
}

func NewNull() Value {
	return Value{Type: BulkString, IsNull: true}
}
