package executor

import (
	db "github.com/nicolasvancan/monvandb/src/database"
	df "github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	system "github.com/nicolasvancan/monvandb/src/system"
)

func RoleCreate(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	name := params[0].(string)
	infos := params[1].(map[string]interface{})

	// Create role
	_, err := system.CreateRole(name, infos)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Role created": name,
	}}), nil
}

func RoleDrop(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	name := params[0].(string)

	// Drop role
	err := system.DeleteRole(name)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Role dropped": name,
	}}), nil
}

func RoleGrant(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	roleName := params[0].(string)
	userName := params[1].(string)

	// Grant role to user
	err := system.AssignRole(roleName, userName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Role granted": roleName,
	}}), nil
}

func RoleRevoke(params ...interface{}) (df.Dataframe, error) {
	// First parameter is always the database name
	// Second parameter is always the context holding execution results
	// The third parameter is always the columns to select

	roleName := params[0].(string)
	userName := params[1].(string)

	// Revoke role from user
	err := system.RevokeRole(roleName, userName)

	if err != nil {
		return df.NewDataframe(nil), err
	}

	return df.NewDataframe([]db.RawRow{{
		"Role revoked": roleName,
	}}), nil
}
