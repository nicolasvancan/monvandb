package utils

import (
	"fmt"
	"os"
)

func init() {
	os.Setenv("MONVANDB_PATH", "/tmp")
}

// Database file names
const (
	SEPARATOR        = string(os.PathSeparator)
	METDATA_FILE     = "metadata.json"
	TABLE_FILE       = "table.db"
	TABLE_LOGS_FIILE = "hist.dblg"
	INDICES_FOLDER   = "indices"
)

var systemFolders = []string{"base", "databases", "users", "system"}

// Create base Folders for holding database and other files
func CreateBaseFolders(basePath string) error {
	// Set the environment variable
	if basePath != "" {
		os.Setenv("MONVANDB_PATH", basePath)
	}

	// Create the base directory

	for _, value := range systemFolders {
		fmt.Println(os.Getenv("MONVANDB_PATH"))
		fmt.Println(GetPath(value))

		err := os.MkdirAll(GetPath(value), os.ModePerm)

		if err != nil {
			fmt.Println("Error creating monvandb directory:", err)
			return err
		}
	}

	return nil
}

func WriteToFile(path string, data []byte) error {
	// Check if the file exists
	if !FileExists(path) {
		return fmt.Errorf("file %s does not exist", path)
	}

	// Open the file
	f, err := os.OpenFile(path, os.O_RDWR, 0644)

	if err != nil {
		return err
	}

	// Write data to the file
	_, err = f.Write(data)

	if err != nil {
		return err
	}

	defer f.Close()

	return nil
}

func CreateFolder(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func ReadFromFile(path string) ([]byte, error) {
	// Check if the file exists
	if !FileExists(path) {
		return nil, fmt.Errorf("file %s does not exist", path)
	}

	// Open the file
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	// Read the file
	data, err := os.ReadFile(f.Name())

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return data, nil
}

func CreateFile(path string) (*os.File, error) {
	// Create file based on given path
	f, err := os.Create(path)

	return f, err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ListFilesInDir(dirPath string) ([]os.FileInfo, error) {
	var files []os.FileInfo

	// Open the directory
	d, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	// Read the files in the directory
	filesInfo, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Add the files to the list
	files = append(files, filesInfo...)

	return files, nil
}

// Get path for the mapped systems paths, that contain specific files for the system
func GetPath(path string) string {

	var monvanPaths = map[string]string{
		"base":      os.Getenv("MONVANDB_PATH") + string(os.PathSeparator) + "monvandb",
		"databases": os.Getenv("MONVANDB_PATH") + string(os.PathSeparator) + "monvandb" + string(os.PathSeparator) + "databases",
		"users":     os.Getenv("MONVANDB_PATH") + string(os.PathSeparator) + "monvandb" + string(os.PathSeparator) + "users",
		"system":    os.Getenv("MONVANDB_PATH") + string(os.PathSeparator) + "monvandb" + string(os.PathSeparator) + "system",
	}
	return monvanPaths[path]
}
