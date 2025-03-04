package executor

import (
	"fmt"

	db "github.com/nicolasvancan/monvandb/src/database"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

func TableFileRead(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the table name
	// The last parameter is always an slice of ColumnComparsion

	databaseName := params[0].(string)
	tableName := params[1].(string)

	columnComparsions := params[2].([]db.ColumnComparsion)
	limit := params[3].(int)
	qualifier := params[4].(string)
	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.Dataframe{}, err
	}

	table, err := database.GetTable(tableName)

	if err != nil {
		return df.Dataframe{}, err
	}

	// Get raw data from table
	rangeResults := table.Range(columnComparsions, limit, -1)

	if len(rangeResults) == 0 {
		fmt.Println("No results found for query")
		dataframe := df.Dataframe{}
		colDfs := make([]df.DFColumn, 0)
		for _, col := range table.Columns {

			tmpDFCol := df.DFColumn{
				Info:  df.DFColInfo{Name: col.Name, Qualifier: ""},
				DType: col.Type,
				Serie: df.NewSeries(nil),
			}
			colDfs = append(colDfs, tmpDFCol)
		}

		dataframe.Columns = colDfs
		return dataframe, nil
	}

	newDf := df.NewDataframe(table.Range(columnComparsions, -1, 1))
	newDf.SetQualifier(qualifier)

	return newDf, nil
}

func TableFileInsert(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the table name
	// The last parameter is always a dataframe

	databaseName := params[0].(string)
	tableName := params[1].(string)

	dataframe := params[2].(df.Dataframe)

	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.Dataframe{}, err
	}

	table, err := database.GetTable(tableName)

	if err != nil {
		return df.Dataframe{}, err
	}

	// Insert data into table
	rawRows, err := dataframe.ToRawRow()

	if err != nil {
		return df.Dataframe{}, err
	}

	numberRows, err := table.Insert(rawRows)

	if err != nil || numberRows == 0 {
		return df.Dataframe{}, err
	}

	return df.NewDataframe([]db.RawRow{{"Rows inserted": numberRows}}), nil
}

func TableFileUpdate(params ...interface{}) (df.Dataframe, error) {

	// First parameter is always the database name
	// Second parameter is always the table name
	// The last parameter is always a dataframe

	databaseName := params[0].(string)
	tableName := params[1].(string)

	dataframe := params[2].(df.Dataframe)

	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.Dataframe{}, err
	}

	table, err := database.GetTable(tableName)

	if err != nil {
		return df.Dataframe{}, err
	}

	// Insert data into table
	rawRows, err := dataframe.ToRawRow()

	if err != nil {
		return df.Dataframe{}, err
	}

	numberRows, err := table.Update(rawRows)

	if err != nil || numberRows == 0 {
		return df.Dataframe{}, err
	}

	return df.NewDataframe([]db.RawRow{{"Rows Updated": numberRows}}), nil
}

func TableFileDelete(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the table name
	// The last parameter is always a dataframe

	databaseName := params[0].(string)
	tableName := params[1].(string)

	dataframe := params[2].(df.Dataframe)

	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.Dataframe{}, err
	}

	table, err := database.GetTable(tableName)

	if err != nil {
		return df.Dataframe{}, err
	}

	// Insert data into table
	rawRows, err := dataframe.ToRawRow()

	if err != nil {
		return df.Dataframe{}, err
	}

	numberRows, err := table.Delete(rawRows)

	if err != nil || numberRows == 0 {
		return df.Dataframe{}, err
	}

	return df.NewDataframe([]db.RawRow{{"Rows Deleted": numberRows}}), nil
}

func TableFileCreate(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the table name
	// The last parameter is always a []db.Column
	// fourth param is truncate
	// fith param is if exists

	databaseName := params[0].(string)
	tableName := params[1].(string)
	columns := params[2].([]db.Column)
	truncate := params[3].(bool)
	verifyExistence := params[4].(bool)

	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.Dataframe{}, err
	}

	// Create a new table
	err = database.CreateTable(tableName, columns, truncate, verifyExistence)

	if err != nil {
		return df.Dataframe{}, err
	}

	return df.NewDataframe([]db.RawRow{{"Table Created": tableName}}), nil
}
