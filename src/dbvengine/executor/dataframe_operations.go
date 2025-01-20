package executor

import (
	"github.com/nicolasvancan/monvandb/src/dbvengine/contexts"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
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
	on := params[2].([]string)
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
