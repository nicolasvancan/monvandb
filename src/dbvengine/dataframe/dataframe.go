package dataframe

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	db "github.com/nicolasvancan/monvandb/src/database"
)

type AggType int
type Aggregation int

type SelectColumnInput struct {
	Column string
	Alias  string
}

// Abstraction of a column name, such as t.col1
type DFColInfo struct {
	Name      string
	Qualifier string
}

type DFColumn struct {
	Info  DFColInfo
	Serie Series
	DType int
}

type Dataframe struct {
	Columns []DFColumn
	nRows   int
	Alias   string
}

type DfOn struct {
	Left     string
	Right    string
	Operator string
}

func fromStringSliceToDFColumnSlice(slice []string) []DFColumn {
	columns := make([]DFColumn, 0)
	for _, col := range slice {
		// Evaluate if the column has a qualifier
		if strings.Contains(col, ".") {
			splitted := strings.Split(col, ".")
			columns = append(columns, DFColumn{
				Info: DFColInfo{
					Name:      splitted[1],
					Qualifier: splitted[0],
				},
				Serie: NewSeries(nil),
				DType: STRING,
			})
			continue
		}

		columns = append(columns, DFColumn{Info: DFColInfo{Name: col, Qualifier: ""}})
	}

	return columns
}

func DataframeFromElementsSlice(rows []HashedElementSlice, columns []string, Dtypes []int) Dataframe {
	dataframe := Dataframe{}
	dataframe.Columns = fromStringSliceToDFColumnSlice(columns)

	for _, dType := range Dtypes {
		dataframe.Columns[dType].DType = dType
		dataframe.Columns[dType].Serie = NewSeries(nil)
		dataframe.Columns[dType].Serie.Type = dType
	}

	dataframe.nRows = len(rows)

	for _, row := range rows {
		for i, value := range row {
			dataframe.Columns[i].Serie.AddValue(value)
		}
	}

	return dataframe
}

