# Columns Serialization

One eventual problem is how to serialize column data, in other words, how to convert a variable struct into bytes? Let's say we have a table with four different columns, each of them contains data of one specific type, how could it be transformed into bytes?

My last experiment with serialization was successful and I converted a generic struct type to bytes and the deserialization process as well. I don't know if it works with a generic type. In C, to declare a generic type we use a pointer to void, in java is the Object itself, in JavaScript is an Empty Json {}, which is the Object of JavaScript, and in Golang the generic type is represented by the **interface{}**.

I've done some experiments with the generic type, whose useful alias is the type **any**, I can use both types to represent that a variable can have any type.

My point here is that I don't know if I store a variable with any type, the value could be easly retrieved, I mean, if I serialize and save an **uint8** value in a variable of *any* type and convert this variable into a byte array, when I convert it back to a variable, will it be a **uint8**?

To avoid any kind of problem regarding serialized variables, I decided that I would serialize the **Value** variable, In other words, I would have a struct type that would contain the respective type of a value and it's value as a byte array, as presented below:

```go
type SerializedColumnValue struct {
	Type  int
	Value []byte
}
``` I 