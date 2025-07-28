package monvan_parser

import (
	"encoding/json"
	"errors"
)

/*
Since this parser is simple and theoretically it will not contain any Tree for the moment
the AST will be a simple struct with the query and the tokens
*/

type Json = map[string]interface{}

// Alter operation
type AlterOperation = string

const (
	AddColumn    AlterOperation = "ADD COLUMN"
	ModifyColumn AlterOperation = "MODIFY COLUMN"
	DropColumn   AlterOperation = "DROP COLUMN"
)

type Statement interface {
	isStatement()
}

func Parse(query string) (Statement, bool, error) {
	lexer := NewLexer(query)

	// Tokenize the query
	tokens, err, shouldPass := lexer.Tokenize()

	if err != nil {
		return nil, shouldPass, err
	}

	stmt, err := getStatement(tokens)

	return stmt, false, err
}

// Types of statements
type UserCreate struct {
	Name     string
	Username string
	Password string
}

type UserDrop struct {
	Name string
}

type UserModify struct {
	Name     string
	Username string
	Password string
}

type DatabaseCreate struct {
	Name string
}

type DatabaseDrop struct {
	Name string
}

type IndexCreate struct {
	Name    string
	Table   string
	Columns []string
}

type IndexDrop struct {
	Name  string
	Table string
}

type RoleCreate struct {
	Name    string
	Options Json
}

type RoleAlter struct {
	Name    string
	Options Json
}

type RoleDrop struct {
	Name string
}

type RoleGrant struct {
	Role string
	To   string
}

type RoleRevoke struct {
	Role string
	From string
}

type TableDrop struct {
	Name string
}

type TableAlter struct {
	Name          string
	Operation     AlterOperation
	Column        string
	ModifyType    string
	ModifyOptions string
}

type UseDatabase struct {
	Name string
}

type ShowDatabases struct{}

type ShowTables struct {
	DatabaseName string
}

func (u *ShowTables) isStatement()     {}
func (u *ShowDatabases) isStatement()  {}
func (u *UseDatabase) isStatement()    {}
func (u *UserCreate) isStatement()     {}
func (u *UserDrop) isStatement()       {}
func (u *UserModify) isStatement()     {}
func (d *DatabaseCreate) isStatement() {}
func (d *DatabaseDrop) isStatement()   {}
func (i *IndexCreate) isStatement()    {}
func (i *IndexDrop) isStatement()      {}
func (r *RoleCreate) isStatement()     {}
func (r *RoleAlter) isStatement()      {}
func (r *RoleDrop) isStatement()       {}
func (r *RoleGrant) isStatement()      {}
func (r *RoleRevoke) isStatement()     {}
func (t *TableDrop) isStatement()      {}
func (t *TableAlter) isStatement()     {}

func getUseDatabase(tokens []*Token) *UseDatabase {
	return &UseDatabase{
		Name: tokens[2].Value,
	}
}

func getUserCreate(tokens []*Token) (*UserCreate, error) {

	// example:
	// CREATE USER nicolas JSON {"username": "nicolas", "password": "nicolas"}
	// The user name (Index) is nicolas and username and password are in the JSON
	// The JSON is a string, so we need to parse it

	tokenSlice := tokens[4:]

	// convert to Json
	convertedJson, err := tokenSliceToJson(tokenSlice)

	if err != nil {
		return nil, err
	}

	if convertedJson == nil {
		return nil, errors.New("error converting to json")
	}

	// check if the json has the username and password
	if _, ok := convertedJson["username"]; !ok {
		return nil, errors.New("error: username not found in json")
	}

	if _, ok := convertedJson["password"]; !ok {
		return nil, errors.New("error: password not found in json")
	}

	return &UserCreate{
		Name:     tokens[2].Value,
		Username: convertedJson["username"].(string),
		Password: convertedJson["password"].(string),
	}, nil
}

func getUserDrop(tokens []*Token) *UserDrop {
	return &UserDrop{
		Name: tokens[2].Value,
	}
}

func getUserModify(tokens []*Token) (*UserModify, error) {

	tokenSlice := tokens[4:]

	// convert to Json
	convertedJson, err := tokenSliceToJson(tokenSlice)

	if err != nil {
		return nil, err
	}

	if convertedJson == nil {
		return nil, errors.New("error converting to json")
	}

	if convertedJson["username"] == nil && convertedJson["password"] == nil {
		return nil, errors.New("error: username not found in json")
	}

	username := ""
	password := ""

	if convertedJson["username"] != nil {
		username = convertedJson["username"].(string)
	}

	if convertedJson["password"] != nil {
		password = convertedJson["password"].(string)
	}

	return &UserModify{
		Name:     tokens[2].Value,
		Username: username,
		Password: password,
	}, nil
}

