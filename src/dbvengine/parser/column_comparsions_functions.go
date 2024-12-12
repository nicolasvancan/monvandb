package parser

import database "github.com/nicolasvancan/monvandb/src/database"

func createColumnComparsionFullScan(
	ColName string,
	tableName string,
	tableAlias string,
	id int,
	parentId int,
	parentLogicalOp int,
	layerLogicalOp int,
) database.ColumnComparsion {
	return database.ColumnComparsion{
		ColumnName: ColName,
		TableName:  tableName,
		Alias:      tableAlias,
		Condition:  database.EQ,
		Value: database.ColumnConditionValue{
			IsOtherColumn:        true,
			IsOtherTable:         true,
			ColumnName:           ColName,
			TableHash:            tableAlias,
			Value:                nil,
			Transformation:       nil,
			TransformationParams: nil,
		},
		ParentId:        parentId,
		ParentLogicalOp: parentLogicalOp,
		LayerLogicalOp:  layerLogicalOp,
		Id:              id,
	}
}

// This function is used by the function createColumnComparsionFromComprExpr.
// It returns the object and a boolean indicating whether to append it or not, since
// there may be some subqueries involved. When that happens, we don't want to append the object
//
// The verifying proccess to check whether or not the table exists and it is not a subquery it based
// on tableAlias, which is a map that contains the alias of the table as key and the table name as value
func createColumnComparsionFromSingleColumn(
	expr CompExpr,
	tableAlias *map[string]string,
	id int,
	parentId int,
	logicalOp int,
	parenLogicalOp int,
	left bool,
) (database.ColumnComparsion, bool) {
	colName := expr.LeftValue.(string)
	tableName, exists := (*tableAlias)[expr.LeftAlias]
	alias := expr.LeftAlias
	isOtherTable := false
	isOtherColumn := false

	if (expr.RightType == "column" && left) || (expr.LeftType == "column" && !left) {
		isOtherColumn = true
		isOtherTable = true
	}

	if !exists {
		return database.ColumnComparsion{}, false
	}

	if !left {
		colName = expr.RightValue.(string)
		tableName = (*tableAlias)[expr.RightAlias]
		alias = expr.RightAlias
	}

	return database.ColumnComparsion{
		ColumnName: colName,
		TableName:  tableName,
		Alias:      alias,
		Condition:  fromStringToDatabaseCondition(expr.Operator),
		Value: database.ColumnConditionValue{
			IsOtherColumn: isOtherColumn,
			IsOtherTable:  isOtherTable,
			Value:         expr.RightValue,
		},
		ParentId:        parentId,
		ParentLogicalOp: parenLogicalOp,
		LayerLogicalOp:  logicalOp,
		Id:              id,
	}, true
}

// Create ColumnComparsion struct from a ComprExpr returning a slice of ColumnComparsion
// If the expression is a comparsion with a single column, such as: t.a = 1, it will return a slice with a single element
// otherwise it returns both columns comparsions
func createColumnComparsionFromComprExpr(
	expr CompExpr,
	tablesAlias *map[string]string,
	id int,
	parentId int,
	logicalOp int,
	parenLogicalOp int,
) []database.ColumnComparsion {
	cComparsions := make([]database.ColumnComparsion, 0)

	if expr.LeftType == "column" {
		colComp, appendIt := createColumnComparsionFromSingleColumn(expr, tablesAlias, id, parentId, logicalOp, parenLogicalOp, true)

		if appendIt {
			cComparsions = append(cComparsions, colComp)
		}
	}

	if expr.RightType == "column" {
		colComp, appendIt := createColumnComparsionFromSingleColumn(expr, tablesAlias, id, parentId, logicalOp, parenLogicalOp, false)

		if appendIt {
			cComparsions = append(cComparsions, colComp)
		}
	}

	return cComparsions
}

func fromStringToDatabaseCondition(op string) int {
	switch op {
	case "=":
		return database.EQ
	case "<=":
		return database.LTE
	case ">=":
		return database.GTE
	case "<":
		return database.LT
	case ">":
		return database.GT
	case "!=", "<>":
		return database.NE
	case "LIKE":
		return database.LIKE
	case "NOT LIKE":
		return database.NLIKE
	case "IN":
		return database.IN
	case "NOT IN":
		return database.NIN
	}

	return database.EQ
}
