package system

import (
	"errors"
	"strconv"
	"strings"

	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/utils"
)

type IdentifierType string

// SystemIdentifierType is the type of the system identifier
// It can be either USER or ROLE

var allRoles map[string]*Role = make(map[string]*Role)

const (
	TableIdentType     IdentifierType = "TABLE"
	IndexIdentType     IdentifierType = "INDEX"
	DatabaseIdentType  IdentifierType = "DATABASE"
	UserIdentifierType IdentifierType = "USER"
	RoleIdentifierType IdentifierType = "ROLE"
)

type IdentifierOpetionsMapKeys string

// Possible map keys
const (
	DatabaseKey IdentifierOpetionsMapKeys = "database"
	TableKey    IdentifierOpetionsMapKeys = "table"
)

type Role struct {
	Name        string
	IsAdmin     bool
	Permissions []Permissions
}

type Permissions struct {
	Read     bool // Read permission
	Write    bool // Write permission - This
	View     bool // View permission - Used to view items from a database object
	Drop     bool // Drop permission - Used to drop items from a database object
	Alter    bool // Alter permission - Used to alter items from a database object
	Elements []Identifier
}

type Identifier struct {
	Name    string
	All     bool // All permission - Used to grant all permissions
	Type    IdentifierType
	Options map[string]string // In case it is a index or a table that need to be identified by database
}

func NewRole(name string) *Role {
	return &Role{
		Name:        name,
		IsAdmin:     false,
		Permissions: make([]Permissions, 0),
	}
}

func CanRead(roles []Role, identifierType IdentifierType, identifierName string, options ...string) bool {
	for _, role := range roles {
		if role.IsAdmin {
			return true
		}

		for _, permission := range role.Permissions {
			if !permission.Read {
				continue
			}

			return validateElement(permission.Elements, identifierType, identifierName, options...)
		}
	}

	return false
}

func CanWrite(roles []Role, identifierType IdentifierType, identifierName string, options ...string) bool {
	for _, role := range roles {
		if role.IsAdmin {
			return true
		}

		for _, permission := range role.Permissions {
			if !permission.Write {
				continue
			}

			return validateElement(permission.Elements, identifierType, identifierName, options...)
		}
	}

	return false
}

func CanView(roles []Role, identifierType IdentifierType, identifierName string, options ...string) bool {
	for _, role := range roles {
		if role.IsAdmin {
			return true
		}

		for _, permission := range role.Permissions {
			if !permission.View {
				continue
			}

			return validateElement(permission.Elements, identifierType, identifierName, options...)
		}
	}

	return false
}

func CanDrop(roles []Role, identifierType IdentifierType, identifierName string, options ...string) bool {
	for _, role := range roles {
		if role.IsAdmin {
			return true
		}

		for _, permission := range role.Permissions {
			if !permission.Drop {
				continue
			}

			return validateElement(permission.Elements, identifierType, identifierName, options...)
		}
	}

	return false
}

func CanAlter(roles []Role, identifierType IdentifierType, identifierName string, options ...string) bool {
	for _, role := range roles {
		if role.IsAdmin {
			return true
		}

		for _, permission := range role.Permissions {
			if !permission.Alter {
				continue
			}

			return validateElement(permission.Elements, identifierType, identifierName, options...)
		}
	}

	return false
}

func validateElement(
	elements []Identifier,
	identifierType IdentifierType,
	identifierName string,
	options ...string,
) bool {
	for _, element := range elements {

		if strings.EqualFold(element.Name, identifierName) && element.Type == identifierType {
			if len(options) > 0 {
				// cases for table and index
				roleDatabase := options[0] // Database

				if value, ok := element.Options["database"]; ok && value == roleDatabase || element.All {
					if identifierType == IndexIdentType {
						if value, ok := element.Options["table"]; ok && value == options[1] {
							return true
						}
					}
					return true
				}

				if value, ok := element.Options[options[0]]; ok && value == options[1] || element.All {
					return true
				}
			} else {
				return true
			}
		}
	}
	return false
}