func getDatabaseCreate(tokens []*Token) *DatabaseCreate {
	return &DatabaseCreate{
		Name: tokens[2].Value,
	}
}

func getDatabaseDrop(tokens []*Token) *DatabaseDrop {
	return &DatabaseDrop{
		Name: tokens[2].Value,
	}
}

func getIndexCreate(tokens []*Token) *IndexCreate {

	// Get columns
	cols := make([]string, 0)

	for _, token := range tokens {
		if token.Type == TokenIdentifier {
			cols = append(cols, token.Value)
		}
	}

	return &IndexCreate{
		Name:    tokens[2].Value,
		Table:   tokens[4].Value,
		Columns: cols,
	}
}

func getIndexDrop(tokens []*Token) *IndexDrop {
	return &IndexDrop{
		Name:  tokens[2].Value,
		Table: tokens[4].Value,
	}
}

func getRoleCreate(tokens []*Token) (*RoleCreate, error) {

	// Just parse the json and send it
	tokenSlice := tokens[4:]
	// convert to Json
	convertedJson, err := tokenSliceToJson(tokenSlice)
	if err != nil {
		return nil, err
	}
	if convertedJson == nil {
		return nil, err
	}

	return &RoleCreate{
		Name:    tokens[2].Value,
		Options: convertedJson,
	}, nil
}

func getRoleDrop(tokens []*Token) *RoleDrop {
	return &RoleDrop{
		Name: tokens[2].Value,
	}
}

func getRoleGrant(tokens []*Token) *RoleGrant {
	return &RoleGrant{
		Role: tokens[2].Value,
		To:   tokens[4].Value,
	}
}

func getRoleRevoke(tokens []*Token) *RoleRevoke {
	return &RoleRevoke{
		Role: tokens[2].Value,
		From: tokens[4].Value,
	}
}

func getTableDrop(tokens []*Token) *TableDrop {
	return &TableDrop{
		Name: tokens[2].Value,
	}
}

func getTableAlter(tokens []*Token) *TableAlter {

	if tokens[3].LowerValue == "modify" {
		modifyOptions := ""

		if len(tokens) > 6 {
			modifyOptions = tokens[7].Value
		}
		return &TableAlter{
			Name:          tokens[2].Value,
			Operation:     ModifyColumn,
			Column:        tokens[4].Value,
			ModifyType:    tokens[5].Value,
			ModifyOptions: modifyOptions,
		}

	}

	return &TableAlter{
		Name:      tokens[2].Value,
		Operation: AlterOperation(tokens[3].Value),
		Column:    tokens[5].Value,
	}
}

func getStatement(tokens []*Token) (Statement, error) {
	switch tokens[0].LowerValue {
	case "use":
		return getUseDatabase(tokens), nil
	case "create":
		switch tokens[1].LowerValue {
		case "user":
			stmt, err := getUserCreate(tokens)
			if err != nil {
				return nil, err
			}

			return stmt, nil
		case "database":
			return getDatabaseCreate(tokens), nil
		case "index":
			return getIndexCreate(tokens), nil
		case "role":
			stmt, err := getRoleCreate(tokens)
			if err != nil {
				return nil, err
			}

			return stmt, nil
		}
	case "drop":
		switch tokens[1].LowerValue {
		case "user":
			return getUserDrop(tokens), nil
		case "database":
			return getDatabaseDrop(tokens), nil
		case "index":
			return getIndexDrop(tokens), nil
		case "role":
			return getRoleDrop(tokens), nil
		case "table":
			return getTableDrop(tokens), nil
		}
	case "alter":
		switch tokens[1].LowerValue {
		case "user":
			stmt, err := getUserModify(tokens)
			if err != nil {
				return nil, err
			}
			return stmt, nil
		case "table":
			return getTableAlter(tokens), nil
		}
	case "grant":
		return getRoleGrant(tokens), nil
	case "revoke":
		return getRoleRevoke(tokens), nil
	case "show":
		switch tokens[1].LowerValue {
		case "databases":
			return &ShowDatabases{}, nil
		case "tables":
			return &ShowTables{
				DatabaseName: tokens[2].Value,
			}, nil
		}
	}

	return nil, errors.New("unknown statement")
}

func tokenSliceToJson(array []*Token) (Json, error) {
	jsonString := ""
	res := make(Json)

	for _, v := range array {

		//iterate over string
		newString := ""
		for i := 0; i < len(v.Value); i++ {
			tmp := v.Value[i]
			if v.Value[i] == '\'' {
				tmp = '"'
			}
			newString += string(tmp)
		}

		jsonString += newString
	}

	err := json.Unmarshal([]byte(jsonString), &res)

	if err != nil {
		return nil, err
	}

	return res, nil
}
