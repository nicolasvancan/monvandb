package dataframe

import (
	"fmt"
	"strconv"
	"time"
)

type IntElement struct {
	value int
	null  bool
}

func (e *IntElement) SetValue(value interface{}) {
	if value == nil {
		e.setNull()
		return
	}

	switch v := value.(type) {
	case int:
		e.value = v
	case float64:
		e.value = int(v)
	case string:
		parsedValue, err := strconv.Atoi(v)
		if err != nil || v == "NULL" {
			e.setNull()
		} else {
			e.value = parsedValue
		}
	case bool:
		if v {
			e.value = 1
		} else {
			e.value = 0
		}
	case Element:
		e.value = v.ToInt().GetValue().(int)
		if v.IsNull() {
			e.setNull()
		}
	case time.Time:
		e.value = int(v.Unix())
	default:
		e.setNull()
	}

}

func (e *IntElement) GetValue() interface{} {
	return e.value
}

func (e *IntElement) String() string {
	if e.null {
		return "null"
	}

	return fmt.Sprintf("%d", e.GetValue())
}

func (e *IntElement) GetType() int {
	return INT
}

func (e *IntElement) Copy() Element {
	return &IntElement{value: e.value, null: e.null}
}

func (e *IntElement) IsNull() bool {
	return e.value == 0
}

func (e *IntElement) setNull() {
	e.value = 0
	e.null = true
}

func (e *IntElement) Equal(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	if e.value == elem.GetValue() {
		return true
	}

	return false
}
func (e *IntElement) NotEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && elem.IsNull() {
			return false
		}
		return true
	}

	if e.value != elem.ToInt().GetValue() {
		return true
	}

	return false
}
func (e *IntElement) Greater(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if !e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	if e.value > elem.ToInt().GetValue().(int) {
		return true
	}

	return false
}
func (e *IntElement) GreaterEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if !e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	if e.value >= elem.ToInt().GetValue().(int) {
		return true
	}

	return false
}
func (e *IntElement) Less(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	if e.value < elem.ToInt().GetValue().(int) {
		return true
	}

	return false
}
func (e *IntElement) LessEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	if e.value <= elem.ToInt().GetValue().(int) {
		return true
	}

	return false
}
func (e *IntElement) IsIn(elem []Element) bool {
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

func (e *IntElement) ToInt() Element {
	return e.Copy()
}

func (e *IntElement) ToFloat() Element {
	fe := &FloatElement{}
	fe.SetValue(e.value)
	fe.null = e.null
	return fe
}

func (e *IntElement) ToString() Element {
	se := &StringElement{}
	se.SetValue(e.value)
	se.null = e.null
	return se
}

func (e *IntElement) ToBool() Element {
	be := &BoolElement{}
	be.SetValue(e.value)
	be.null = e.null
	return be
}
func (e *IntElement) ToTimestamp() Element {
	te := &TimestampElement{}
	te.SetValue(e.value)
	te.null = e.null
	return te
}
