package dataframe

import (
	"fmt"
	"time"
)

// Basic types used to represent dataframes
const (
	STRING = iota
	INT
	FLOAT
	BOOL
	TIMESTAMP
	OBJECT
)

type ComparatorType string
type Indexes []int

type AppliableFunction func(elem Element, inputs ...interface{}) (Element, error)

// Comparator types
const (
	EQ      ComparatorType = "=="
	EQS     ComparatorType = "="
	NE      ComparatorType = "!="
	GT      ComparatorType = ">"
	GE      ComparatorType = ">="
	LT      ComparatorType = "<"
	LE      ComparatorType = "<="
	IN      ComparatorType = "in"
	LIKE    ComparatorType = "like"
	NIN     ComparatorType = "not in"
	BETWEEN ComparatorType = "between"
)

type Series struct {
	Elements []Element
	Type     int
}

func (s *Series) SetValues(values []Element) {
	s.Elements = values
}

func (s Series) String() string {
	str := ""
	for _, v := range s.Elements {
		str += " | " + fmt.Sprintf("%v ", v)
	}
	return str
}

func (s Series) Len() int {
	return len(s.Elements)
}

func NewSeries(values interface{}) Series {
	switch v := values.(type) {
	case int:
		return Series{
			Elements: make([]Element, v),
		}
	case []Element:
		// Creates a copy of the values
		copyElements := make([]Element, len(v))
		copy(copyElements, v)

		return Series{
			Elements: copyElements,
			Type:     v[0].GetType(),
		}
	case Series:
		// Creates a copy of the values
		return v.Copy()
	default:
		return Series{}
	}
}

func (s Series) Copy() Series {
	copiedElements := make([]Element, len(s.Elements))
	copy(copiedElements, s.Elements)
	return Series{
		Elements: copiedElements,
		Type:     s.Type,
	}
}

func (s Series) Concat(series Series) Series {
	newSeries := NewSeries(s)
	newSeries.Elements = append(newSeries.Elements, series.Elements...)
	return newSeries
}

func (s Series) GetType() int {
	return s.Type
}

func (s *Series) AddValue(value Element) {
	s.Elements = append(s.Elements, value)
}

// Set specific value in the series for given index
// if index is out of range, it returns an error
func (s Series) Set(index int, value interface{}) error {

	// Set whole series with same value
	if index == -1 {
		for i := 0; i < len(s.Elements); i++ {
			s.Elements[i] = castElement(Elem(value), s.GetType())
		}
		return nil
	}

	if index >= len(s.Elements) {
		return fmt.Errorf("index out of range for method Set Element in Series")
	}

	s.Elements[index] = castElement(Elem(value), s.GetType())
	return nil
}

func (s Series) Cast(dType int) Series {
	newSeries := NewSeries(s)
	newSeries.Type = dType
	for i, elem := range s.Elements {
		newSeries.Elements[i] = castElement(elem, dType)
	}

	return newSeries
}

func (s Series) Subset(indexes Indexes) Series {
	newSeries := NewSeries(len(indexes))
	newSeries.Type = s.GetType()

	for i, index := range indexes {
		newSeries.Set(i, s.Elements[index].GetValue())
	}

	return newSeries
}

// Filter is a method that filters the elements of the series based on the given value and comparator
// It returns the indexes of the elements that satisfy the condition
func (s Series) Filter(comparator ComparatorType, value interface{}) (Indexes, error) {
	indexes := make(Indexes, 0)

	switch v := value.(type) {
	case Series:
		if len(v.Elements) != len(s.Elements) {
			return Indexes{}, fmt.Errorf("series length not equal")
		}

		for i, elem := range s.Elements {
			switch comparator {
			case EQ:
				if elem.Equal(v.Elements[i]) {
					indexes = append(indexes, i)
				}
			case NE:
				if elem.NotEqual(v.Elements[i]) {
					indexes = append(indexes, i)
				}
			case GT:
				if elem.Greater(v.Elements[i]) {
					indexes = append(indexes, i)
				}
			case GE:
				if elem.GreaterEqual(v.Elements[i]) {
					indexes = append(indexes, i)
				}
			case LT:
				if elem.Less(v.Elements[i]) {
					indexes = append(indexes, i)
				}
			case LE:
				if elem.LessEqual(v.Elements[i]) {
					indexes = append(indexes, i)
				}
			default:
				return Indexes{}, fmt.Errorf("comparator not found for series")
			}
		}
	}

	for i, elem := range s.Elements {
		switch v := value.(type) {
		case Element:
			switch comparator {
			case EQ, EQS:
				if elem.Equal(v) {
					indexes = append(indexes, i)
				}
			case NE:
				if elem.NotEqual(v) {
					indexes = append(indexes, i)
				}
			case GT:
				if elem.Greater(v) {
					indexes = append(indexes, i)
				}
			case GE:
				if elem.GreaterEqual(v) {
					indexes = append(indexes, i)
				}
			case LT:
				if elem.Less(v) {
					indexes = append(indexes, i)
				}
			case LE:
				if elem.LessEqual(v) {
					indexes = append(indexes, i)
				}
			default:
				return Indexes{}, fmt.Errorf("comparator not found for value")
			}
		case []Element:
			switch comparator {
			case IN:
				if elem.IsIn(v) {
					indexes = append(indexes, i)
				}
			case NIN:
				if !elem.IsIn(v) {
					indexes = append(indexes, i)
				}
			case BETWEEN:
				if elem.GreaterEqual(v[0]) && elem.LessEqual(v[1]) {
					indexes = append(indexes, i)
				}
			default:
				return Indexes{}, fmt.Errorf("comparator not found for slice")
			}

		default:
			return Indexes{}, fmt.Errorf("value type not accepted in comparator")
		}
	}

	return indexes, nil
}

