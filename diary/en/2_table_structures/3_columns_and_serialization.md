# Columns Serialization

One eventual problem is how to serialize table rows. Let's say we have a table with four different columns, each of them contains data of one specific type, how could it be transformed into bytes?

My last experiment with serialization was successful and I converted a generic struct type to bytes and the deserialization process as well. I don't know if it works with a generic type. In C, to declare a generic type we use a pointer to void, in java is the Object itself, in JavaScript is an Empty Json {}, which is the Object of JavaScript, and in Golang the generic type is represented by the **interface{}**.

I've done some tests with the generic type, whose useful alias in Golang is the type **any**. I can use both types to represent that a variable can have any type.

Many doubts emerge when dealing with this kind of serialization. Is there a bytes overhead when using normal struct serialization? Is it fast enought?

I wanted to test some concepts using Golang too, so I decided that implementing my personal serialization for some variable types would be good.

Instead of using the field **Value** as a generic type, I've choosen it to be a byte array, and the serialization would be done by me. For every new variable type, a different approach would be applied.

```go
type SerializedColumnValue struct {
	Type  int
	Value []byte
}
```

For now, I wanted to prove this concept with the following variable types:

```go
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
```

The field **Type** of the SerializedColumnValue struct represents one of the types above. I know that there are many cases, and I'd probably use a switch case statement, but since Golang provides us the map structure, why not use it to implement this logic of choosing what type we would process?

This is something that I always do in JavaScript or even in Python when managing a large switch case statement. I really don't know if it is faster (probably not) but the code remains not so poluted with a lot of conditional statements. What I wanted to do: For every Value type, I had to implement both serialization function and deserialization function, and here is why

```go
var typesMap = map[int]map[string]interface{}{
	COL_TYPE_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeInt,
	}
}
```

Note that I created a map with keys of integer variable, which is the case of our column types, and the value can be anything, even a function. So if I want to increase the possibilities of this map, I'd just insert another case inside my **typesMap**, as shown below:

```go
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
	}
}
```

The field serialize contains the respective function for serialization, obviously, the deserialize, contains the deserialization function reference. Knowing that, it is time to implement both function to serialize and deserialize.

```go
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
```

What is happening for that to work properly? When I get a map value, I must input the type of the variable I want to serialize/deserialize and also a string indicating what function I want to use (This case has only two availables). Doing so, I recover the respective function I want to use.

Since the functions **getFuncForType** returns a generic type, I must cast it to the respective interface function type, which can differ depending on the function.

For the serialization functions, the inputs are diverse and the output is a byte array. For deserialization the opposite happens. All the functions that will be inserted in this interface must be in compliance with those types, as show in the example below:

```go
// Follows the interface correctly
func serializeString(value interface{}) []byte {
	return []byte(value.(string))
}

func deserializeString(value []byte) interface{} {
	return string(value)
}
``` 

## The final form

I had great time understanding some Golang concepts through the first experiment of serialization and deserialization. But all of that could be avoided if I used the **interface{}** instead of the byte array in the Value field of the table. Nontheless, I still validated what I wanted, and learned how to work properly with generic types, casting also serialization.

**How it looks like now with **any** type? **

The projected schema has changed. I don't need to track the type of the value for the generic type in Golang. The *gob* library addresses this problem really well and can convert even **nil** values when needed. It became much more simple. I created another package, **table** that holds all files related to tables and its definitions, alongside some utilities, that will be moved to another place in the future, in which I added the file **serialization.go**, having two functions: Serialize and Deserialize, working for all types of structs, as show below:

```go
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

```

Obviously, I have also created a good test case for these functions above, in which I imitate a row insertion with different types. The struct is called **ColumnValue**, having the **Value** field and also the Col, that is refered to what col position it is related in the table configurations (That will be addressed after this part).

```go
func TestSerializationWithDataFile(t *testing.T) {
	// Create tmp dir path
	tmpDirPath := t.TempDir()

	// Get final temporary path for DataFile file
	dbFilePath := tmpDirPath + string(os.PathSeparator) + "test.db"

	// Create a mock row with 4 columns, each with a different type
	row := make(table.RowValues, 5)

	row[0] = table.ColumnValue{Col: 1, Value: uint8(10)}
	row[1] = table.ColumnValue{Col: 2, Value: "Hello"}
	row[2] = table.ColumnValue{Col: 3, Value: 0.8}
	row[3] = table.ColumnValue{Col: 4, Value: 123456}
	row[4] = table.ColumnValue{Col: 4, Value: nil}

	// Serialize them
	serializedRow, err := table.Serialize(row)

	if err != nil {
		t.Errorf("Error encoding struct: %s", err)
	}

	// Open the DataFile (Create new if does not exist)
	dFile, err := files.OpenDataFile(dbFilePath)

	if err != nil {
		t.Errorf("Error opening data file: %s", err)
	}

	// Insert one key
	dFile.Insert([]byte("1"), serializedRow)
	dFile.ForceSync()

	dFile.Close()

	// Reopen recent closed DataFile
	dFile, err = files.OpenDataFile(dbFilePath)

	if err != nil {
		t.Errorf("Error opening data file: %s", err)
	}

	// Create new variable to hold value returned from DataFile
	serializedRow2 := dFile.Get([]byte("1"))

	if serializedRow2 == nil {
		t.Errorf("Error getting data from data file")
	}

	var row2 table.RowValues
	// Deserialize it
	err = table.Deserialize(serializedRow2[0].Value, &row2)

	if err != nil {
		t.Errorf("Error decoding struct: %s", err)
	}

	if len(row) != len(row2) {
		t.Errorf(error_v_string, row, row2)
	}

	for i := 0; i < len(row); i++ {
		if row[i].Col != row2[i].Col || row[i].Value != row2[i].Value {
			t.Errorf(error_v_string, row, row2)
		}
	}
}
```

Look how cool is this. I create a bTree for storing the row. I serialize the row and insert the key value into the binary tree for key = '1'. After that, I close the binary Tree file and reopen it, read the stored value for key = '1', deserialize it and check whether or not I could store and retrieve the value maintaining its integrity. It worked.

# Conclusion

The experiment went well, I could effectivelly serialize and deserialize all those values types that were listed in iota enumerate, eventhough I still use the **gob** serializing functions for struct type, I could validate some concepts that might be used for me in the next chapters. 

I think that It will worth all the hard work of serializing and deserializing using this kind of schema, only if the time consumption reduces extremly, which is still not the case. I am going to persist the use of serializing structs with my data and only refactor this for gaining performance in the future.