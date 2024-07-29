package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
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
