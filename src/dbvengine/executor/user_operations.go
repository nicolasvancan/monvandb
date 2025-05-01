package executor

import (
	db "github.com/nicolasvancan/monvandb/src/database"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	system "github.com/nicolasvancan/monvandb/src/system"
)

func UserCreate(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	name := params[0].(string)
	password := params[1].(string)

	// Create user
	_, err := system.CreateUser(name, password, []string{})

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"User created": name,
	}}), nil
}

func UserDrop(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	name := params[0].(string)

	// Drop user
	err := system.DropUser(name)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"User dropped": name,
	}}), nil
}