func CreateRole(name string, roleJson map[string]interface{}) (*Role, error) {
	// Check if role already exists
	// This is a placeholder implementation

	/*
		Example of json for all tables from database db1
		{
			"admin": false,
			"elements": [
				{
					"permissions": ["read", "write", "view", "drop", "alter"],
					"elements": [
						{
							"type": "table",
							"name": "*",
							"database": "db1"
						}
					]
				},{
					"permissions": ["read", "view"],
					"elements": [
						{
							"type":"database",
							"name": "db1"
						}
					]
				}
		]
		}
	*/
	newRole := NewRole(name)
	newRole.IsAdmin = false

	if admin, ok := roleJson["admin"]; ok {

		switch a := admin.(type) {
		case bool:
			newRole.IsAdmin = a
		default:
			return nil, errors.New("admin must be a boolean")
		}
	}

	if !newRole.IsAdmin {

		if _, ok := roleJson["elements"]; !ok &&
			roleJson["elements"] == nil || roleJson["elements"] == false {
			return nil, errors.New("elements are required")
		}

		if len(roleJson["elements"].([]interface{})) == 0 {
			return nil, errors.New("elements are required")
		}

		for i, element := range roleJson["elements"].([]interface{}) {
			permissions := Permissions{}

			// Validate fields
			if _, ok := element.(map[string]interface{})["permissions"]; !ok {
				return nil, errors.New("permissions are required in the element " + strconv.Itoa(i))
			}

			if _, ok := element.(map[string]interface{})["elements"]; !ok {
				return nil, errors.New("elements are required in the element " + strconv.Itoa(i))
			}

			perm := element.(map[string]interface{})["permissions"].([]interface{})
			if len(perm) == 0 {
				return nil, errors.New("permissions are required in the element " + strconv.Itoa(i))
			}

			elems := element.(map[string]interface{})["elements"].([]interface{})
			if len(elems) == 0 {
				return nil, errors.New("elements are required in the element " + strconv.Itoa(i))
			}

			for _, p := range perm {
				switch p.(string) {
				case "read":
					permissions.Read = true
				case "write":
					permissions.Write = true
				case "view":
					permissions.View = true
				case "drop":
					permissions.Drop = true
				case "alter":
					permissions.Alter = true
				default:
					return nil, errors.New("invalid permission in the element " + strconv.Itoa(i))
				}
			}
			for _, e := range elems {
				identifier := Identifier{}
				identifier.Name = e.(map[string]interface{})["name"].(string)

				identifier.All = false
				if identifier.Name == "*" {
					identifier.All = true
				}

				identifier.Type = IdentifierType(e.(map[string]interface{})["type"].(string))

				if identifier.Type == TableIdentType || identifier.Type == IndexIdentType {
					return nil, errors.New("database is required in the element " + strconv.Itoa(i))
				}

				identifier.Options = make(map[string]string)
				identifier.Options["database"] = e.(map[string]interface{})["database"].(string)
				if identifier.Type == IndexIdentType {
					if _, ok := e.(map[string]interface{})["table"]; !ok {
						return nil, errors.New("table is required in the element " + strconv.Itoa(i))
					}
					identifier.Options["table"] = e.(map[string]interface{})["table"].(string)
				}

				permissions.Elements = append(permissions.Elements, identifier)
			}
			newRole.Permissions = append(newRole.Permissions, permissions)
		}
	}

	// Store in database
	table, err := database.GetTable("system", "roles")
	if err != nil {
		return nil, err
	}

	// Check if role already exists
	rawRows, err := table.Get("id", name)
	if err != nil {
		return nil, err
	}

	if len(rawRows) > 0 {
		return nil, errors.New("role already exists")
	}

	// Convert to Blob
	role, err := utils.Serialize(newRole.Permissions)

	if err != nil {
		return nil, err
	}

	// Insert into database
	_, err = table.Insert([]database.RawRow{{"id": name, "admin": newRole.IsAdmin, "permissions": role}})
	if err != nil {
		return nil, err
	}

	// Store in memory
	allRoles[name] = newRole

	return newRole, nil
}

func GetRole(name string) (*Role, error) {
	// Check if role exists

	if allRoles["name"] != nil {
		return allRoles["name"], nil
	}

	table, err := database.GetTable("system", "roles")

	if err != nil {
		return nil, err
	}

	// Get the role from the database
	rawRows, err := table.Get("id", name)

	if err != nil {
		return nil, err
	}

	if len(rawRows) == 0 {
		return nil, errors.New("role does not exist")
	}

	// Convert raw rows to role
	role := &Role{
		Name:        rawRows[0]["id"].(string),
		IsAdmin:     rawRows[0]["admin"].(bool),
		Permissions: make([]Permissions, 0),
	}

	// Convert permissions
	perm := rawRows[0]["permissions"].([]byte)
	var dst []Permissions

	err = utils.Deserialize(perm, &dst)

	if err != nil {
		return nil, err
	}

	role.Permissions = dst

	allRoles[name] = role
	// This is a placeholder implementation
	return role, nil
}

func GetRoles(names []string) ([]Role, error) {
	roles := make([]Role, 0)

	for _, name := range names {
		role, err := GetRole(name)
		if err != nil {
			return nil, err
		}

		if role == nil {
			return nil, errors.New("role does not exist")
		}

		roles = append(roles, *role)
	}

	return roles, nil
}

func LoadAllRoles() error {
	allRoles = make(map[string]*Role)

	table, err := database.GetTable("system", "roles")

	if err != nil {
		return err
	}

	// Get all roles from the database
	rawRows := table.Range(nil, -1, -1)

	for _, rawRow := range rawRows {
		role := &Role{
			Name:        rawRow["id"].(string),
			IsAdmin:     rawRow["admin"].(bool),
			Permissions: make([]Permissions, 0),
		}

		// Convert permissions
		perm := rawRow["permissions"].([]byte)
		var dst []Permissions

		err = utils.Deserialize(perm, &dst)

		if err != nil {
			return err
		}

		role.Permissions = dst

		allRoles[role.Name] = role
	}

	return nil
}

func DeleteRole(name string) error {
	// Check if role exists
	role, err := GetRole(name)
	if err != nil {
		return err
	}

	if role == nil {
		return errors.New("role does not exist")
	}

	// Remove role from the in-memory map
	delete(allRoles, name)
	// Remove role from the database
	table, err := database.GetTable("system", "roles")
	if err != nil {
		return err
	}

	// Remove role from the database
	_, err = table.Delete([]database.RawRow{{"id": role.Name}})
	if err != nil {
		return err
	}

	return nil
}