func (df Dataframe) String() string {
	finalString := "\n"

	// Print columns
	for _, column := range df.Columns {
		tmp := "| "

		if column.Info.Qualifier != "" {
			tmp += column.Info.Qualifier + "." + column.Info.Name
		} else {
			tmp += column.Info.Name
		}

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

	// Print separator
	final_string_len := len(finalString)
	for i := 0; i < final_string_len-2; i++ {
		finalString += "_"
	}

	finalString += "\n"

	// Print values
	for i := 0; i < df.Len(); i++ {
		for j := 0; j < len(df.Columns); j++ {
			value := df.Columns[j].Serie.Elements[i]
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

func (df Dataframe) Values() [][]Element {
	values := make([][]Element, 0)

	for _, col := range df.Columns {
		tmp := make([]Element, 0)
		tmp = append(tmp, col.Serie.Elements...)
		values = append(values, tmp)
	}

	return values
}

func (df Dataframe) Len() int {
	return df.nRows
}

func (df Dataframe) IsEmpty() bool {
	return df.Len() == 0
}

func (df Dataframe) getRow(index int) []Element {
	row := make([]Element, 0)
	for _, col := range df.Columns {
		row = append(row, col.Serie.Elements[index])
	}

	return row
}

func (df Dataframe) setRow(index int, row []Element) error {
	if len(row) != len(df.Columns) {
		return fmt.Errorf("row length is different from dataframe columns")
	}

	for i, value := range row {
		df.Columns[i].Serie.Elements[index] = value
	}

	return nil
}

func (df *Dataframe) AddRow(row []Element) {
	for i, value := range row {
		df.Columns[i].Serie.AddValue(value)
	}
	df.nRows = df.nRows + 1
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

func (df *Dataframe) setDtypes(dTypes []int) error {

	if len(dTypes) != len(df.Columns) {
		return fmt.Errorf("dTypes length is different from dataframe columns")
	}

	for i, _ := range df.Columns {
		df.Columns[i].Serie.Type = dTypes[i]
	}

	return nil
}

func newDataFrameFromColumnsList(cols []string) Dataframe {
	newDf := Dataframe{}
	newDf.Columns = fromStringSliceToDFColumnSlice(cols)
	newDf.nRows = 0
	newDf.Alias = ""

	return newDf
}

func newDataFrameFromMatrixElements(rows [][]Element) Dataframe {
	columns := make([]DFColumn, 0)
	nRows := 0
	// Iterate over each row or rawRows
	for _, row := range rows {
		// Iterate over each column in the roww
		for i, value := range row {
			// Fill columns
			if len(columns) < len(row) {
				dfCol := DFColumn{}
				dfCol.Info.Name = fmt.Sprintf("column_%d", i)
				dfCol.Serie = NewSeries(nil)
				dfCol.Serie.Type = value.GetType()
				dfCol.DType = value.GetType()
				dfCol.Info.Qualifier = ""
				columns = append(columns, dfCol)
			}

			columns[i].Serie.AddValue(value)
			nRows += 1
		}
	}

	return Dataframe{
		Columns: columns,
		nRows:   nRows,
		Alias:   "",
	}
}

func newDataframeRawRow(rawRows []db.RawRow) Dataframe {
	columns := make([]DFColumn, 0)
	nRows := 0
	// Iterate over each row or rawRows
	for _, row := range rawRows {
		// Iterate over each column in the row
		for column, value := range row {
			// Check if the column already exists
			index := indexOf(strings.ToLower(column), columns)
			elem := Elem(value)
			if index == -1 {
				// No column found, add it
				dfCol := DFColumn{}
				dfCol.Info.Name = column
				dfCol.Info.Qualifier = ""
				dfCol.DType = elem.GetType()
				dfCol.Serie = NewSeries(nil)
				dfCol.Serie.AddValue(elem)
				columns = append(columns, dfCol)
				continue
			}

			// Infer the schema of the value

			dType := elem.GetType()

			// This avoid null values to change the type of the column
			if columns[index].DType != dType && columns[index].DType == STRING {
				columns[index].DType = dType
				columns[index].Serie = columns[index].Serie.Cast(dType)
			}

			// Add the value to the series
			columns[index].Serie.AddValue(elem)
		}
		nRows += 1

	}

	return Dataframe{
		Columns: columns,
		nRows:   nRows,
		Alias:   "",
	}
}

// Set qualifier replacing old qualifiers
func (df Dataframe) SetQualifier(qualifier string) {
	for i := range df.Columns {
		df.Columns[i].Info.Qualifier = qualifier
	}
}

func (df Dataframe) GetColumn(column string) (Series, error) {
	index := indexOf(strings.ToLower(column), df.Columns)

	if index == -1 {
		return Series{}, fmt.Errorf("column %s not found", column)
	}

	//TODO: I don't know if it should come as a copy or as a reference
	newSeries := NewSeries(df.Columns[index].Serie)

	return newSeries, nil
}

func (df Dataframe) SetColumn(column string, series interface{}) error {
	colIndex := indexOf(strings.ToLower(column), df.Columns)
	if colIndex == -1 {
		return fmt.Errorf("column %s not found", column)
	}

	switch v := series.(type) {
	case Series:

		if v.Len() != df.Len() {
			return fmt.Errorf("series length is different from dataframe length")
		}

		df.Columns[colIndex].Serie = v
	case string:
		index := indexOf(strings.ToLower(v), df.Columns)
		if index == -1 {
			// find column index
			df.Columns[colIndex].Serie.Set(-1, v)
			return nil
		}
	default:
		df.Columns[colIndex].Serie.Set(-1, v)
	}

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

func indexOf(element string, data []DFColumn) int {
	// Search for the element in data
	splitedElement := strings.Split(element, ".")
	qualifier := ""
	colName := element

	if len(splitedElement) > 1 {
		qualifier = splitedElement[0]
		colName = splitedElement[1]
	}

	for i, v := range data {
		if qualifier != "" {
			if v.Info.Qualifier == qualifier && v.Info.Name == colName {
				return i
			}
		}

		if v.Info.Name == element {
			return i
		}
	}

	return -1 // Return -1 if the element is not found
}

func (df Dataframe) Select(columnsInputs []SelectColumnInput) (Dataframe, error) {

	columns := make([]DFColumn, 0)
	// Fill indexes array and validate if there is a column that does not exist
	for _, column := range columnsInputs {
		if strings.Contains(column.Column, "*") {
			// Verify if there is a qualifier
			if strings.Contains(column.Column, ".") {
				qualifier := strings.Split(column.Column, ".")[0]
				for _, col := range df.Columns {
					if col.Info.Qualifier == qualifier {
						columns = append(columns, col)
					}
				}
				continue
			}

			columns = append(columns, df.Columns...)
			break
		}

		indexOf := indexOf(strings.ToLower(column.Column), df.Columns)

		// Returns if column is not found
		if indexOf == -1 {
			return Dataframe{}, fmt.Errorf("column %s not found", column)
		}

		columns = append(columns, df.Columns[indexOf])
	}

	return Dataframe{
		Columns: columns,
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
	newColumns := make([]DFColumn, 0)
	for _, col := range df.Columns {
		col.Serie = col.Serie.Subset(indexes).Copy()
		newColumns = append(newColumns, col)
	}

	return Dataframe{
		Columns: newColumns,
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
		fullColName := column.Info.Name

		if column.Info.Qualifier != "" {
			fullColName = column.Info.Qualifier + "." + column.Info.Name
		}

		serieToBeConcatenated, err := dataframe.GetColumn(fullColName) // Copy the Series

		if err != nil {
			return Dataframe{}, err
		}

		serie, err := df.GetColumn(fullColName)

		if err != nil {
			return Dataframe{}, err
		}

		serie.Concat(serieToBeConcatenated)
		df.nRows = serie.Len()

		df.SetColumn(fullColName, serie)

		if newSeriesSize > 0 && newSeriesSize != serie.Len() {
			return Dataframe{}, fmt.Errorf("series have different lengths")
		}
	}

	return Dataframe{}, nil
}

func (df Dataframe) Drop(columns []string) (Dataframe, error) {

	return Dataframe{}, nil
}

func (df Dataframe) Join(df2 Dataframe, on JoinOn, how string) (Dataframe, error) {
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

func (df Dataframe) GroupBy(columns []string, aggregators []GroupByAgg) (Dataframe, error) {
	// Create a hashed table containing rows of columns in columns field and add the columns values and aggregator values
	// when there are some
	//
	// Example:
	//
	// columns = ["name", "age"]
	// aggregators = [{"column": "height", "agg": "max"}, {"column": "height", "agg": "min"}]
	// the columns within he Dataframe will be only three, name, age and height
	finalDf := Dataframe{}
	hashedTable := createHashTableForColumnsAndAggregators(df, columns, aggregators)

	// Now its time to convert all the hashed values into a new Dataframe and aggregate values
	// based on the aggregators
	// If there is no aggregator, we get only the first row of the dataframe
	allColumns := columns
	DTypes := make([]int, 0)
	for _, agg := range aggregators {
		allColumns = append(allColumns, agg.Column)
	}

	for _, col := range allColumns {
		index := indexOf(col, df.Columns)

		if index == -1 {
			return Dataframe{}, fmt.Errorf("column %s not found in df1", col)
		}

		DTypes = append(DTypes, df.Columns[index].DType)
	}

	// Bind output dataframe informations
	aggregate := len(aggregators) > 0

	if !aggregate {
		colDfs := fromStringSliceToDFColumnSlice(allColumns)

		for i, colDf := range colDfs {
			colDf.Serie = NewSeries(nil)
			colDf.Serie.Type = DTypes[i]
		}

		finalDf.Columns = colDfs
	} else {
		// Create new columns for the aggregators
		// The types are copied from the original column
		// The column names are eiher the one in AS field or
		// the function name with column within parenthesis, such as
		// sum(height)
		tmpDfColumns := make([]DFColumn, 0)

		// Get columns Dtypes and columns names, without aggregation
		for _, col := range columns {
			tmpDfCol := DFColumn{}

			index := indexOf(col, df.Columns)
			if index == -1 {
				return Dataframe{}, fmt.Errorf("column %s not found in df1", col)
			}

			tmpName := col
			tmpQualifier := ""
			if strings.Contains(col, ".") {
				splitted := strings.Split(col, ".")
				tmpName = splitted[1]
				tmpQualifier = splitted[0]
			}

			tmpDfCol.Info.Name = tmpName
			tmpDfCol.Info.Qualifier = tmpQualifier
			tmpDfCol.DType = df.Columns[index].DType
			tmpDfCol.Serie = NewSeries(nil)
			tmpDfColumns = append(tmpDfColumns, tmpDfCol)
		}

		// Get aggergations infos for new DataFrame
		for _, agg := range aggregators {
			newDfCol := DFColumn{}
			newAggColName := ""

			if agg.As == "" {
				newAggColName = fmt.Sprintf("%s(%s)", agg.Agg, agg.Column)
			} else {
				newAggColName = agg.As
			}

			// Get index in order to get type
			indexOf := indexOf(agg.Column, df.Columns)
			if indexOf == -1 {
				return Dataframe{}, fmt.Errorf("column %s not found in df1", agg.Column)
			}

			newDfCol.Info.Name = newAggColName
			newDfCol.Info.Qualifier = ""
			newDfCol.DType = df.Columns[indexOf].DType
			newDfCol.Serie = NewSeries(nil)

			tmpDfColumns = append(tmpDfColumns, newDfCol)
		}

		finalDf.Columns = tmpDfColumns
	}

	for _, elementsSlice := range hashedTable {
		elementsDf := DataframeFromElementsSlice(elementsSlice, allColumns, DTypes)

		// Do not aggregate, just returns the first row of dataframe
		firstRow, err := elementsDf.Limit(1)

		if err != nil {
			return Dataframe{}, err
		}

		if !aggregate {
			finalDf, err := finalDf.Concat(firstRow)

			if err != nil {
				return Dataframe{}, err
			}

			return finalDf, nil
		}

		// Get columns values
		selectInputs := make([]SelectColumnInput, 0)

		for _, col := range columns {
			selectInputs = append(selectInputs, SelectColumnInput{Column: col, Alias: ""})
		}

		firstRowColumns, err := firstRow.Select(selectInputs)
		if err != nil {
			return Dataframe{}, err
		}

		firstRowValues := firstRowColumns.Values()[0]

		newRow := make([]Element, 0)
		newRow = append(newRow, firstRowValues...)

		// Otherwise, we need to aggregate the values and add the new aggregated columns
		// to the final dataframe

		for i := len(columns); i < len(columns)+len(aggregators); i++ {
			agg := aggregators[i-len(columns)]
			// function
			function, ok := Aggregators[strings.ToLower(agg.Agg)]

			if !ok {
				return Dataframe{}, fmt.Errorf("aggregator %s not found", agg.Agg)
			}

			res, err := function(elementsDf.Columns[i].Serie)

			if err != nil {
				return Dataframe{}, err
			}

			newRow = append(newRow, res.Copy())
		}

		finalDf.AddRow(newRow)

	}
	return finalDf, nil
}

func (df Dataframe) RenameCol(oldName string, newName string) error {
	index := indexOf(oldName, df.Columns)

	if index == -1 {
		return fmt.Errorf("column %s not found", oldName)
	}

	df.Columns[index].Info.Name = newName

	return nil
}

// ByColumn implements sort.Interface for [][]Element based on a specific column.
type ByColumn struct {
	Dataframe *Dataframe
	ColIndex  int
	Ascending bool
}

func (a ByColumn) Len() int {
	return a.Dataframe.Len()
}

func (a ByColumn) Swap(i, j int) {
	tmp := a.Dataframe.getRow(i)
	a.Dataframe.setRow(i, a.Dataframe.getRow(j))
	a.Dataframe.setRow(j, tmp)
}

func (a ByColumn) Less(i, j int) bool {
	if a.Ascending {
		return a.Dataframe.getRow(i)[a.ColIndex].Less(a.Dataframe.getRow(j)[a.ColIndex])
	}
	return a.Dataframe.getRow(i)[a.ColIndex].Greater(a.Dataframe.getRow(j)[a.ColIndex])
}

func (df Dataframe) Sort(columns []string, ascending bool) (Dataframe, error) {
	if len(columns) == 0 {
		return df, fmt.Errorf("no columns specified for sorting")
	}

	// Find the index of the first column to sort by
	for _, column := range columns {
		colIndex := indexOf(column, df.Columns)
		if colIndex == -1 {
			return df, fmt.Errorf("column not found: %s", column)
		}

		// Sort the dataframe by the specified column
		sorter := ByColumn{Dataframe: &df, ColIndex: colIndex, Ascending: ascending}
		sort.Sort(sorter)
	}

	return df, nil
}

func (df Dataframe) Limit(n int) (Dataframe, error) {

	newDf := Dataframe{}

	newDf.Columns = df.Columns
	newDf.Alias = df.Alias

	for i, col := range df.Columns {
		newSerie := col.Serie.Limit(n)

		newDf.Columns[i].Serie = newSerie
		newDf.nRows = newSerie.Len()
	}

	return newDf, nil
}

func (df Dataframe) ToRawRow() ([]db.RawRow, error) {
	rawRows := make([]db.RawRow, 0)

	for i := 0; i < df.Len(); i++ {
		row := make(db.RawRow)
		for _, column := range df.Columns {
			fullColName := column.Info.Name

			if column.Info.Qualifier != "" {
				fullColName = column.Info.Qualifier + "." + column.Info.Name
			}
			serie, err := df.GetColumn(fullColName)

			if err != nil {
				return nil, err
			}

			// Qualifier should not come out in the final result
			row[strings.ToLower(column.Info.Name)] = serie.Elements[i].GetValue()
		}
		rawRows = append(rawRows, row)
	}

	return rawRows, nil
}
