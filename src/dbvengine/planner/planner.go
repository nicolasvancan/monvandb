package planner

import (
	"fmt"
	"strings"

	"github.com/nicolasvancan/monvandb/src/database"
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
	case *parser.AnalyzedQueryCreateTable:
		resultId := buildCreateTablePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryInsert:
		resultId := buildInsertRowsTablePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryUpdate:
		resultId := buildUpdateTablePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryDelete:
		resultId := buildDeleteTablePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryDropTable:
		resultId := buildDropTablePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryTruncateTable:
		resultId := buildTruncateTablePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryCreateDatabase:
		resultId := buildCreateDatabasePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryDropDatabase:
		resultId := buildDropDatabasePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryCreateIndex:
		resultId := buildCreateIndexPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryDropIndex:
		resultId := buildDropIndexPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryCreateUser:
		resultId := buildCreateUserPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryDropUser:
		resultId := buildDropUserPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryAssignRole:
		resultId := buildGrantUserPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryRevokeRole:
		resultId := buildRevokeUserPlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryCreateRole:
		resultId := buildCreateRolePlan(exec, aq)
		exec.Nodes[resultId].Final = true
	case *parser.AnalyzedQueryDropRole:
		resultId := buildDropRolePlan(exec, aq)
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

	// Build queries for table
	previousNode = buildQueryForTables(exec, aq, &nodes, previousNode)

	// Sub selects
	previousNode = buildQueryForSubSelects(exec, aq, &nodes, previousNode)

	// Create the filter layer
	previousNode = buildQueriesForFilterLayer(exec, &aq.TablesFilters, &nodes, previousNode)

	// Create the join layer
	previousNode = buildQueriesForJoins(exec, &aq.TablesFilters, &aq.Joins, &nodes, previousNode)

	// Create queries for GroupBy
	groupCols := make([]string, 0)
	groupByAggregators := make([]dataframe.GroupByAgg, 0)
	for _, groupBy := range aq.GroupBy {
		colName := groupBy.Column
		groupCols = append(groupCols, colName)
	}

	// This is the final node
	selectColumns := make([]dataframe.SelectColumnInput, 0)

	// Select based on previous join
	previousNode = fillUpAggregatorsAndSelect(
		exec,
		aq,
		&groupByAggregators,
		&selectColumns,
		previousNode,
		len(groupCols) > 0,
	)

	// Create operations for select and group by
	previousNode = selectAndGroupBy(
		exec,
		&nodes,
		previousNode,
		groupCols,
		groupByAggregators,
		selectColumns,
	)

	// OrderBy
	previousNode = buildOrderByPlan(
		exec,
		aq,
		&nodes,
		previousNode,
	)

	returnNode = previousNode

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

func buildQueryForTables(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQuerySelect,
	nodes *map[string]*executor.ExecutionNode,
	previousNode string,
) string {
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
		operation.Args = []interface{}{
			aq.DatabaseName,
			table,
			columnComparsions,
			limit,
			alias, // This will be the qualifier
		}

		// Set operation
		execNode.Operation = operation

		// Add node to the map
		(*nodes)[alias] = execNode
	}

	return previousNode
}

func buildQueryForSubSelects(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQuerySelect,
	nodes *map[string]*executor.ExecutionNode,
	previousNode string,
) string {
	prevNode := previousNode
	for alias, subSelect := range aq.Subqueries {

		// Calls this function recursivelly
		resultNodeName := buildSelectPlan(exec, subSelect)

		// Create reference to node using subquery alias
		(*nodes)[alias] = (*nodes)[resultNodeName]

		if previousNode == "" && aq.From.IsSubQuery && aq.From.Alias == alias {
			prevNode = alias
		}
	}

	return prevNode
}

func buildQueriesForFilterLayer(
	exec *executor.ExecutionLayer,
	TablesFilters *map[string]dataframe.Filters,
	nodes *map[string]*executor.ExecutionNode,
	previousNode string,
) string {
	for alias, filter := range *TablesFilters {
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
			(*nodes)[nodeName] = execNode

			// Edit the parent node to add notify
			(*nodes)[alias].AddNotify(nodeName)
		}
	}

	return previousNode
}

