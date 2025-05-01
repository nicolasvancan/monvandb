package system

import (
	"fmt"
	"os"
	"testing"
)

func TestRoleGet(t *testing.T) {
	// Create system database
	os.Setenv("MONVANDB_PATH", t.TempDir())
	createSystemDatabase()
	// Create system roles
	CreateRole("admin", map[string]interface{}{
		"admin":       true,
		"permissions": []string{},
	})

	role, err := GetRole("admin")
	if err != nil {
		t.Errorf("Role not found")
	}

	if role.Name != "admin" {
		t.Errorf("Role name not found")
	}
	if role.IsAdmin != true {
		t.Errorf("Role admin not found")
	}

}

func TestUserGet(t *testing.T) {
	// Create system database
	os.Setenv("MONVANDB_PATH", t.TempDir())
	createSystemDatabase()
	// Create system users
	CreateUser("admin", "admin", []string{"admin"})

	user, err := GetUser("admin")

	if err != nil {
		fmt.Println(err)
		t.Errorf("User not found")
		return
	}

	if user.Name != "admin" {
		t.Errorf("User name not found")
	}
	if user.Password != "admin" {
		t.Errorf("User password not found")
	}
}

func TestUserCanView(t *testing.T) {
	// Create system database
	os.Setenv("MONVANDB_PATH", t.TempDir())
	createSystemDatabase()
	// Create system users

	CreateRole("admin", map[string]interface{}{
		"admin":       true,
		"permissions": []string{},
	})

	CreateUser("admin", "admin", []string{"admin"})

	user, err := GetUser("admin")
	LoadAllRoles()
	if err != nil {
		fmt.Println(err)
		t.Errorf("User not found")
		return
	}

	if CanView(user.Roles, DatabaseIdentType, "system") != true {
		t.Errorf("User can not view system database")
	}
}
