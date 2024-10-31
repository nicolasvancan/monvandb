package dataframe

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	db "github.com/nicolasvancan/monvandb/src/database"
)

type AggType int
type Aggregation int

type Dataframe struct {
	Columns []string
	Series  []Series // Map of column series
	DTypes  []int
	nRows   int
	Alias   string
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

func (df Dataframe) getRow(index int) []Element {
	row := make([]Element, 0)
	for _, serie := range df.Series {
		row = append(row, serie.Elements[index])
	}

	return row
}

func (df Dataframe) AddRow(row []Element) {
	for i, value := range row {
		df.Series[i].AddValue(value)
	}
}

// Since Dataframe will be in memory, we can use a simple iterator
// to iterate over the rows and avoid copying the data again to complete
// joins operations
type RowsIterator struct {
	pDf *Dataframe
	i   int
}

func NewRowsIterator(df *Dataframe) *RowsIterator {
	return &RowsIterator{
		pDf: df,
		i:   0,
	}
}

func (ri *RowsIterator) Next() []Element {
	df := *ri.pDf
	row := df.getRow(ri.i)
	ri.i += 1

	return row
}

func (ri *RowsIterator) HasNext() bool {
	df := *ri.pDf
	return ri.i < df.Len()
}

func NewDataframe(rawRows interface{}) Dataframe {
	switch v := rawRows.(type) {
	case []db.RawRow:
		return newDataframeRawRow(v)
	case [][]Element:
		return newDataFrameFromMatrixElements(v)
	case []string:
		return newDataFrameFromColumnsList(v)
	default:
		return Dataframe{}
	}
}

func (df *Dataframe) setDtypes(dTypes []int) {
	df.DTypes = dTypes

	for i, _ := range df.Series {
		df.Series[i].Type = dTypes[i]
	}
}

func newDataFrameFromColumnsList(cols []string) Dataframe {
	columns := make([]string, 0)
	series := make([]Series, 0)
	dTypes := make([]int, 0)
	nRows := 0

	for _, col := range cols {
		columns = append(columns, col)
		series = append(series, NewSeries(nil))
		dTypes = append(dTypes, STRING)
	}

	return Dataframe{
		Columns: columns,
		Series:  series,
		DTypes:  dTypes,
		nRows:   nRows,
		Alias:   "",
	}
}

func newDataFrameFromMatrixElements(rows [][]Element) Dataframe {
	series := make([]Series, 0)
	columns := make([]string, 0)
	dTypes := make([]int, 0)
	nRows := len(rows)

	// Iterate over each row or rawRows
	for _, row := range rows {
		// Iterate over each column in the roww
		for i, value := range row {
			// Fill columns
			if len(columns) < len(row) {
				columns = append(columns, fmt.Sprintf("column_%d", i))
				series = append(series, NewSeries(nil))
			}
			// Fill dTypes
			if len(dTypes) < len(row) {
				dTypes = append(dTypes, value.GetType())
				series[i].Type = value.GetType()
			}

			series[i].AddValue(value)
		}
	}

	return Dataframe{
		Columns: columns,
		Series:  series,
		DTypes:  dTypes,
		nRows:   nRows,
		Alias:   "",
	}
}

func newDataframeRawRow(rawRows []db.RawRow) Dataframe {
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
		Alias:   "",
	}
}

