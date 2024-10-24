package dataframe

import (
	"strings"
	"time"
)

type TimestampElement struct {
	value time.Time
	null  bool
}

func (e *TimestampElement) SetValue(value interface{}) {
	if value == nil {
		e.setNull()
		return
	}

	switch v := value.(type) {
	case string:
		if strings.ToLower(v) == "null" {
			e.setNull()
		} else {
			parsedValue, err := time.Parse(time.RFC3339, v)
			if err != nil {
				e.setNull()
			} else {
				e.value = parsedValue
				e.null = false
			}
		}
	case time.Time:
		e.value = v
		e.null = false
	case Element:
		e.value = v.ToTimestamp().GetValue().(time.Time)
		if v.IsNull() {
			e.setNull()
		}
	case int:
		e.value = time.Unix(int64(v), 0)
		e.null = false
	case float64:
		e.value = time.Unix(int64(v), 0)
		e.null = false
	default:
		e.setNull()
	}
}

func (e *TimestampElement) GetValue() interface{} {
	return e.value
}

func (e *TimestampElement) String() string {
	if e.null {
		return "null"
	}

	return e.value.Format(time.RFC3339)
}

func (e *TimestampElement) GetType() int {
	return TIMESTAMP
}

func (e *TimestampElement) Copy() Element {
	return &TimestampElement{value: e.value, null: false}
}

func (e *TimestampElement) IsNull() bool {
	return e.null
}

func (e *TimestampElement) setNull() {
	e.value = time.Time{}
	e.null = true
}

func (e *TimestampElement) Equal(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	if e.value == elem.ToTimestamp().GetValue().(time.Time) {
		return true
	}

	return false
}

func (e *TimestampElement) NotEqual(elem Element) bool {
	return !e.Equal(elem)
}
func (e *TimestampElement) Greater(elem Element) bool {
	if e.IsNull() {
		return false
	}

	if elem.IsNull() {
		return true
	}

	if e.value.After(elem.ToTimestamp().GetValue().(time.Time)) {
		return true
	}

	return false
}
func (e *TimestampElement) GreaterEqual(elem Element) bool {
	if e.Greater(elem) || e.Equal(elem) {
		return true
	}
	return false
}
func (e *TimestampElement) Less(elem Element) bool {
	if !e.Greater(elem) && !e.Equal(elem) {
		return true
	}

	return false
}
func (e *TimestampElement) LessEqual(elem Element) bool {
	if e.Less(elem) || e.Equal(elem) {
		return true
	}

	return false
}

func (e *TimestampElement) IsIn(elem []Element) bool {
	for _, e := range elem {
		if e.Equal(e) {
			return true
		}
	}

	return false
}

func (e *TimestampElement) ToInt() Element {

	ie := &IntElement{}
	ie.SetValue(e.value)
	return ie
}

func (e *TimestampElement) ToFloat() Element {
	fe := &FloatElement{}
	fe.SetValue(e.value)
	return fe
}

func (e *TimestampElement) ToString() Element {
	te := &TimestampElement{}
	te.SetValue(e.value)
	return te
}

func (e *TimestampElement) ToBool() Element {
	be := &BoolElement{}
	be.SetValue(e.value)

	return be
}
func (e *TimestampElement) ToTimestamp() Element {
	return e
}
