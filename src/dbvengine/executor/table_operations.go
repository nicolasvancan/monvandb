package executor

import (
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

	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.Dataframe{}, err
	}

	table, err := database.GetTable(tableName)

	if err != nil {
		return df.Dataframe{}, err
	}

	// Get raw data from table
	return df.NewDataframe(table.Range(columnComparsions, -1, 1)), nil
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
