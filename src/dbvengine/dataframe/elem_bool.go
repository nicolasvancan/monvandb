package dataframe

import "strings"

type BoolElement struct {
	value bool
	null  bool
}

func (e *BoolElement) SetValue(value interface{}) {
	if value == nil {
		e.setNull()
		return
	}

	switch v := value.(type) {
	case bool:
		e.value = v
	case string:
		switch strings.ToLower(v) {
		case "true", "t", "1":
			e.value = true
		case "false", "f", "0":
			e.value = false
		default:
			e.setNull()
		}
	case int:
		if v == 1 {
			e.value = true
		}
		if v == 0 {
			e.value = false
		}
	case float64:
		if v > 0.0 {
			e.value = true
		} else {
			e.value = false
		}
	case Element:
		e.value = v.ToBool().GetValue().(bool)
		if v.IsNull() {
			e.setNull()
		}
	default:
		e.setNull()
	}
}

func (e *BoolElement) GetValue() interface{} {
	return e.value
}

func (e *BoolElement) String() string {
	if e.null {
		return "NULL"
	}

	if e.value {
		return "true"
	} else {
		return "false"
	}
}

func (e *BoolElement) GetType() int {
	return BOOL
}

func (e *BoolElement) Copy() Element {
	return &BoolElement{value: e.value, null: false}
}

func (e *BoolElement) IsNull() bool {
	return e.null
}

func (e *BoolElement) setNull() {
	e.value = false
	e.null = true
}

func (e *BoolElement) Equal(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return false
		}
		return true
	}

	if e.value == elem.ToBool().GetValue().(bool) {
		return true
	}

	return false
}
func (e *BoolElement) NotEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	if e.value != elem.ToBool().GetValue().(bool) {
		return true
	}

	return false
}
func (e *BoolElement) Greater(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	if e.value && !elem.ToBool().GetValue().(bool) {
		return true
	}

	return false
}
func (e *BoolElement) GreaterEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if e.IsNull() && !elem.IsNull() {
			return true
		}
		return false
	}

	return true
}
func (e *BoolElement) Less(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if !e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	if !e.value && elem.ToBool().GetValue().(bool) {
		return true
	}

	return false
}
func (e *BoolElement) LessEqual(elem Element) bool {
	if e.IsNull() || elem.IsNull() {
		if !e.IsNull() && elem.IsNull() {
			return true
		}
		return false
	}

	return true
}
func (e *BoolElement) IsIn(elem []Element) bool {
	if e.IsNull() {
		return false
	}

	for _, v := range elem {
		if e.Equal(v) {
			return true
		}
	}
	return false
}

func (e *BoolElement) ToInt() Element {
	ie := &IntElement{}
	ie.SetValue(e.value)

	if e.IsNull() {
		ie.setNull()
	}

	return ie
}

func (e *BoolElement) ToFloat() Element {
	fe := &FloatElement{}
	fe.SetValue(e.value)
	if e.IsNull() {
		fe.setNull()
	}
	return fe
}

func (e *BoolElement) ToString() Element {
	se := &StringElement{}
	se.SetValue(e.value)
	if e.IsNull() {
		se.setNull()
	}
	return se
}

func (e *BoolElement) ToBool() Element {
	return e.Copy()
}
func (e *BoolElement) ToTimestamp() Element {
	te := &TimestampElement{}
	te.SetValue(e.value)

	if e.IsNull() {
		te.setNull()
	}

	return te
}
