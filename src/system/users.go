package system

import (
	"errors"
	"fmt"

	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/utils"
)

var allUsers = make(map[string]*User)

type User struct {
	Name     string
	Password string
	Roles    []*Role
}

func (u *User) AddRole(role *Role) {
	u.Roles = append(u.Roles, role)
}

func (u *User) RemoveRole(role string) {
	for i, r := range u.Roles {
		if r.Name == role {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			break
		}
	}
}

func UserExists(name string) bool {
	// Check if user exists in the database
	// This is a placeholder implementation
	return false
}

func CreateUser(name string, password string) (*User, error) {
	if UserExists(name) {
		return nil, fmt.Errorf("user %s already exists", name)
	}

	user := &User{
		Name:     name,
		Password: password,
		Roles:    []*Role{},
	}

	// Save user to the database
	// This is a placeholder implementation

	return user, nil
}

func AssignRole(name string, roleName string) error {
	user, err := GetUser(name)

	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user %s does not exist", name)
	}

	role, err := GetRole(roleName)

	if err != nil {
		return errors.New("role does not exist")
	}

	user.AddRole(role)

	// Save changes to the database
	// This is a placeholder implementation

	return nil
}

func RevokeRole(name string, role string) error {
	user := allUsers["users"]

	if user == nil {
		return fmt.Errorf("user %s does not exist", name)
	}

	user.RemoveRole(role)

	// Remove user from the database
	table, err := database.GetTable("system", "users")

	if err != nil {
		return err
	}

	// Remove user from the database
	rolesNames := make([]string, 0)

	for _, r := range user.Roles {
		rolesNames = append(rolesNames, r.Name)
	}

	json, err := utils.ToJson(map[string]interface{}{"roles": rolesNames})

	if err != nil {
		return err
	}

	_, err = table.Update([]database.RawRow{{"id": user.Name, "password": user.Password, "roles": json}})

	if err != nil {
		return err
	}

	return nil
}

func GetUser(name string) (*User, error) {
	// Retrieve user from the database
	// This is a placeholder implementation
	user := allUsers["users"]

	if user == nil {
		return nil, fmt.Errorf("user %s does not exist", name)
	}

	return user, nil
}

func DropUser(name string) error {
	user := allUsers["users"]

	// Check if user exists
	if user == nil {
		return fmt.Errorf("user %s does not exist", name)
	}

	// Remove user from the in-memory map
	delete(allUsers, name)
	// Remove user from the database
	table, err := database.GetTable("system", "users")

	if err != nil {
		return err
	}

	// Remove user from the database
	_, err = table.Delete([]database.RawRow{{"id": user.Name}})

	if err != nil {
		return err
	}

	return nil
}

func LoadAllUsersInMemory() error {
	table, err := database.GetTable("system", "users")

	if err != nil {
		return err
	}

	// scan
	rows := table.Range(nil, -1, -1)

	for _, row := range rows {
		user := &User{
			Name:     row["id"].(string),
			Password: row["password"].(string),
			Roles:    []*Role{},
		}

		var dst map[string]interface{}

		utils.FromJson(row["roles"].([]byte), &dst)

		user.Roles, err = GetRoles(dst["roles"].([]string))

		if err != nil {
			return fmt.Errorf("failed to get roles for user %s: %w", user.Name, err)
		}

		allUsers[user.Name] = user
	}

	return nil
}