func buildQueriesForJoins(
	exec *executor.ExecutionLayer,
	TablesFilters *map[string]dataframe.Filters,
	Joins *map[string]parser.JoinAnalysis,
	nodes *map[string]*executor.ExecutionNode,
	previousNode string,
) string {

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
	prevNode := previousNode
	for _, join := range *Joins {
		// Create new execution Node

		nodeName := fmt.Sprintf("%s-%s_join", join.LeftAlias, join.RightAlias)

		// Append joins alias to existingJoinedSources
		existingJoinedSources = appendIfMissing(existingJoinedSources, join.LeftAlias)
		existingJoinedSources = appendIfMissing(existingJoinedSources, join.RightAlias)

		execNode := executor.NewExecutionNode(nodeName, exec)

		leftTable := join.LeftAlias
		rightTable := join.RightAlias

		if prevNode != "" {
			if isInSlice(existingJoinedSources, leftTable) {
				leftTable = prevNode
			} else {
				rightTable = prevNode
			}
		}

		dfJoin := dataframe.JoinOn{
			Left:       join.On.LeftValue.(string),
			Right:      join.On.RightValue.(string),
			Comparator: join.On.Operator,
		}
		// Create Operation
		operation := executor.Operation{}
		operation.Name = executor.JOIN
		operation.Args = []interface{}{[]string{
			leftTable, rightTable},
			exec.Context,
			dfJoin,
			join.How,
		}

		// Set operation
		execNode.Operation = operation

		// Add dependencies
		// When there is no filter for the each table, such as t_filter, we must
		// add the table itself as dependency
		if _, ok := (*nodes)[leftTable+"_filter"]; ok {
			leftTable = leftTable + "_filter"
		}

		if _, ok := (*nodes)[rightTable+"_filter"]; ok {
			rightTable = rightTable + "_filter"
		}
		execNode.AddDependency(leftTable)
		execNode.AddDependency(rightTable)
		// Add node to the map
		(*nodes)[nodeName] = execNode

		// Edit the parent node to add notify
		(*nodes)[leftTable].AddNotify(nodeName)
		(*nodes)[rightTable].AddNotify(nodeName)

		prevNode = nodeName

		// Verify filtering part
		// If there is a filter for the join, we must create a new node
		// to apply the filter to the join result
		var filters dataframe.Filters = nil
		if (*TablesFilters)[join.LeftAlias+"-"+join.RightAlias] != nil {
			filters = (*TablesFilters)[join.LeftAlias+"-"+join.RightAlias]

		} else if (*TablesFilters)[join.RightAlias+"-"+join.LeftAlias] != nil {
			filters = (*TablesFilters)[join.RightAlias+"-"+join.LeftAlias]
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
			(*nodes)[filterNodeName] = execNode

			// Edit the parent node to add notify
			(*nodes)[nodeName].AddNotify(filterNodeName)

			// Edit the previous join node to add notify
			prevNode = filterNodeName
		}
	}

	return prevNode
}

func fillUpAggregatorsAndSelect(
	_ *executor.ExecutionLayer,
	aq *parser.AnalyzedQuerySelect,
	groupByAggregators *[]dataframe.GroupByAgg,
	selectColumns *[]dataframe.SelectColumnInput,
	previousNode string,
	hasGroupBy bool,
) string {
	for _, colFunc := range aq.Select {
		// For col functions, when there is either case, function, operation or subquery
		// we must create node handlers for that, otherwise we just append the column name
		if colFunc.Func != "" {
			// When there is a function to be called first we check
			// whether the function is an aggregation function
			if hasGroupBy {
				_, isInAggregation := dataframe.Aggregators[strings.ToLower(colFunc.Func)]
				if isInAggregation {
					aggregator := dataframe.GroupByAgg{
						Column: colFunc.Column,
						Agg:    strings.ToLower(colFunc.Func),
						As:     colFunc.Alias,
					}

					*groupByAggregators = append(*groupByAggregators, aggregator)
				}
				continue
			}

			// Otherwise we create a new node for the function
		}
		tmpSelectColumn := dataframe.SelectColumnInput{
			Column: colFunc.Column,
			Alias:  colFunc.Alias,
		}

		*selectColumns = append(*selectColumns, tmpSelectColumn)
	}

	return previousNode
}

