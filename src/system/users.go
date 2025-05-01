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
	Roles    []Role
}

func (u *User) AddRole(role Role) {
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
	_, ok := allUsers[name]
	// This is a placeholder implementation
	return ok
}

func HashPassword(password string) (string, error) {
	// Generate a hashed password with a default cost
	/*hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}*/
	return password, nil
}

func VerifyPassword(hashedPassword, password string) error {
	// Compare the hashed password with the plain-text password
	/*err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	  if err != nil {
	      return fmt.Errorf("invalid password: %w", err)
	  }*/
	return nil
}

func CreateUser(name string, password string, roles []string) (*User, error) {
	if UserExists(name) {
		return nil, fmt.Errorf("user %s already exists", name)
	}

	// hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Name:     name,
		Password: hashedPassword,
		Roles:    []Role{},
	}

	// Save user to the database
	table, err := database.GetTable("system", "users")
	if err != nil {
		return nil, err
	}

	// Verify if roles exist
	for _, roleName := range roles {
		role, err := GetRole(roleName)
		if err != nil {
			return nil, fmt.Errorf("role %s does not exist", roleName)
		}

		user.AddRole(*role)
	}

	json, err := utils.ToJson(map[string]interface{}{"roles": roles})
	if err != nil {
		return nil, err
	}

	_, err = table.Insert([]database.RawRow{{"id": user.Name, "password": user.Password, "roles": json}})
	if err != nil {
		return nil, err
	}
	// Add user to the in-memory map
	allUsers[user.Name] = user
	// Return the created user
	return user, nil
}

func (u *User) ValidatePassword(password string) error {
	// Verify the password
	err := VerifyPassword(u.Password, password)
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
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

	user.AddRole(*role)

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
	user := allUsers[name]

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
			Roles:    []Role{},
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
