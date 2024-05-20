package main

import (
	"fmt"
	"os"
	"testing"

	database "github.com/nicolasvancan/monvandb/src/database"
	utils "github.com/nicolasvancan/monvandb/src/utils"
)

var err_creating_file_string string = "error creating file: %s"
var err_creating_base_folders_string string = "error creating base folders: %s"
var file_is_nil_string string = "file is nil"

func createBaseFolders(t *testing.T) string {
	// CreateBaseFolders creates the base folders for holding database and other files
	basePath := t.TempDir()
	// CreateBaseFolders creates the base folders for holding database and other files
	err := utils.CreateBaseFolders(basePath)

	if err != nil {
		t.Errorf(err_creating_base_folders_string, err)
	}
	return basePath
}

func createFile(t *testing.T, basePath string) *os.File {
	// CreateFile creates a file in the specified path
	file, err := utils.CreateFile(basePath + string(os.PathSeparator) + "test.db")

	if err != nil {
		t.Errorf(err_creating_file_string, err)
	}

	if file == nil {
		t.Error(file_is_nil_string)
	}

	return file
}
func TestCreateBaseFolders(t *testing.T) {
	// Get tmp path
	basePath := createBaseFolders(t)

	// verify if all folders were created
	if _, err := os.Stat(basePath +
		string(os.PathSeparator) +
		"monvandb"); os.IsNotExist(err) {
		t.Errorf("Error creating monvandb directory: %s", err)
	}

	if _, err := os.Stat(basePath +
		string(os.PathSeparator) +
		"monvandb" + string(os.PathSeparator) + "databases"); os.IsNotExist(err) {
		t.Errorf("Error creating monvandb/databases directory: %s", err)
	}

	if _, err := os.Stat(basePath +
		string(os.PathSeparator) +
		"monvandb" + string(os.PathSeparator) + "users"); os.IsNotExist(err) {
		t.Errorf("Error creating monvandb/users directory: %s", err)
	}

}

func TestListFilesInDir(t *testing.T) {
	errorString := "Error listing files in directory: %s"
	// Get tmp path
	basePath := createBaseFolders(t)

	os.Create(basePath + string(os.PathSeparator) + "monvandb" + string(os.PathSeparator) + "databases" + string(os.PathSeparator) + "test1.db")

	// ListFilesInDir lists all files in a directory
	files, err := utils.ListFilesInDir(basePath + string(os.PathSeparator) + "monvandb")
	fmt.Println(files)
	if err != nil {
		t.Errorf(errorString, err)
	}

	if len(files) != 2 {
		t.Errorf(errorString, err)
	}

	files, err = utils.ListFilesInDir(basePath + string(os.PathSeparator) + "monvandb" + string(os.PathSeparator) + "databases")

	if err != nil {
		t.Errorf(errorString, err)
	}

	if len(files) != 1 {
		t.Errorf(errorString, err)
	}

	if files[0].Name() != "teste1.db" {
		t.Error("Not file name")
	}

}

func TestCreateFile(t *testing.T) {
	// Get tmp path
	basePath := createBaseFolders(t)

	// CreateFile creates a file in the specified path
	file := createFile(t, basePath)

	if file == nil {
		t.Error(file_is_nil_string)
	}

}

func TestWriteAndReadToFile(t *testing.T) {
	// Get tmp path
	basePath := createBaseFolders(t)
	// CreateFile creates a file in the specified path
	file := createFile(t, basePath)

	if file == nil {
		t.Error(file_is_nil_string)
	}

	filePath := basePath + string(os.PathSeparator) + "test.db"
	// WriteToFile writes the specified content to the file
	err := utils.WriteToFile(filePath, []byte("test"))

	if err != nil {
		t.Errorf("Error writing to file: %s", err)
	}

	fileContent, err := utils.ReadFromFile(filePath)

	if err != nil {
		t.Errorf("Error reading file: %s", err)
	}

	if string(fileContent) != "test" {
		t.Error("File content is not the same")
	}
}

func TestToJson(t *testing.T) {
	// Get tmp path
	basePath := createBaseFolders(t)
	mockFilePath := basePath + string(os.PathSeparator) + "metadata.db"

	// CreateFile creates a file in the specified path
	f, err := utils.CreateFile(mockFilePath)

	if err != nil {
		t.Errorf("Error creating file: %s", err)
	}

	mock := new(database.Table)

	mock.Name = "test"
	mock.Path = mockFilePath
	mock.Columns = []database.Column{
		{
			Name: "test",
			Type: 1,
		},
	}

	json, err := utils.ToJson(&mock)

	if err != nil {
		t.Errorf("Error converting to json: %s", err)
	}
	_, err = f.Write(json)

	if err != nil {
		t.Errorf("Error writing to file: %s", err)
	}

	fileContent, err := utils.ReadFromFile(mockFilePath)

	if err != nil {
		t.Errorf("Error reading file: %s", err)
	}

	if string(fileContent) != string(json) {
		t.Error("File content is not the same")
	}
	decodedJson := new(database.Table)
	err = utils.FromJson(fileContent, decodedJson)

	if err != nil {
		t.Errorf("Error decoding json: %s", err)
	}

	if decodedJson.Name != mock.Name {
		t.Error("Decoded json is not the same")
	}

	if decodedJson.Columns[0].Name != mock.Columns[0].Name {
		t.Error("Decoded json is not the same")
	}
}
