package executor

import (
	db "github.com/nicolasvancan/monvandb/src/database"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
)

func DatabaseCreate(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	databaseName := params[0].(string)

	// Create database
	_, err := db.CreateDatabase(databaseName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Database created": databaseName,
	}}), nil
}

func DatabaseDrop(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	databaseName := params[0].(string)

	// Drop database
	err := db.DropDatabase(databaseName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Database dropped": databaseName,
	}}), nil
}

func IndexCreate(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	databaseName := params[0].(string)
	indexName := params[1].(string)
	tableName := params[2].(string)
	columnName := params[3].(string)
	// Create index
	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	err = database.CreateIndex(tableName, columnName, indexName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	// Return the index created
	return df.NewDataframe([]db.RawRow{{
		"Index created": databaseName,
	}}), nil
}

func IndexDrop(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	databaseName := params[0].(string)
	indexName := params[1].(string)
	tableName := params[2].(string)

	// Drop index
	database, err := db.GetDatabase(databaseName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	err = database.DropIndex(tableName, indexName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Index dropped": databaseName,
	}}), nil
}