func (s Series) DeleteRow(index interface{}) (Series, error) {
	switch v := index.(type) {
	case int:
		if v >= len(s.Elements) {
			return Series{}, fmt.Errorf("index out of range")
		}
		s.Elements = append(s.Elements[:v], s.Elements[v+1:]...)
		return s, nil
	case []int:
		for _, i := range v {
			if i >= len(s.Elements) {
				return Series{}, fmt.Errorf("index out of range")
			}
		}
		for _, i := range v {
			s.Elements = append(s.Elements[:i], s.Elements[i+1:]...)
		}
		return s, nil
	default:
		return Series{}, fmt.Errorf("index type not accepted")
	}
}

func (s Series) Limit(n int) Series {
	newSeries := NewSeries(n)
	newSeries.Type = s.GetType()
	for i := 0; i < n; i++ {
		newSeries.Set(i, s.Elements[i].GetValue())
	}

	return newSeries
}

// Apply is a method that applies a function to the elements of the series
func (s Series) Apply(f AppliableFunction, args ...interface{}) (Series, error) {
	newSeries := NewSeries(s)
	newSeries.Type = s.GetType()
	for i, elem := range s.Elements {
		result, err := applyFunction(elem, f, args...)
		if err != nil {
			return Series{}, err
		}

		newSeries.Set(i, result.GetValue())
	}

	return newSeries, nil
}

// The Apply method applies a function to the elements of the series
func applyFunction(elem Element, f AppliableFunction, args ...interface{}) (Element, error) {
	result, err := f(elem, args...)
	if err != nil {
		return nil, err
	}

	return castElement(
		Elem(
			result.GetValue(),
		), elem.GetType()), nil
}

func castElement(elem Element, dType int) Element {
	switch dType {
	case STRING:
		return elem.ToString()
	case INT:
		return elem.ToInt()
	case FLOAT:
		return elem.ToFloat()
	case BOOL:
		return elem.ToBool()
	case TIMESTAMP:
		return elem.ToTimestamp()
	default:
		return elem.ToString()
	}
}

type Element interface {
	// Getter and setter
	SetValue(value interface{})
	GetValue() interface{}

	// Comparsion methods
	Equal(Element) bool
	NotEqual(Element) bool
	Greater(Element) bool
	GreaterEqual(Element) bool
	Less(Element) bool
	LessEqual(Element) bool
	IsIn([]Element) bool
	Copy() Element
	// String representation
	String() string
	// Cast methods
	ToInt() Element
	ToFloat() Element
	ToString() Element
	ToBool() Element
	ToTimestamp() Element

	// Generic Information
	IsNull() bool
	GetType() int
}

/* Elements Types */

/* End element types */

func Elem(value interface{}) Element {
	var elem Element
	switch v := value.(type) {
	case Element:
		// Copy it
		return v.Copy()
	case string:
		elem = &StringElement{}
		elem.SetValue(v)
		return elem
	case int:
		elem = &IntElement{}
		elem.SetValue(v)
		return elem
	case float64:
		elem = &FloatElement{}
		elem.SetValue(v)
		return elem
	case bool:
		elem = &BoolElement{}
		elem.SetValue(v)
		return elem
	case time.Time:
		elem = &TimestampElement{}
		elem.SetValue(v)
		return elem
	default:
		elem = &IntElement{}
		elem.SetValue(v)
	}

	return elem
}