func selectAndGroupBy(
	exec *executor.ExecutionLayer,
	nodes *map[string]*executor.ExecutionNode,
	previousNode string,
	groupCols []string,
	groupByAggregators []dataframe.GroupByAgg,
	selectColumns []dataframe.SelectColumnInput,
) string {
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
		(*nodes)[nodeName] = execNode

		// Edit the parent node to add notify
		(*nodes)[previousNode].AddNotify(nodeName)
		previousNode = nodeName
	} else {
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
		(*nodes)[nodeName] = execNode

		// Edit the parent node to add notify
		(*nodes)[previousNode].AddNotify(nodeName)
		previousNode = nodeName
	}

	return previousNode
}

func buildOrderByPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQuerySelect,
	nodes *map[string]*executor.ExecutionNode,
	previousNode string,
) string {
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
		(*nodes)[nodeName] = execNode

		// Edit the parent node to add notify
		(*nodes)[previousNode].AddNotify(nodeName)

		previousNode = nodeName
	}

	return previousNode
}

func buildInsertRowsTablePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryInsert,
) string {
	// Create new execution Node
	previousNode := ""
	// Verify if values are a subquery or static values
	switch aq.Values.(type) {
	case parser.AnalyzedQuerySelect:
		// Subquery
		subSelect := aq.Values.(parser.AnalyzedQuerySelect)
		subSelectNodeName := buildSelectPlan(exec, &subSelect)
		previousNode = subSelectNodeName
	case []database.RawRow:
		// Static values
		name := fmt.Sprintf("%s-%s", "DataframeCreate", aq.TableName)
		newExecNode := executor.NewExecutionNode(name, exec)

		operation := executor.Operation{}
		operation.Name = executor.DATAFRAME_CREATE
		operation.Args = []interface{}{aq.Values}

		// Set operation
		newExecNode.Operation = operation

		// Add node to the map
		exec.AddNode(newExecNode, false)

		previousNode = name
	}

	nodeName := fmt.Sprintf("insert_rows_%s", aq.TableName)
	execNode := executor.NewExecutionNode(nodeName, exec)
	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.TABLE_FILE_INSERT
	operation.Args = []interface{}{
		previousNode,
		exec.Context,
		aq.DatabaseName,
		aq.TableName,
	}

	// Set operation
	execNode.Operation = operation
	execNode.AddDependency(previousNode)
	exec.Nodes[previousNode].AddNotify(nodeName)

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildUpdateTablePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryUpdate,
) string {
	// Create new execution Node
	nodes := make(map[string]*executor.ExecutionNode)
	previousNode := ""
	tableName := aq.TableName.Table
	if aq.TableName.Alias != "" {
		tableName = aq.TableName.Alias
	}

	execNode := executor.NewExecutionNode(tableName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.TABLE_FILE_READ
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName.Table,
		aq.TablesColumnComparsions[tableName],
		-1,
		aq.TableName.Alias,
	}

	// Add to execution node
	execNode.Operation = operation
	nodes[tableName] = execNode
	previousNode = tableName

	// Sub selects
	for alias, subSelect := range aq.Subqueries {

		// Calls this function recursivelly
		resultNodeName := buildSelectPlan(exec, subSelect)

		// Create reference to node using subquery alias
		nodes[alias] = nodes[resultNodeName]

		nodes[previousNode].AddNotify(alias)
		previousNode = alias
	}

	// Create the filter layer
	previousNode = buildQueriesForFilterLayer(exec, &aq.TablesFilters, &nodes, previousNode)

	// Create the join layer
	if len(aq.Joins) > 0 {
		previousNode = buildQueriesForJoins(exec, &aq.TablesFilters, &aq.Joins, &nodes, previousNode)
	} else {
		// update previous node to be the filtered table
		previousNode = fmt.Sprintf("%s_filter", previousNode)
	}

	// For each set, we must create a new node
	for _, set := range aq.Set {
		nodeName := fmt.Sprintf("update_set_%s_%s", aq.TableName.Table, set.Column)
		execNode = executor.NewExecutionNode(nodeName, exec)

		// Create Operation
		operation = executor.Operation{}
		operation.Name = executor.SET_ROW
		operation.Args = []interface{}{
			previousNode,
			exec.Context,
			set.Column,
			set.Value,
		}

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

	// Select only base table fields
	nodeName := fmt.Sprintf("select_%s", aq.TableName.Table)
	execNode = executor.NewExecutionNode(nodeName, exec)

	// Get all base table columns
	columns := make([]dataframe.SelectColumnInput, 0)

	tableToBeUpdated := aq.TableName

	// Get table
	db, _ := database.GetDatabase(aq.DatabaseName)

	table, _ := db.GetTable(tableToBeUpdated.Table)

	for _, column := range table.Columns {
		colName := strings.ToLower(column.Name)
		if tableToBeUpdated.Alias != "" {
			colName = fmt.Sprintf("%s.%s", tableToBeUpdated.Alias, colName)
		}

		columns = append(columns, dataframe.SelectColumnInput{
			Column: colName,
			Alias:  column.Name,
		})
	}
	// Create Operation
	operation = executor.Operation{}
	operation.Name = executor.SELECT
	operation.Args = []interface{}{previousNode, exec.Context, columns}

	// Set operation
	execNode.Operation = operation

	// Add dependency

	execNode.AddDependency(previousNode)

	// Add node to the map
	nodes[nodeName] = execNode

	// Edit the parent node to add notify
	nodes[previousNode].AddNotify(nodeName)
	previousNode = nodeName

	// Update the table
	nodeName = fmt.Sprintf("update_table_%s", aq.TableName.Table)
	execNode = executor.NewExecutionNode(nodeName, exec)
	execNode.AddDependency(previousNode)
	// Create Operation
	operation = executor.Operation{}
	operation.Name = executor.TABLE_FILE_UPDATE
	operation.Args = []interface{}{
		previousNode,
		exec.Context,
		aq.DatabaseName,
		aq.TableName.Table,
	}

	execNode.Operation = operation

	nodes[previousNode].AddNotify(nodeName)
	nodes[nodeName] = execNode

	// Add notify to parent
	previousNode = nodeName

	// Add all nodes to the execution layer
	for _, node := range nodes {
		exec.AddNode(node, false)
	}

	return previousNode
}

func buildDeleteTablePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryDelete,
) string {
	// Create new execution Node
	nodes := make(map[string]*executor.ExecutionNode)
	previousNode := ""
	tableName := aq.TableName.Table
	if aq.TableName.Alias != "" {
		tableName = aq.TableName.Alias
	}

	execNode := executor.NewExecutionNode(tableName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.TABLE_FILE_READ
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName.Table,
		aq.TablesColumnComparsions[tableName],
		-1,
		aq.TableName.Alias,
	}

	// Add to execution node
	execNode.Operation = operation
	nodes[tableName] = execNode
	previousNode = tableName

	// Sub selects
	for alias, subSelect := range aq.Subqueries {

		// Calls this function recursivelly
		resultNodeName := buildSelectPlan(exec, subSelect)

		// Create reference to node using subquery alias
		nodes[alias] = nodes[resultNodeName]

		nodes[previousNode].AddNotify(alias)
		previousNode = alias
	}

	// Create the filter layer
	previousNode = buildQueriesForFilterLayer(exec, &aq.TablesFilters, &nodes, previousNode)

	// Create the join layer
	if len(aq.Joins) > 0 {
		previousNode = buildQueriesForJoins(exec, &aq.TablesFilters, &aq.Joins, &nodes, previousNode)
	} else {
		// update previous node to be the filtered table
		previousNode = fmt.Sprintf("%s_filter", previousNode)
	}

	// Select only base table fields
	nodeName := fmt.Sprintf("select_%s", aq.TableName.Table)
	execNode = executor.NewExecutionNode(nodeName, exec)

	// Get all base table columns
	columns := make([]dataframe.SelectColumnInput, 0)

	tableToBeUpdated := aq.TableName

	// Get table
	db, _ := database.GetDatabase(aq.DatabaseName)

	table, _ := db.GetTable(tableToBeUpdated.Table)

	for _, column := range table.Columns {
		colName := strings.ToLower(column.Name)
		if tableToBeUpdated.Alias != "" {
			colName = fmt.Sprintf("%s.%s", tableToBeUpdated.Alias, colName)
		}

		if column.Primary {
			columns = append(columns, dataframe.SelectColumnInput{
				Column: colName,
				Alias:  column.Name,
			})
		}
	}
	// Create Operation
	operation = executor.Operation{}
	operation.Name = executor.SELECT
	operation.Args = []interface{}{previousNode, exec.Context, columns}

	// Set operation
	execNode.Operation = operation

	// Add dependency

	execNode.AddDependency(previousNode)

	// Add node to the map
	nodes[nodeName] = execNode

	// Edit the parent node to add notify
	nodes[previousNode].AddNotify(nodeName)
	previousNode = nodeName

	// Create delete operation

	nodeName = fmt.Sprintf("delete_table_%s", aq.TableName.Table)
	execNode = executor.NewExecutionNode(nodeName, exec)
	execNode.AddDependency(previousNode)
	// Create Operation
	operation = executor.Operation{}
	operation.Name = executor.TABLE_FILE_DELETE
	operation.Args = []interface{}{
		previousNode,
		exec.Context,
		aq.DatabaseName,
		aq.TableName.Table,
	}

	execNode.Operation = operation

	nodes[previousNode].AddNotify(nodeName)
	nodes[nodeName] = execNode

	// Add notify to parent
	previousNode = nodeName

	// Insert all nodes
	for _, node := range nodes {
		exec.AddNode(node, false)
	}

	return previousNode
}

