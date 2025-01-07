package dbvengine

import (
	"fmt"
	"strings"

	"github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	executor "github.com/nicolasvancan/monvandb/src/dbvengine/executor"
	parser "github.com/nicolasvancan/monvandb/src/dbvengine/parser"
)

func BuildPlan(analyzedQuery parser.AnalyzedQueryData, exec *executor.ExecutionLayer) {
	// Parse the query
	switch aq := analyzedQuery.(type) {
	case *parser.AnalyzedQuerySelect:
		// Select query
		resultId := buildSelectPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	}
}

// Function responsible for generating execution plan for select query.
// In general this function may be called recursivelly when sub queries are present
// How it works:
// 1. Creates execution nodes related to data present in the table files
// 2. After that, it creates the filter layer related to the specific table values
// 3. Having all filtered results, we can start to join the dataframes
// 4. Joined tables may also have filters, so we apply filtering into joined tables
// 5. The fifth step is either to select the columns or to apply a group by operation
// 6. The last step is to apply the order by operation
// 7. The final result is returned

func buildSelectPlan(exec *executor.ExecutionLayer, aq *parser.AnalyzedQuerySelect) string {
	// Create the execution nodes for the tables
	nodes := make(map[string]*executor.ExecutionNode)
	returnNode := ""
	previousNode := ""

	if !aq.From.IsSubQuery {
		previousNode = aq.From.Alias
	}

	for alias, table := range aq.TablesAlias {
		// Create new execution Noe
		nodeName := alias
		execNode := executor.NewExecutionNode(nodeName, exec)

		// It comes as -1, if it is different from -1, we must limit the results
		// of the main table
		limit := -1

		if nodeName == aq.From.Alias {
			limit = aq.Limit
		}

		// Create Operation
		operation := executor.Operation{}
		operation.Name = executor.TABLE_FILE_READ
		// Get column comparsions for the table
		columnComparsions := aq.TablesColumnComparsions[alias]
		operation.Args = []interface{}{aq.DatabaseName, table, columnComparsions, limit}

		// Set operation
		execNode.Operation = operation

		// Add node to the map
		nodes[alias] = execNode
	}

	// Sub selects
	for alias, subSelect := range aq.Subqueries {

		// Calls this function recursivelly
		resultNodeName := buildSelectPlan(exec, subSelect)

		// Create reference to node using subquery alias
		nodes[alias] = nodes[resultNodeName]

		if previousNode == "" && aq.From.IsSubQuery && aq.From.Alias == alias {
			previousNode = alias
		}
	}

	// Create the filter layer
	for alias, filter := range aq.TablesFilters {
		// Create new execution Node
		isTableFilter := len(strings.Split(alias, "-")) <= 1

		if isTableFilter {
			nodeName := fmt.Sprintf("%s_filter", alias)
			execNode := executor.NewExecutionNode(nodeName, exec)

			// Create Operation
			operation := executor.Operation{}
			operation.Name = executor.FILTER
			operation.Args = []interface{}{alias, exec.Context, filter}

			// Set operation
			execNode.Operation = operation

			execNode.AddDependency(alias)

			// Add node to the map
			nodes[nodeName] = execNode

			// Edit the parent node to add notify
			nodes[alias].AddNotify(nodeName)
		}
	}

	// Join the dataframes
	// Search for table filtering when there is a join
	// This part turned quite tricky to implement.
	// Let's say we have multiple joins, such as:

	// SELECT * FROM table1 t1
	// JOIN table2 t2 ON t1.id = t2.id AND t1.name = 'John'
	// JOIN table3 t3 ON t2.id = t3.id
	// JOIN table4 t4 ON t3.id = t4.id

	// There must be a join variable that holds joins results between tables
	// whenever a join is made, for example t1-t2, we search for filters related
	// to the joins and apply it to the join result. This is done by adding a filter
	//
	// The process is to complete the join, store it into a temporary variable that
	// holds the join result, and then apply the filter to the join result when
	// the corresponding filter table has been already joinned.

	// this will store the joined tables sources
	// for example, if the resulting join is from t1 to t2, than both will be present
	// in this slice
	existingJoinedSources := make([]string, 0)

	for _, join := range aq.Joins {
		// Create new execution Node

		nodeName := fmt.Sprintf("%s-%s_join", join.LeftAlias, join.RightAlias)

		// Append joins alias to existingJoinedSources
		existingJoinedSources = appendIfMissing(existingJoinedSources, join.LeftAlias)
		existingJoinedSources = appendIfMissing(existingJoinedSources, join.RightAlias)

		execNode := executor.NewExecutionNode(nodeName, exec)

		leftTable := join.LeftAlias
		rightTable := join.RightAlias

		if previousNode != "" {
			if isInSlice(existingJoinedSources, leftTable) {
				leftTable = previousNode
			} else {
				rightTable = previousNode
			}
		}

		// Create Operation
		operation := executor.Operation{}
		operation.Name = executor.JOIN
		operation.Args = []interface{}{[]string{
			leftTable, rightTable},
			exec.Context,
			[]string{join.On.LeftValue.(string)},
			join.How,
		}

		// Set operation
		execNode.Operation = operation

		// Add dependencies
		// When there is no filter for the each table, such as t_filter, we must
		// add the table itself as dependency
		if _, ok := nodes[leftTable+"_filter"]; ok {
			leftTable = leftTable + "_filter"
		}

		if _, ok := nodes[rightTable+"_filter"]; ok {
			rightTable = rightTable + "_filter"
		}
		execNode.AddDependency(leftTable)
		execNode.AddDependency(rightTable)
		// Add node to the map
		nodes[nodeName] = execNode

		// Edit the parent node to add notify
		nodes[leftTable].AddNotify(nodeName)
		nodes[rightTable].AddNotify(nodeName)

		previousNode = nodeName

		// Verify filtering part
		// If there is a filter for the join, we must create a new node
		// to apply the filter to the join result
		var filters dataframe.Filters = nil
		if aq.TablesFilters[join.LeftAlias+"-"+join.RightAlias] != nil {
			filters = aq.TablesFilters[join.LeftAlias+"-"+join.RightAlias]

		} else if aq.TablesFilters[join.RightAlias+"-"+join.LeftAlias] != nil {
			filters = aq.TablesFilters[join.RightAlias+"-"+join.LeftAlias]
		}

		if filters != nil {
			filterNodeName := fmt.Sprintf("%s-%s_join_filter", join.LeftAlias, join.RightAlias)
			execNode = executor.NewExecutionNode(filterNodeName, exec)

			// Create Operation
			operation = executor.Operation{}
			operation.Name = executor.FILTER
			operation.Args = []interface{}{nodeName, exec.Context, filters}

			// Set operation
			execNode.Operation = operation

			// Add dependency
			execNode.AddDependency(nodeName)

			// Add node to the map
			nodes[filterNodeName] = execNode

			// Edit the parent node to add notify
			nodes[nodeName].AddNotify(filterNodeName)

			// Edit the previous join node to add notify
			previousNode = filterNodeName
		}
	}

	// TODO: GroupBy
	groupCols := make([]string, 0)
	groupByAggregators := make([]dataframe.GroupByAgg, 0)
	for _, groupBy := range aq.GroupBy {
		colName := groupBy.Column
		groupCols = append(groupCols, colName)
	}
	// Select based on previous join
	// This is the final node
	selectColumns := make([]string, 0)

	for _, colFunc := range aq.Select {
		// For col functions, when there is either case, function, operation or subquery
		// we must create node handlers for that, otherwise we just append the column name
		if colFunc.Func != "" {
			// When there is a function to be called first we check
			// whether the function is an aggregation function
			if len(groupCols) > 0 {
				_, isInAggregation := dataframe.Aggregators[strings.ToLower(colFunc.Func)]
				if isInAggregation {
					aggregator := dataframe.GroupByAgg{
						Column: colFunc.Column,
						Agg:    strings.ToLower(colFunc.Func),
						As:     colFunc.Alias,
					}

					groupByAggregators = append(groupByAggregators, aggregator)
				}
				continue
			}

			// Otherwise we create a new node for the function
		}

		selectColumns = append(selectColumns, colFunc.Column)
	}

	// If group by is present, we must create a new node for the group by operation
	if len(groupCols) > 0 {
		nodeName := fmt.Sprintf("groupby_%s", previousNode)
		execNode := executor.NewExecutionNode(nodeName, exec)

		// Create Operation
		operation := executor.Operation{}
		operation.Name = executor.GROUPBY
		operation.Args = []interface{}{previousNode, exec.Context, groupCols, groupByAggregators}

		// Set operation
		execNode.Operation = operation

		// Add dependency
		execNode.AddDependency(previousNode)

		// Add node to the map
		nodes[nodeName] = execNode

		// Edit the parent node to add notify
		nodes[previousNode].AddNotify(nodeName)
		previousNode = nodeName
	}

	// Create new execution Node
	nodeName := fmt.Sprintf("select_%s", previousNode)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.SELECT
	operation.Args = []interface{}{previousNode, exec.Context, selectColumns}

	// Set operation
	execNode.Operation = operation

	// Add dependency
	execNode.AddDependency(previousNode)

	// Add node to the map
	nodes[nodeName] = execNode

	// Edit the parent node to add notify
	nodes[previousNode].AddNotify(nodeName)
	previousNode = nodeName
	returnNode = nodeName

	// OrderBy
	if len(aq.Order) > 0 {
		// Create new execution Node
		stringOrderSlice := make([]string, 0)

		for _, order := range aq.Order {
			stringOrderSlice = append(stringOrderSlice, order.Column)
		}

		nodeName := fmt.Sprintf("order_%s", previousNode)
		execNode := executor.NewExecutionNode(nodeName, exec)

		// Create Operation
		operation := executor.Operation{}
		operation.Name = executor.ORDERBY
		operation.Args = []interface{}{previousNode, exec.Context, stringOrderSlice, aq.Asc}

		// Set operation
		execNode.Operation = operation

		// Add dependency
		execNode.AddDependency(previousNode)

		// Add node to the map
		nodes[nodeName] = execNode

		// Edit the parent node to add notify
		nodes[previousNode].AddNotify(nodeName)

		returnNode = nodeName
	}

	// dump the nodes to the execution layer
	for _, node := range nodes {
		exec.AddNode(node, false)
	}
	return returnNode
}

// This should not be here, I must create a module for this kind of functions
func appendIfMissing(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

func isInSlice(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
