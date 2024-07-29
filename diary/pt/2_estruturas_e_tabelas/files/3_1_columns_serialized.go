package table_structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Column represents a column in a table
const (
	COL_TYPE_INT = iota
	COL_TYPE_SMALL_INT
	COL_TYPE_BIG_INT
	COL_TYPE_STRING
	COL_TYPE_FLOAT
	COL_TYPE_DOUBLE
	COL_TYPE_BOOL
	COL_TYPE_TIMESTAMP
	COL_TYPE_BLOB
)

var typesMap = map[int]map[string]interface{}{
	COL_TYPE_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeInt,
	},
	COL_TYPE_SMALL_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeSmallInt,
	},
	COL_TYPE_BIG_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeBigInt,
	},
	COL_TYPE_STRING: {
		"serialize":   serializeString,
		"deserialize": deserializeString,
	},
	COL_TYPE_FLOAT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeFloat,
	},
	COL_TYPE_DOUBLE: {
		"serialize":   genericSerialize,
		"deserialize": deserializeDouble,
	},
	COL_TYPE_BOOL: {
		"serialize":   genericSerialize,
		"deserialize": deserializeBool,
	},
	COL_TYPE_TIMESTAMP: {
		"serialize":   serializeTimestamp,
		"deserialize": deserializeTimestamp,
	},
	COL_TYPE_BLOB: {
		"serialize":   serializeBlob,
		"deserialize": deserializeBlob,
	},
}

func genericSerialize(value interface{}) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func genericDeserialize(value []byte, i any) {
	buf := bytes.NewReader(value)
	err := binary.Read(buf, binary.LittleEndian, i)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
}

func serializeString(value interface{}) []byte {
	return []byte(value.(string))
}

func serializeTimestamp(value interface{}) []byte {
	return []byte(value.(string))
}

func serializeBlob(value interface{}) []byte {
	return value.([]byte)
}

func deserializeInt(value []byte) interface{} {
	var i int32
	genericDeserialize(value, &i)
	return i
}

func deserializeString(value []byte) interface{} {
	return string(value)
}

func deserializeFloat(value []byte) interface{} {
	var i float32
	genericDeserialize(value, &i)
	return i
}

func deserializeDouble(value []byte) interface{} {
	var i float64
	genericDeserialize(value, &i)
	return i
}

func deserializeBool(value []byte) interface{} {
	var i bool
	genericDeserialize(value, &i)
	return i
}

func deserializeTimestamp(value []byte) interface{} {
	return string(value)
}

func deserializeBlob(value []byte) interface{} {
	return value
}

func deserializeBigInt(value []byte) interface{} {
	var i int64
	genericDeserialize(value, &i)
	return i
}

func deserializeSmallInt(value []byte) interface{} {
	var i int16
	genericDeserialize(value, &i)
	return i
}

type SerializedColumnValue struct {
	Type  int
	Value []byte
}

func getFuncForType(t int, fn string) (interface{}, bool) {
	f, ok := typesMap[t][fn]
	return f, ok
}

func (s *SerializedColumnValue) DeserializeValue() interface{} {

	fn, ok := getFuncForType(s.Type, "deserialize")
	if !ok {
		fmt.Println("Falhou pra pegar a função deserialize")
		return nil
	}

	deserializeFn, ok := fn.(func([]byte) interface{})
	if !ok {
		fmt.Println("Falhou cast deserialize")
		return nil
	}

	return deserializeFn(s.Value)
}

func SerializeValue(value interface{}, valType int) *SerializedColumnValue {
	fn, ok := getFuncForType(valType, "serialize")

	if !ok {
		fmt.Println("Falhou pra pegar a função")
		return nil
	}

	serializeFn, ok := fn.(func(interface{}) []byte)
	// print serializeFn type
	if !ok {
		fmt.Println("Falhou cast da função")
		return nil
	}

	return &SerializedColumnValue{
		Type:  valType,
		Value: serializeFn(value),
	}
}

func TestSerialize() {
	fmt.Println("Int32")
	val := SerializeValue(int16(2), COL_TYPE_SMALL_INT)
	fmt.Println(val.DeserializeValue())
	fmt.Println("String")
	val = SerializeValue("asdasdasdasd", COL_TYPE_STRING)
	fmt.Println(val.DeserializeValue())

}