func (df Dataframe) SetAlias(alias string) {
	df.Alias = alias
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

func (df Dataframe) SetColumn(column string, series Series) error {
	index := indexOf(strings.ToLower(column), df.Columns)
	if index == -1 {
		return fmt.Errorf("column %s not found", column)
	}

	if series.Len() != df.Len() {
		return fmt.Errorf("series length is different from dataframe length")
	}

	df.Series[index] = series
	return nil
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

// Filter DataFrame based on given filters
func (df Dataframe) Filter(filters Filters) (Dataframe, error) {
	resolvedFilters := filters.Resolve()
	tmpDf := df

	for _, filter := range resolvedFilters {
		filterType := filter.Type
		res, err := filterColumn(tmpDf, filter)
		if err != nil {
			return Dataframe{}, err
		}

		// If the dataframe is empty, we return it
		if tmpDf.IsEmpty() {
			return tmpDf, nil
		}

		// Treat filter type when case is OR
		if filterType == FilterTypeOr {
			tmpDf, err = tmpDf.Concat(res)
			if err != nil {
				return Dataframe{}, err
			}
			continue
		}

		tmpDf = res
	}

	return tmpDf, nil
}

func isSlice(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

func (df Dataframe) indexes(indexes []int) Dataframe {
	newSeries := make([]Series, 0)
	for _, serie := range df.Series {
		newSeries = append(newSeries, serie.Subset(indexes))
	}

	return Dataframe{
		Columns: df.Columns,
		Series:  newSeries,
		DTypes:  df.DTypes,
		nRows:   len(indexes),
	}
}

func sliceToElementSlice(slice interface{}) ([]Element, error) {
	t := make([]Element, 0)

	switch v := slice.(type) {
	case []Element:
		return v, nil
	case []int:
		for _, value := range v {
			t = append(t, Elem(value))
		}
	case []float64:
		for _, value := range v {
			t = append(t, Elem(value))
		}
	case []string:
		for _, value := range v {
			t = append(t, Elem(value))
		}
	case []time.Time:
		for _, value := range v {
			t = append(t, Elem(value))
		}
	case []bool:
		for _, value := range v {
			t = append(t, Elem(value))
		}

	default:
		return nil, fmt.Errorf("slice type not supported")
	}

	return t, nil
}

func filterColumn(df Dataframe, filter ResolvedFilters) (Dataframe, error) {
	res := df
	allIndexes := make([]int, 0)
	for _, value := range filter.Values {
		columnSeries, err := df.GetColumn(value.Column.Name)

		if err != nil {
			return Dataframe{}, err
		}

		// Apply filter to the column
		if value.Column.Function != "" {
			// Apply function to the Series
		} else {
			// Verify if it is a slice type
			isSl := isSlice(value.Comparando)
			// If it is a slice type, convert it to Element
			if isSl {
				fmt.Printf("Value: %v\n", value)
				fmt.Printf("Comparando: %v\n", value.Comparando)

				value.Comparando, err = sliceToElementSlice(value.Comparando)
				if err != nil {
					return Dataframe{}, err
				}

			} else {
				// Convert any value into element
				switch v := value.Comparando.(type) {
				case ColumnFilter:
					serie, err := res.GetColumn(v.Name)
					if err != nil {
						return Dataframe{}, err
					}

					value.Comparando = serie
				default:
					value.Comparando = Elem(v)
				}
			}

			indexes, err := columnSeries.Filter(value.Comparator, value.Comparando)

			if err != nil {
				return Dataframe{}, err
			}
			// Filter the dataframe
			allIndexes = append(allIndexes, indexes...)
		}
	}

	if len(allIndexes) == 0 {
		return Dataframe{}, nil
	}

	res = res.indexes(allIndexes)

	return res, nil
}

func (df Dataframe) Apply(column string, fn AppliableFunction) (Dataframe, error) {
	col, err := df.GetColumn(column)

	if err != nil {
		return Dataframe{}, err
	}

	newCol, err := col.Apply(fn)

	if err != nil {
		return Dataframe{}, err
	}

	err = df.SetColumn(column, newCol)

	if err != nil {
		return Dataframe{}, err
	}

	return Dataframe{}, nil
}

func (df Dataframe) Concat(dataframe Dataframe) (Dataframe, error) {
	newSeriesSize := 0
	for _, column := range dataframe.Columns {
		serieToBeConcatenated, err := dataframe.GetColumn(column) // Copy the Series
		if err != nil {
			return Dataframe{}, err
		}

		serie, err := df.GetColumn(column)

		if err != nil {
			return Dataframe{}, err
		}

		serie.Concat(serieToBeConcatenated)
		df.nRows = serie.Len()
		df.SetColumn(column, serie)

		if newSeriesSize > 0 && newSeriesSize != serie.Len() {
			return Dataframe{}, fmt.Errorf("series have different lengths")
		}
	}

	return Dataframe{}, nil
}

func (df Dataframe) Drop(columns []string) (Dataframe, error) {

	return Dataframe{}, nil
}

func (df Dataframe) Join(df2 Dataframe, on []string, how string) (Dataframe, error) {
	switch how {
	case "inner":
		return innerJoin(df, df2, on)
	case "left":
		return leftJoin(df, df2, on)
	case "right":
		return rightJoin(df, df2, on)
	case "outer":
		return outerJoin(df, df2, on)
	default:
		return Dataframe{}, fmt.Errorf("join type not supported")
	}
}

func (df Dataframe) GroupBy(columns []string) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) Sort(columns []string, ascending bool) (Dataframe, error) {
	return Dataframe{}, nil
}

func (df Dataframe) Limit(n int) (Dataframe, error) {

	newDf := Dataframe{}
	newDf.Columns = df.Columns
	newDf.DTypes = df.DTypes
	newDf.Alias = df.Alias
	newDf.Series = make([]Series, 0)

	for _, serie := range df.Series {
		newSerie := serie.Limit(n)
		newDf.Series = append(newDf.Series, newSerie)
	}

	return newDf, nil
}
