package parser

import dataframe "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"

func createOrUpdateFilterBasedOnCompExpr(
	expr CompExpr,
	tableFilters *map[string]dataframe.Filters,
	id int,
	parentId int,
	logicalLayer int,
	parentLogicalLayer int,
	on bool, // If true, it means that the filter is being created based on a ON clause from a JOIN
) {
	// Skip it if both columns are values
	if expr.LeftType == "value" && expr.RightType == "value" {
		return
	}

	// Add filter if not exists
	_, ok := (*tableFilters)[expr.LeftAlias]
	if !ok {
		(*tableFilters)[expr.LeftAlias] = dataframe.NewFilter()
	}

	// Create filter node based on expr
	filterValue := dataframe.FilterValue{}
	filterValue.Comparator = dataframe.ComparatorType(expr.Operator)

	if expr.LeftType == "column" {
		filterValue.Column = dataframe.ColumnFilter{}
		filterValue.Column.Name = expr.LeftValue.(string)
		filterValue.Column.Alias = expr.LeftAlias
		filterValue.Column.Function = ""
		filterValue.Column.Parameters = nil
	} else if expr.LeftType == "function" {
		function := expr.LeftValue.(ColFunction)
		filterValue.Column.Name = function.Column
		filterValue.Column.Function = function.Func
		filterValue.Column.Parameters = function.Args
	}

	if expr.RightType == "column" {
		columnFilter := dataframe.ColumnFilter{}
		columnFilter.Name = expr.RightValue.(string)
		columnFilter.Alias = expr.RightAlias
		columnFilter.Function = ""
		columnFilter.Parameters = nil
		filterValue.Comparando = columnFilter

	} else if expr.RightType == "value" {
		filterValue.Comparando = expr.RightValue
	} else if expr.RightType == "function" {
		function := expr.RightValue.(ColFunction)
		columnFilter := dataframe.ColumnFilter{}
		columnFilter.Name = function.Column
		columnFilter.Alias = function.Alias
		columnFilter.Function = function.Func
		columnFilter.Parameters = function.Args
		filterValue.Comparando = columnFilter
	}

	filterValue.Type = dataframe.FilterNodeType("OR")
	if logicalLayer == 0 {
		filterValue.Type = dataframe.FilterNodeType("AND")
	}

	/* This part represents the map key to the correct table
	When there is a comparsion between two tables, that means that we must do the filtering stage
	after the join stage, not after the from stage

	In other words, all filtering keys that contain "-" rune, will be kept as join filters and will be applied
	after the join stage.

	The variable finalAlias is used to store the final alias that will be used as key in the map
	*/

	finalAlias := expr.LeftAlias
	if !on {
		if expr.LeftAlias != expr.RightAlias {
			finalAlias = expr.LeftAlias + "-" + expr.RightAlias
		}
	}

	_, ok = (*tableFilters)[finalAlias]

	if !ok {
		(*tableFilters)[finalAlias] = dataframe.NewFilter()
	}

	err := (*tableFilters)[finalAlias].InsertValue(id, filterValue)

	if err != nil {
		// Create node if there is no node
		nodeFilterType := dataframe.FilterNodeType("AND")
		if parentLogicalLayer != 0 {
			nodeFilterType = dataframe.FilterNodeType("OR")
		}
		(*tableFilters)[finalAlias].AddNode(parentId, dataframe.FilterNode{
			ID:       id,
			Type:     nodeFilterType,
			Value:    []dataframe.FilterValue{filterValue},
			Children: make([]*dataframe.FilterNode, 0),
		}, nodeFilterType)
	}
}
