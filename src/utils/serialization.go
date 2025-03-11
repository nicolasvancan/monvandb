package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strconv"
)

func Deserialize(value []byte, dst interface{}) error {

	// Create a new buffer from the serialized data
	c := bytes.NewBuffer(value)

	// Create a new gob decoder and use it to decode the person struct
	dec := gob.NewDecoder(c)
	if err := dec.Decode(dst); err != nil {
		fmt.Println("Error decoding struct:", err)
		return err
	}

	return nil
}

func Serialize(value interface{}) ([]byte, error) {
	var b bytes.Buffer

	// Create a new gob encoder and use it to encode the person struct
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(value); err != nil {
		fmt.Println("Error encoding struct:", err)
		return nil, err
	}

	// The serialized data can now be found in the buffer
	return b.Bytes(), nil
}

// Function that converts struct to json
func ToJson(value interface{}) ([]byte, error) {

	// Create a new gob decoder and use it to decode the person struct
	enc, err := json.Marshal(value)
	if err != nil {
		fmt.Println("Error encoding struct:", err)
		return nil, err
	}

	return enc, nil
}

// Function that converts json to struct
func FromJson(value []byte, dst interface{}) error {
	// Create a new buffer from the serialized data
	err := json.Unmarshal(value, dst)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Casting

func CastToInt(key interface{}) int {
	switch key.(type) {
	case int:
		return key.(int)
	case int64:
		return int(key.(int64))
	case float64:
		return int(key.(float64))
	case string:
		val, _ := strconv.Atoi(key.(string))
		return val
	default:
		return 0
	}
}

func CastToString(key interface{}) string {
	switch key.(type) {
	case int:
		return strconv.Itoa(key.(int))
	case int64:
		return strconv.Itoa(int(key.(int64)))
	case float64:
		return strconv.Itoa(int(key.(float64)))
	case string:
		return key.(string)
	default:
		return ""
	}
}
func CastToFloat(key interface{}) float32 {
	switch key.(type) {
	case int:
		return float32(key.(int))
	case int64:
		return float32(key.(int64))
	case float64:
		return float32(key.(float64))
	case string:
		val, _ := strconv.ParseFloat(key.(string), 64)
		return float32(val)
	default:
		return 0
	}
}
func CastToBool(key interface{}) bool {
	switch key.(type) {
	case bool:
		return key.(bool)
	case string:
		val, _ := strconv.ParseBool(key.(string))
		return val
	default:
		return false
	}
}

func CastToTimestamp(key interface{}) int64 {
	return int64(CastToBigInt(key))
}

func CastToDouble(key interface{}) float64 {
	switch key.(type) {
	case int:
		return float64(key.(int))
	case int64:
		return float64(key.(int64))
	case float64:
		return key.(float64)
	case string:
		val, _ := strconv.ParseFloat(key.(string), 64)
		return val
	default:
		return 0
	}
}
func CastToBigInt(key interface{}) int64 {
	switch key.(type) {
	case int:
		return int64(key.(int))
	case int64:
		return key.(int64)
	case float64:
		return int64(key.(float64))
	case string:
		val, _ := strconv.ParseInt(key.(string), 10, 64)
		return val
	default:
		return 0
	}
}
func CastToSmallInt(key interface{}) int16 {
	switch key.(type) {
	case int:
		return int16(key.(int))
	case int64:
		return int16(key.(int64))
	case float64:
		return int16(key.(float64))
	case string:
		val, _ := strconv.ParseInt(key.(string), 10, 16)
		return int16(val)
	default:
		return 0
	}
}
