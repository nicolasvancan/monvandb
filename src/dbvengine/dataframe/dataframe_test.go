package dataframe

import (
	"fmt"
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

	df2, err := df.Select([]SelectColumnInput{{Column: "Name"}, {Column: "age"}})
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

func TestDataframeFilter(t *testing.T) {
	// Test dataframe filter
	rawRows := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	df := NewDataframe(rawRows)
	filters := NewFilter()
	filters.InsertValue(0, FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "age"}, Comparator: GT, Comparando: 50})
	filters.InsertValue(0, FilterValue{Type: FilterTypeOr, Column: ColumnFilter{Name: "age"}, Comparator: EQ, Comparando: 25})
	df2, err := df.Filter(filters)

	if err != nil {
		t.Errorf("Error filtering dataframe")
	}

	if df2.Len() != 1 {
		t.Errorf("Expected 3 rows, got %d", df2.Len())
	}

	df3 := NewDataframe(rawRows)

	filters = NewFilter()
	filters.InsertValue(0, FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "name"}, Comparator: IN, Comparando: []string{"Nicolas", "John"}})
	df4, err := df3.Filter(filters)

	if err != nil {
		t.Error("Expected no error")
	}

	if df4.Len() != 2 {
		t.Error("Expected two values")
	}

	filters = NewFilter()
	filters.InsertValue(0, FilterValue{Type: FilterTypeAnd, Column: ColumnFilter{Name: "age"}, Comparator: BETWEEN, Comparando: []int{1, 26}})

	df5, err := df3.Filter(filters)

	if err != nil {
		t.Error("Expected no error")
	}

	if df5.Len() != 1 {
		t.Error("Expected one value")
	}
}

func TestInnerJoin(t *testing.T) {
	// Test inner join
	rawRows1 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	rawRows2 := []database.RawRow{
		{"name": "Nicolas", "age": 21, "height1": 1.75},
		{"name": "John", "age": 30, "height1": nil},
		{"name": "Jane", "age": 35, "height1": 1.70},
	}

	df1 := NewDataframe(rawRows1)
	df2 := NewDataframe(rawRows2)
	df2.Alias = "t"
	joined, err := df1.Join(df2, []string{"name", "age"}, "inner")

	if err != nil {
		t.Errorf("Expected no error got %v", err)
	}

	if joined.Len() != 2 {
		t.Errorf("Expected 3 rows found %v columns %v", joined.Len(), joined.Columns)
	}

	// If all columns are in
	for _, col := range joined.Columns {
		isIn := false
		for _, col2 := range []string{"name", "age", "height", "height1"} {
			if col2 == col.Info.Name {
				isIn = true
				break
			}
		}
		if !isIn {
			t.Errorf("column %s not found", col.Info.Name)
		}
	}
	fmt.Printf("%v\n", joined)
}

func TestLeftJoin(t *testing.T) {
	// Test inner join
	rawRows1 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	rawRows2 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height1": 1.75},
		{"name": "John", "age": 30, "height1": nil},
		{"name": "Peter", "age": 35, "height1": 1.70},
	}

	df1 := NewDataframe(rawRows1)
	df2 := NewDataframe(rawRows2)
	df2.Alias = "t"
	joined, err := df1.Join(df2, []string{"name", "age"}, "left")

	if err != nil {
		t.Errorf("Expected no error got %v", err)
	}

	if joined.Len() != 3 {
		t.Errorf("Expected 3 rows found %v columns %v", joined.Len(), joined.Columns)
	}

	// If all columns are in
	for _, col := range joined.Columns {
		isIn := false
		for _, col2 := range []string{"name", "age", "height", "height1"} {
			if col2 == col.Info.Name {
				isIn = true
				break
			}
		}
		if !isIn {
			t.Errorf("column %s not found", col.Info.Name)
		}
	}
}

