package dataframe

import (
	"testing"

	"github.com/nicolasvancan/monvandb/src/database"
)

func TestDataFrameCreation(t *testing.T) {
	// Test dataframe creation
	rawRows := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	df := NewDataframe(rawRows)
	df2, err := df.Select([]string{"Name", "age"})

	if df.Len() != 3 {
		t.Errorf("Expected 3 rows, got %d", df.Len())
	}

	if df.IsEmpty() {
		t.Errorf("Dataframe should not be empty")
	}

	if err != nil {
		t.Errorf("Error selecting columns")
	}

	if len(df2.Columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(df.Columns))
	}
}

func TestGetColumn(t *testing.T) {
	// Test get column
	rawRows := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	df := NewDataframe(rawRows)
	col, err := df.GetColumn("name")
	if err != nil {
		t.Errorf("Error getting column")
	}

	if len(col.Elements) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(col.Elements))
	}

	if col.Elements[0].GetValue() != "Nicolas" {
		t.Errorf("Expected Nicolas, got %s", col.Elements[0].GetValue())
	}
}
