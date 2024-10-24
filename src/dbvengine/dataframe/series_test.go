package dataframe

import (
	"fmt"
	"testing"
	"time"
)

// Test data containing all types of series
// First is int, second is float, third is bool, fourth is string, fifth is timestamp

var testSeries = []Series{
	NewSeries([]Element{Elem(1), Elem(2), Elem(3)}).Cast(INT),
	NewSeries([]Element{Elem(1.5), Elem(2.5), Elem(3.5)}).Cast(FLOAT),
	NewSeries([]Element{Elem(true), Elem(false), Elem(true)}).Cast(BOOL),
	NewSeries([]Element{Elem("test1"), Elem("test2"), Elem("test3")}).Cast(STRING),
	NewSeries([]Element{Elem(time.Now().Add(10000)), Elem(time.Now().Add(20000)), Elem(time.Now().Add(30000))}).Cast(TIMESTAMP),
}

func TestElementCreation(t *testing.T) {
	// Test element creation
	fe := Elem(1.5)
	if fe.GetValue() != 1.5 {
		t.Errorf("Expected 1.5, got %f", fe.GetValue())
	}

	ie := Elem(1)
	if ie.GetValue() != 1 {
		t.Errorf("Expected 1, got %d", ie.GetValue())
	}

	be := Elem(true)
	if be.GetValue() != true {
		t.Errorf("Expected true, got %t", be.GetValue())
	}

	se := Elem("test")
	if se.GetValue() != "test" {
		t.Errorf("Expected test, got %s", se.GetValue())
	}

	ne := Elem(nil)
	if ne == nil || !ne.IsNull() {
		t.Errorf("Expected nil, got %v", ne.GetValue())
	}

	nue := Elem("NULL")
	if !nue.IsNull() {
		t.Errorf("Expected nil, got %v", nue.GetValue())
	}

	now := time.Now()
	te := Elem(now)
	if te.GetValue() != now {
		t.Errorf("Expected time, got %v", te.GetValue())
	}

}

func TestCasts(t *testing.T) {
	// Test casts
	fe := Elem(1.5)
	if fe.ToFloat().GetValue() != 1.5 {
		t.Errorf("Expected 1.5, got %f", fe.ToFloat().GetValue())
	}
	fmt.Printf("Int value: %d\n", fe.ToInt())
	// for int
	if fe.ToInt().GetValue() != 1 {
		t.Errorf("Expected 1, got %d", fe.ToInt().GetValue())
	}

	// for bool
	if fe.ToBool().GetValue() != true {
		t.Errorf("Expected true, got %t", fe.ToBool().GetValue())
	}

	// for string
	if fe.ToString().GetValue() != "1.5" {
		t.Errorf("Expected 1.5, got %s", fe.ToString().GetValue())
	}

	// for time
	if fe.ToTimestamp().GetValue() == nil {
		t.Errorf("Expected nil, got %v", fe.ToTimestamp().String())
	}

	ie := Elem(1)
	if ie.ToInt().GetValue() != 1 {
		t.Errorf("Expected 1, got %d", ie.ToInt().GetValue())
	}

	be := Elem(true)
	if be.ToBool().GetValue() != true {
		t.Errorf("Expected true, got %t", be.ToBool().GetValue())
	}

	se := Elem("test")
	if se.ToString().GetValue() != "test" {
		t.Errorf("Expected test, got %s", se.ToString().GetValue())
	}

	ne := Elem(nil)
	if !ne.IsNull() {
		t.Errorf("Expected nil, got %v", ne.GetValue())
	}

	nue := Elem("NULL")
	if !nue.IsNull() {
		t.Errorf("Expected nil, got %v", nue.GetValue())
	}

}

