package main

import (
	"fmt"

	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/system"
)

func te() {
	db, err := database.CreateDatabase("system")
	if err != nil {
		panic(err)
	}

	db.CreateTable("users", []database.Column{
		{Name: "id", Type: database.COL_TYPE_STRING, Primary: true, Nullable: false},
		{Name: "password", Type: database.COL_TYPE_STRING, Nullable: false},
		{Name: "roles", Type: database.COL_TYPE_BLOB, Nullable: true},
	}, false, false)

	// Create system tables
	db.CreateTable("roles", []database.Column{
		{Name: "id", Type: database.COL_TYPE_STRING, Primary: true, Nullable: false},
		{Name: "admin", Type: database.COL_TYPE_BOOL, Nullable: false},
		{Name: "permissions", Type: database.COL_TYPE_BLOB, Nullable: true},
	}, false, false)

	system.CreateRole("admin", map[string]interface{}{"admin": true})
	system.CreateUser("admin", "password", []string{"admin"})
	system.LoadAllRoles()

	server := system.NewServer(8080)

	server.Run()
}

func main() {
	err := system.LoadAllUsersInMemory()
	if err != nil {
		panic(err)
	}

	a, err := system.GetUser("admin")
	if err != nil {
		panic(err)
	}

	fmt.Println(a)
	//testJson()
}
