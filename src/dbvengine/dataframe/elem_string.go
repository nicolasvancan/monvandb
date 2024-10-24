package dataframe

import (
	"strconv"
	"strings"
	"time"
)

type StringElement struct {
	value string
	null  bool
}

func (e *StringElement) SetValue(value interface{}) {
	if value == nil {
		e.setNull()
		return
	}

	switch v := value.(type) {
	case string:
		e.value = v
		if strings.ToLower(e.value) == "null" {
			e.setNull()
		}
	case int:
		e.value = strconv.Itoa(v)
	case float64:
		e.value = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			e.value = "TRUE"
		} else {
			e.value = "FALSE"
		}
	case time.Time:
		e.value = v.Format("2006-01-02 15:04:05")
	case Element:
		e.value = v.String()
		if v.IsNull() {
			e.SetValue("null")
			e.setNull()
		}
	default:
		e.SetValue("null")
		e.setNull()
	}
}

func (e *StringElement) GetValue() interface{} {
	return e.value
}

func (e *StringElement) String() string {
	if e.null {
		return "null"
	}

	return e.value
}

func (e *StringElement) GetType() int {
	return STRING
}

func (e *StringElement) Copy() Element {
	return &StringElement{value: e.value, null: false}
}

func (e *StringElement) IsNull() bool {
	return e.value == "NULL"
}

func (e *StringElement) setNull() {
	e.value = "NULL"
	e.null = true
}

func (e *StringElement) Equal(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && elem.IsNull() {
			return true
		}

		return false
	}
	return e.GetValue() == elem.ToString().GetValue()
}

func (e *StringElement) NotEqual(elem Element) bool {

	return !e.Equal(elem)
}

func (e *StringElement) Greater(elem Element) bool {
	if e.IsNull() {
		return false
	}

	return e.value > elem.String()
}

func (e *StringElement) GreaterEqual(elem Element) bool {
	if e.IsNull() {
		return false
	}

	return e.value >= elem.String()
}

func (e *StringElement) Less(elem Element) bool {
	if e.IsNull() {
		return false
	}

	return e.value < elem.String()
}

func (e *StringElement) LessEqual(elem Element) bool {
	if e.IsNull() {
		return false
	}

	return e.value <= elem.String()
}

func (e *StringElement) IsIn(elems []Element) bool {
	for _, elem := range elems {
		if e.Equal(elem) {
			return true
		}
	}

	return false
}

func (e *StringElement) ToInt() Element {

	ie := &IntElement{}
	ie.SetValue(e.value)
	return ie
}

func (e *StringElement) ToFloat() Element {
	fe := &FloatElement{}
	fe.SetValue(e.value)
	return fe
}

func (e *StringElement) ToString() Element {
	return e.Copy()
}

func (e *StringElement) ToBool() Element {
	be := &BoolElement{}
	be.SetValue(e.value)

	return be
}
func (e *StringElement) ToTimestamp() Element {
	te := &TimestampElement{}
	te.SetValue(e.value)
	return te
}