// Here we test the comparsion methods
func TestElementComparisons(t *testing.T) {
	// Test element creation
	fe := Elem(1.5)
	ie := Elem(1)
	be := Elem(true)
	se := Elem("test")
	ne := Elem(nil)
	nue := Elem("NULL")
	now := time.Now()
	te := Elem(now)
	ieSlice := []Element{Elem(1), Elem(2), Elem(3)}
	//feSlice := []Element{Elem(1.5), Elem(2.5), Elem(3.5)}
	//beSlice := []Element{Elem(true), Elem(false), Elem(true)}
	//seSlice := []Element{Elem("test1"), Elem("test2"), Elem("test3")}
	//teSlice := []Element{Elem(now.Add(10000)), Elem(now.Add(20000)), Elem(now.Add(30000))}

	for _, serie := range testSeries {
		switch serie.GetType() {
		case INT:
			// INT ELEMENTS
			if !ie.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}
			// Test not equal
			if ie.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}
			// Test greater
			if ie.Greater(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if ie.Less(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !ie.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ie.LessEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ie.IsIn(ieSlice) {
				t.Errorf("Expected true, got false")
			}

			// FLOAT ELEMENT
			if fe.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value %f and %f", fe.GetValue(), serie.Elements[0].ToFloat().GetValue())
			}
			// Test not equal
			if !fe.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}
			// Test greater
			if !fe.Greater(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if fe.Less(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !fe.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if fe.LessEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if fe.IsIn(ieSlice) {
				t.Errorf("Expected true, got false")
			}

			// BOOL ELEMENT
			if !be.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value")
			}
			// Test not equal
			if be.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}
			// STRING ELEMENT
			if se.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value")
			}

			if !se.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			// TIMESTAMP ELEMENT
			if te.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value")
			}

			// Null element
			if !ne.IsNull() {
				t.Errorf("Expected true, got false")
			}

			if !nue.IsNull() {
				t.Errorf("Expected true, got false")
			}

			if ne.Equal(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !ne.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if ne.Greater(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if ne.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ne.Less(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ne.LessEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if ne.IsIn(ieSlice) {
				t.Errorf("Expected false, got true")
			}

		case FLOAT:
			// FLOAT ELEMENTS
			if !ie.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}
			// Test not equal
			if ie.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}
			// Test greater
			if ie.Greater(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if ie.Less(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !ie.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ie.LessEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ie.IsIn(ieSlice) {
				t.Errorf("Expected true, got false")
			}

			// FLOAT ELEMENT
			if !fe.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value %f and %f", fe.GetValue(), serie.Elements[0].ToFloat().GetValue())
			}
			// Test not equal
			if fe.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}
			// Test greater
			if fe.Greater(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if fe.Less(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !fe.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !fe.LessEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !fe.IsIn(ieSlice) {
				t.Errorf("Expected true, got false")
			}

			// BOOL ELEMENT
			if !be.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value")
			}
			// Test not equal
			if be.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}
			// STRING ELEMENT
			if se.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value")
			}

			if !se.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			// TIMESTAMP ELEMENT
			if te.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false for value")
			}

			// Null element
			if !ne.IsNull() {
				t.Errorf("Expected true, got false")
			}

			if !nue.IsNull() {
				t.Errorf("Expected true, got false")
			}

			if ne.Equal(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !ne.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if ne.Greater(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if ne.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ne.Less(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if !ne.LessEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			if ne.IsIn(ieSlice) {
				t.Errorf("Expected false, got true")
			}
		case BOOL:
			// BOOL ELEMENTS
			if !ie.Equal(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

			// Test not equal
			if ie.NotEqual(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			// Test greater
			if ie.Greater(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if ie.Less(serie.Elements[0]) {
				t.Errorf("Expected false, got true")
			}

			if !ie.GreaterEqual(serie.Elements[0]) {
				t.Errorf("Expected true, got false")
			}

		case STRING:
			// STRING ELEMENTS
			// TODO: Implement
		case TIMESTAMP:
			// TIMESTAMP ELEMENTS
			//TODO: Implement
		default:
			t.Errorf("Unknown type")
		}
	}
}

func TestSubSet(t *testing.T) {
	// Test subseries
	for _, serie := range testSeries {
		ss := serie.Subset([]int{1, 2})

		if ss.Len() != 2 {
			t.Errorf("Expected 2, got %d", ss.Len())
		}

		if !ss.Elements[0].Equal(serie.Elements[1]) {
			fmt.Printf("SubSerie: %v\n", ss.Elements[0])
			fmt.Printf("Serie: %v\n", serie.Elements[1])
			t.Errorf("Expected true, got false")
		}
	}
}

func TestIntFilter(t *testing.T) {
	// Test filters
	intSerie := testSeries[0]

	// Test int filter
	indexes, err := intSerie.Filter(EQ, Elem(1))
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 1 && indexes[0] != 1 {
		t.Errorf("Expected 1, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(NE, Elem(1))
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 2 && indexes[0] != 2 && indexes[1] != 3 {
		t.Errorf("Expected 2, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(GT, Elem(0))
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 3 && indexes[0] != 1 && indexes[1] != 2 && indexes[2] != 3 {
		t.Errorf("Expected 3, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(GE, Elem(1))
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 3 && indexes[0] != 1 && indexes[1] != 2 && indexes[2] != 3 {
		t.Errorf("Expected 3, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(LT, Elem(2))
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 1 && indexes[0] != 1 {
		t.Errorf("Expected 1, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(LE, Elem(2))
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 2 && indexes[0] != 1 && indexes[1] != 2 {
		t.Errorf("Expected 2, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(IN, []Element{Elem(1), Elem(2)})
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 2 && indexes[0] != 1 && indexes[1] != 2 {
		t.Errorf("Expected 2, got %d", len(indexes))
	}

	indexes, err = intSerie.Filter(NIN, []Element{Elem(1), Elem(2)})
	if err != nil {
		t.Errorf("Error filtering")
	}

	if len(indexes) != 1 && indexes[0] != 3 {
		t.Errorf("Expected 1, got %d", len(indexes))
	}

}

func TestApply(t *testing.T) {
	// Test apply
	intSerie := testSeries[0]

	tmpFunc := func(elem Element, inputs ...interface{}) (Element, error) {
		return Elem(0), nil
	}

	newSerie, err := intSerie.Apply(tmpFunc, "+", 1)

	if err != nil {
		t.Errorf("Error applying function")
	}

	if newSerie.Len() != 3 {
		t.Errorf("Expected 3, got %d", newSerie.Len())
	}

	if newSerie.Elements[0].GetValue() != 0 {
		t.Errorf("Expected 2, got %d", newSerie.Elements[0].GetValue())
	}

	mathFunc := func(elem Element, inputs ...interface{}) (Element, error) {
		inputsLen := len(inputs)

		if inputsLen == 0 {
			return nil, fmt.Errorf("No inputs")
		}

		switch inputs[0] {
		case "+":
			return Elem(elem.GetValue().(int) + inputs[1].(int)), nil
		case "-":
			return Elem(elem.GetValue().(int) - inputs[1].(int)), nil
		case "*":
			return Elem(elem.GetValue().(int) * inputs[1].(int)), nil
		case "/":
			return Elem(elem.GetValue().(int) / inputs[1].(int)), nil
		default:
			return nil, fmt.Errorf("Unknown operator")
		}
	}

	newSerie, err = intSerie.Apply(mathFunc, "+", 1)

	if err != nil {
		t.Errorf("Error applying function")
	}

	if newSerie.Len() != 3 {
		t.Errorf("Expected 3, got %d", newSerie.Len())
	}

	if newSerie.Elements[0].GetValue() != 2 {
		t.Errorf("Expected 2, got %d", newSerie.Elements[0].GetValue())
	}
}
