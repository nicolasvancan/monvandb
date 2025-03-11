package executor

import (
	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/dbvengine/contexts"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser"
)

func DataframeSelect(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the reference name of the dataframe
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	dataframeRef := params[0].(string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRef]
	selectCols := params[2].([]df.SelectColumnInput)

	// Select data from dataframe
	selectedDf, err := dataframe.Select(selectCols)

	// Rename columns that have alias
	for _, col := range selectCols {
		if col.Alias != "" {
			selectedDf.RenameCol(col.Column, col.Alias)
		}
	}

	if err != nil {
		return df.Dataframe{}, err
	}

	return selectedDf, nil
}

func DataframeJoin(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the reference name of the dataframes
	// as a slice
	// Second parameter is always the join dataframe
	// The third parameter is always the join condition

	dataframeRefs := params[0].([]string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRefs[0]]
	joinDataframe := context.OperationsResults[dataframeRefs[1]]
	// On
	on := params[2].(df.JoinOn)
	// how
	how := params[3].(string)

	// Join dataframes
	joinedDf, err := dataframe.Join(joinDataframe, on, how)

	if err != nil {
		return df.Dataframe{}, err
	}

	return joinedDf, nil
}

func DataframeFilter(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the reference name of the dataframe
	// The last parameter is always the filters

	dataframeRef := params[0].(string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRef]
	filters := params[2].(df.Filters)

	// Filter data from dataframe
	filteredDf, err := dataframe.Filter(filters)

	if err != nil {
		return df.Dataframe{}, err
	}

	return filteredDf, nil
}

func DataframeGroupBy(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the reference name of the dataframe
	// Second parameter is always the group by columns
	// The last parameter is always the aggregate functions

	dataframeRef := params[0].(string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRef]
	groupByCols := params[2].([]string)
	aggregateFns := params[3].([]df.GroupByAgg)

	// Group by dataframe
	groupedDf, err := dataframe.GroupBy(groupByCols, aggregateFns)

	if err != nil {
		return df.Dataframe{}, err
	}

	return groupedDf, nil
}

func DataframeOrderBy(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the reference name of the dataframe
	// The last parameter is always the order by columns

	dataframeRef := params[0].(string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRef]
	cols := params[2].([]string)
	orderByCols := params[3].(string) // ASC or DESC

	// Order by dataframe
	orderedDf, err := dataframe.Sort(cols, orderByCols == "asc")

	if err != nil {
		return df.Dataframe{}, err
	}

	return orderedDf, nil
}

func DataframeLimit(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the dataframe
	// The last parameter is always the limit

	dataframeRef := params[0].(string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRef]
	limit := params[2].(int)

	// Limit dataframe
	limitedDf, err := dataframe.Limit(limit)

	if err != nil {
		return df.Dataframe{}, err
	}

	return limitedDf, nil
}

func CreateDataFrameFromRawRow(params ...interface{}) (df.Dataframe, error) {

	// First parameter is always the raw rows
	// Second parameter is always the columns

	rawRows := params[0].([]database.RawRow)

	// Create dataframe from raw rows
	dataframe := df.NewDataframe(rawRows)

	return dataframe, nil
}

func DataframeSetRow(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the reference name of the dataframe
	// The last parameter is always the row to set

	dataframeRef := params[0].(string)
	context := params[1].(*contexts.ExecutionContext)
	dataframe := context.OperationsResults[dataframeRef]
	colToSet := params[2].(string)
	value := params[3]

	var valToSet interface{}
	switch v := value.(type) {
	case parser.ColFunction:
		// Evaluate column function
		valToSet = v.Column
		if v.Alias != "" {
			valToSet = v.Alias + "." + v.Column
		}
	default:
		valToSet = value
	}

	// Set row in dataframe
	dataframe.SetColumn(colToSet, valToSet)

	return dataframe, nil
}
