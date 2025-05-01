package system

import db "github.com/nicolasvancan/monvandb/src/database"

func createSystemDatabase() *db.Database {
	// Create system database
	database, _ := db.CreateDatabase("system")
	// Create system tables
	database.CreateTable("users", []db.Column{
		{Name: "id", Type: db.COL_TYPE_STRING, Primary: true, Nullable: false},
		{Name: "password", Type: db.COL_TYPE_STRING, Nullable: false},
		{Name: "roles", Type: db.COL_TYPE_BLOB, Nullable: true},
	}, false, false)

	// Create system tables
	database.CreateTable("roles", []db.Column{
		{Name: "id", Type: db.COL_TYPE_STRING, Primary: true, Nullable: false},
		{Name: "admin", Type: db.COL_TYPE_BOOL, Nullable: false},
		{Name: "permissions", Type: db.COL_TYPE_BLOB, Nullable: true},
	}, false, false)

	return database
}