/* Table planners */

func buildCreateTablePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryCreateTable,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("create_table_%s", aq.TableName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.TABLE_CREATE
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName,
		aq.Columns,
		aq.Truncate,
		aq.VerifyExistence,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildDropTablePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryDropTable,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("drop_table_%s", aq.TableName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.TABLE_DROP
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildTruncateTablePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryTruncateTable,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("truncate_table_%s", aq.TableName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.TABLE_TRUNCATE
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName,
		make([]database.Column, 0),
		true,
		false,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildCreateDatabasePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryCreateDatabase,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("create_database_%s", aq.DatabaseName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.DATABASE_CREATE
	operation.Args = []interface{}{
		aq.DatabaseName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildDropDatabasePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryDropDatabase,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("drop_database_%s", aq.DatabaseName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.DATABASE_DROP
	operation.Args = []interface{}{
		aq.DatabaseName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildCreateIndexPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryCreateIndex,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("create_index_%s", aq.IndexName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.INDEX_CREATE
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName,
		aq.IndexName,
		aq.Columns,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildDropIndexPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryDropIndex,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("drop_index_%s", aq.IndexName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.INDEX_DROP
	operation.Args = []interface{}{
		aq.DatabaseName,
		aq.TableName,
		aq.IndexName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildCreateUserPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryCreateUser,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("create_user_%s", aq.User)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.USER_CREATE
	operation.Args = []interface{}{
		aq.User,
		aq.Password,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildDropUserPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryDropUser,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("drop_user_%s", aq.Username)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.USER_DROP
	operation.Args = []interface{}{
		aq.Username,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildGrantUserPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryAssignRole,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("grant_user_%s", aq.UserName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.ROLE_ASSIGN
	operation.Args = []interface{}{
		aq.RoleName,
		aq.UserName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildRevokeUserPlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryRevokeRole,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("revoke_user_%s", aq.UserName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.ROLE_REVOKE
	operation.Args = []interface{}{
		aq.RoleName,
		aq.UserName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildCreateRolePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryCreateRole,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("create_role_%s", aq.RoleName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.ROLE_CREATE
	operation.Args = []interface{}{
		aq.RoleName,
		aq.Options,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}

func buildDropRolePlan(
	exec *executor.ExecutionLayer,
	aq *parser.AnalyzedQueryDropRole,
) string {
	// Create new execution Node
	nodeName := fmt.Sprintf("drop_role_%s", aq.RoleName)
	execNode := executor.NewExecutionNode(nodeName, exec)

	// Create Operation
	operation := executor.Operation{}
	operation.Name = executor.ROLE_DROP
	operation.Args = []interface{}{
		aq.RoleName,
	}

	// Set operation
	execNode.Operation = operation

	// Add node to the map
	exec.AddNode(execNode, true)

	return nodeName
}
