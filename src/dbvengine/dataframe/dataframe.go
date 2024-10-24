package dataframe

import (
	"fmt"
	"strings"
	"time"

	db "github.com/nicolasvancan/monvandb/src/database"
)

type AggType int
type Aggregation int

const (
	Or = iota
	And
)

type Dataframe struct {
	Columns []string
	Series  []Series // Map of column series
	DTypes  []int
	nRows   int
}

func (df Dataframe) String() string {
	finalString := "\n"

	for _, column := range df.Columns {
		tmp := "| "
		tmp += column
		if len(tmp) < 30 {
			for len(tmp) < 30 {
				tmp += " "
			}
		} else {
			tmp = tmp[:30]
			tmp += "..."
		}
		finalString += tmp
	}

	finalString += "|\n"

	final_string_len := len(finalString)
	for i := 0; i < final_string_len-2; i++ {
		finalString += "_"
	}

	finalString += "\n"

	for i := 0; i < df.Len(); i++ {
		for j := 0; j < len(df.Columns); j++ {
			value := df.Series[j].Elements[i]
			tmp := "| "
			tmp += value.String()
			if len(tmp) < 30 {
				for len(tmp) < 30 {
					tmp += " "
				}

			} else {
				tmp = tmp[:30]
				tmp += "..."
			}
			finalString += tmp
		}
		finalString += "|\n"
	}
	return finalString
}

func (df Dataframe) Len() int {
	return df.nRows
}

func (df Dataframe) IsEmpty() bool {
	return df.Len() == 0
}

func NewDataframe(rawRows []db.RawRow) Dataframe {
	columns := make([]string, 0)
	series := make([]Series, 0)
	dTypes := make([]int, 0)
	nRows := len(rawRows)

	// Iterate over each row or rawRows
	for _, row := range rawRows {
		// Iterate over each column in the row
		for column, value := range row {
			// Check if the column already exists
			index := indexOf(strings.ToLower(column), columns)
			if index == -1 {
				// No column found, add it
				columns = append(columns, strings.ToLower(column))
				index = len(columns) - 1
				// Create a new series for the column
				series = append(series, NewSeries(nil))
			}

			// Infer the schema of the value
			dType := inferSchema(value)
			if len(dTypes) < index+1 {
				dTypes = append(dTypes, dType)
				series[index] = series[index].Cast(dType)
			} else {
				// This avoid null values to change the type of the column
				if dTypes[index] != dType && dTypes[index] == STRING {
					dTypes[index] = dType
					series[index] = series[index].Cast(dType)
				}
			}

			// Add the value to the series
			series[index].AddValue(Elem(value))
		}
	}

	return Dataframe{
		Columns: columns,
		Series:  series,
		DTypes:  dTypes,
		nRows:   nRows,
	}
}

func (df Dataframe) GetColumn(column string) (Series, error) {
	index := indexOf(strings.ToLower(column), df.Columns)
	if index == -1 {
		return Series{}, fmt.Errorf("column %s not found", column)
	}

	//TODO: I don't know if it should come as a copy or as a reference
	newSeries := NewSeries(df.Series[index])

	return newSeries, nil
}

func inferSchema(value interface{}) int {
	switch value.(type) {
	case string:
		return STRING
	case int, int8, int16, int32, int64:
		return INT
	case float32, float64:
		return FLOAT
	case bool:
		return BOOL
	case time.Time:
		return TIMESTAMP
	default:
		return STRING
	}
}

func indexOf(element string, data []string) int {
	for i, v := range data {
		if v == element {
			return i
		}
	}
	return -1 // Return -1 if the element is not found
}

func (df Dataframe) Select(columns []string) (Dataframe, error) {

	newSeries := make([]Series, 0)
	dTypes := make([]int, 0)
	indexes := make([]int, len(columns))
	cols := make([]string, len(columns))
	// Fill indexes array and validate if there is a column that does not exist
	for i, column := range columns {
		indexOf := indexOf(strings.ToLower(column), df.Columns)

		// Returns if column is not found
		if indexOf == -1 {
			return Dataframe{}, fmt.Errorf("column %s not found", column)
		}
		indexes[i] = indexOf
		cols[i] = strings.ToLower(df.Columns[indexOf])
	}

	// Fill newRows and dTypes
	for _, col := range cols {
		colSeries, err := df.GetColumn(col)
		if err != nil {
			return Dataframe{}, err
		}

		newSeries = append(newSeries, NewSeries(colSeries))
	}

	return Dataframe{
		Columns: cols,
		Series:  newSeries,
		DTypes:  dTypes,
		nRows:   df.nRows,
	}, nil
}

type Filters struct {
	Column     string
	Comparator ComparatorType
	Value      interface{}
}

func (df Dataframe) Filter(agg Aggregation, filters ...Filters) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) Apply(column string, fn func(interface{}) interface{}) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) Drop(columns []string) (Dataframe, error) {

	return Dataframe{}, nil
}

func (df Dataframe) Join(df2 Dataframe, on []string, how string) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) GroupBy(columns []string) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) Sort(columns []string, ascending bool) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) Limit(n int) (Dataframe, error) {
	return Dataframe{}, nil
}