func TestRightJoin(t *testing.T) {
	// Test inner join
	rawRows1 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	rawRows2 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height1": 1.75},
		{"name": "John", "age": 30, "height1": nil},
		{"name": "Peter", "age": 35, "height1": 1.70},
	}

	df1 := NewDataframe(rawRows1)
	df2 := NewDataframe(rawRows2)
	df2.Alias = "t"
	joined, err := df1.Join(df2, []string{"name", "age"}, "right")

	if err != nil {
		t.Errorf("Expected no error got %v", err)
	}

	if joined.Len() != 3 {
		t.Errorf("Expected 3 rows found %v columns %v", joined.Len(), joined.Columns)
	}

	// If all columns are in
	for _, col := range joined.Columns {
		isIn := false
		for _, col2 := range []string{"name", "age", "height", "height1"} {
			if col2 == col.Info.Name {
				isIn = true
				break
			}
		}
		if !isIn {
			t.Errorf("Column %s not found", col.Info.Name)
		}
	}
}

func TestOuterJoin(t *testing.T) {
	// Test inner join
	rawRows1 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
	}

	rawRows2 := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height1": 1.75},
		{"name": "John", "age": 30, "height1": nil},
		{"name": "Peter", "age": 35, "height1": 1.70},
	}

	df1 := NewDataframe(rawRows1)
	df2 := NewDataframe(rawRows2)
	df2.Alias = "t"
	joined, err := df1.Join(df2, []string{"name", "age"}, "outer")

	if err != nil {
		t.Errorf("Expected no error got %v", err)
	}

	if joined.Len() != 4 {
		t.Errorf("Expected 3 rows found %v and columns %v", joined.Len(), joined.Columns)
	}

	// If all columns are in
	for _, col := range joined.Columns {
		isIn := false
		for _, col2 := range []string{"name", "age", "height", "height1"} {
			if col2 == col.Info.Name {
				isIn = true
				break
			}
		}
		if !isIn {
			t.Errorf("column %s not found", col.Info.Name)
		}
	}
}

func TestSort(t *testing.T) {
	// Test sort
	rawRows := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
		{"name": "Peter", "age": 35, "height": 1.78},
	}

	df := NewDataframe(rawRows)
	df2, err := df.Sort([]string{"age"}, true)

	if err != nil {
		t.Errorf("Error sorting")
	}

	if df2.getRow(0)[indexOf("age", df2.Columns)].GetValue() != 25 {
		t.Errorf("Expected 25, got %d", df2.getRow(0)[indexOf("age", df2.Columns)].GetValue())
	}

	df3, err := df.Sort([]string{"age", "height"}, false)

	if err != nil {
		t.Errorf("Error sorting")
	}

	if df3.getRow(0)[indexOf("name", df2.Columns)].GetValue() != "Peter" {
		t.Errorf("Expected 35, got %d", df3.getRow(0)[indexOf("name", df2.Columns)].GetValue())
	}

	fmt.Printf("%v\n", df3)
}

func TestLimit(t *testing.T) {
	// Test limit
	rawRows := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
		{"name": "Peter", "age": 35, "height": 1.78},
	}

	df := NewDataframe(rawRows)
	df2, err := df.Limit(2)

	if err != nil {
		t.Errorf("Error limiting")
	}

	if df2.Len() != 2 {
		t.Errorf("Expected 2 rows, got %d", df2.Len())
	}
}

func TestGroupBy(t *testing.T) {
	// Test groupby
	rawRows := []database.RawRow{
		{"name": "Nicolas", "age": 25, "height": 1.75},
		{"name": "John", "age": 30, "height": nil},
		{"name": "Jane", "age": 35, "height": 1.70},
		{"name": "Peter", "age": 35, "height": 1.78},
	}

	df := NewDataframe(rawRows)
	grouped, err := df.GroupBy([]string{"age"}, []GroupByAgg{{Column: "height", Agg: "max", As: "max"}, {Column: "height", Agg: "min", As: "min"}})
	fmt.Printf("%v\n", grouped)
	if err != nil {
		t.Errorf("Error grouping")
	}

	if grouped.Len() != 3 {
		t.Errorf("Expected 3 rows, got %d", grouped.Len())
	}
}
