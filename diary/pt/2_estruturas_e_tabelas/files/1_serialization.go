package table_structures

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type TableTest struct {
	Id     int
	Name   string
	City   interface{}
	Number int
}

func Serialization() {
	// Create an instance of the Person struct
	rows := make([]TableTest, 2)
	rows[0] = TableTest{
		Id:     1,
		Name:   "Nicolas",
		City:   float32(0.8),
		Number: 123456,
	}
	rows[1] = TableTest{
		Id:     2,
		Name:   "John",
		City:   "London",
		Number: 789456,
	}

	// Create a new buffer to write the serialized data to
	var b bytes.Buffer

	// Create a new gob encoder and use it to encode the person struct
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(rows); err != nil {
		fmt.Println("Error encoding struct:", err)
		return
	}

	// The serialized data can now be found in the buffer
	serializedData := b.Bytes()
	fmt.Printf("Serialized data: %s and its length = %d\n", serializedData, len(b.Bytes()))

	// Create a new buffer from the serialized data
	c := bytes.NewBuffer(serializedData)

	// Create a new gob decoder and use it to decode the person struct
	var deserialized []TableTest
	dec := gob.NewDecoder(c)
	if err := dec.Decode(&deserialized); err != nil {
		fmt.Println("Error decoding struct:", err)
		return
	}

	// The person struct has now been deserialized
	fmt.Println("Deserialized struct:", deserialized)
}
