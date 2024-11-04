package dataframe

import (
	"fmt"
	"strconv"
	"strings"
)

type FloatElement struct {
	value float64
	null  bool
}

func (e *FloatElement) SetValue(value interface{}) {
	if value == nil {
		e.setNull()
		return
	}

	switch v := value.(type) {
	case string:
		if strings.ToLower(v) == "null" {
			e.setNull()
			return
		}

		parsedValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			e.setNull()
		} else {
			e.value = parsedValue
		}
	case int:
		e.value = float64(v)
	case float64:
		e.value = float64(v)
	case bool:
		if v {
			e.value = 1.0
		} else {
			e.value = 0.0
		}
	case Element:
		e.value = v.ToFloat().GetValue().(float64)
		if v.IsNull() {
			e.setNull()
		}
	default:
		e.setNull()
	}
}

func (e *FloatElement) GetValue() interface{} {
	return e.value
}

func (e *FloatElement) String() string {
	if e.null {
		return "null"
	}

	return fmt.Sprintf("%f", e.GetValue())
}

func (e *FloatElement) GetType() int {
	return FLOAT
}

func (e *FloatElement) Copy() Element {
	return &FloatElement{value: e.value, null: e.null}
}

func (e *FloatElement) IsNull() bool {
	return e.value == 0
}

func (e *FloatElement) setNull() {
	e.value = 0.0
	e.null = true
}

func (e *FloatElement) Equal(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	if e.value == elem.ToFloat().GetValue().(float64) {
		return true
	}

	return false
}
func (e *FloatElement) NotEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && elem.IsNull() {
			return false
		}
		return true
	}

	if e.value > elem.ToFloat().GetValue().(float64) {
		return true
	}

	return false
}
func (e *FloatElement) Greater(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if !e.IsNull() && elem.IsNull() {
			return false
		}
		return true
	}

	if e.value > elem.ToFloat().GetValue().(float64) {
		return true
	}

	return false
}
func (e *FloatElement) GreaterEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if !e.IsNull() && elem.IsNull() {
			return false
		}
		return true
	}

	if e.value >= elem.ToFloat().GetValue().(float64) {
		return true
	}

	return false
}
func (e *FloatElement) Less(elem Element) bool {

	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	if e.value < elem.ToFloat().GetValue().(float64) {
		return true
	}

	return false
}
func (e *FloatElement) LessEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	if e.value <= elem.ToFloat().GetValue().(float64) {
		return true
	}

	return false
}
func (e *FloatElement) IsIn(elem []Element) bool {
	if e.IsNull() {
		return false
	}

	for _, el := range elem {
		if e.Equal(el) {
			return true
		}
	}

	return false
}

func (e *FloatElement) ToInt() Element {
	ie := &IntElement{}
	ie.SetValue(e.value)
	ie.null = e.null
	return ie
}

func (e *FloatElement) ToFloat() Element {
	return e.Copy()
}

func (e *FloatElement) ToString() Element {
	se := &StringElement{}
	se.SetValue(e.value)
	se.null = e.null
	return se
}

func (e *FloatElement) ToBool() Element {
	be := &BoolElement{}
	be.SetValue(e.value)
	be.null = e.null
	return be
}
func (e *FloatElement) ToTimestamp() Element {
	te := &TimestampElement{}
	te.SetValue(e.value)
	te.null = e.null
	return te
}
