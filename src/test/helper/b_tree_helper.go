package test_helper_b_tree

import (
	"fmt"
	"os"
	"testing"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
	files "github.com/nicolasvancan/monvandb/src/files"
)

// Global (Exported only for tests purpose)
var Fp *os.File = nil

func CreateBtreeFileAndSetFile(t *testing.T, basePath string) string {
	fileName := basePath + string(os.PathSeparator) + "teste.db"
	bTree := bTree.NewTree(4096)
	// Fake name to test
	bTree.SetName("db_teste")
	// Should start at zero
	bTree.SetRoot(0)
	// We first create a db at the tmp folder
	fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
	Fp = fp
	if err != nil {
		panic(err)
	}

	fp.Write(bTree.GetBytes())
	// Returns file path
	return fileName
}

// Callback functions
func getPage(page uint64) bTree.TreeNode {
	_, data, err := files.MmapPage(Fp, page, 4096)
	if err != nil {
		fmt.Println(err)
	}
	return *bTree.LoadTreeNode(data)
}

func setPage(node bTree.TreeNode, page uint64) bool {
	_, err := Fp.WriteAt(node.GetBytes(), int64(page*bTree.PAGE_SIZE))
	if err != nil {
		fmt.Errorf("Error writing page %d\nError = %w\n", page, err)
		return false
	}

	return true
}

func newPage(node bTree.TreeNode) uint64 {

	// get Stat from file
	fileInfo, err := Fp.Stat()

	if err != nil {
		fmt.Errorf("Could not write fileInfo %w\n", err)
	}

	// Without header
	lastPage := (fileInfo.Size() / bTree.PAGE_SIZE)

	Fp.WriteAt(node.GetBytes(), fileInfo.Size())

	return uint64(lastPage)
}

func delPage(page uint64) {

}

func setHeader(bTree bTree.BTree) {
	Fp.WriteAt(bTree.GetBytes(), 0)
}

func LoadBTreeFromPath(t *testing.T, filepath string) *bTree.BTree {
	_, data, err := files.MmapPage(Fp, 0, 4096)

	if err != nil {
		t.Errorf("Could not load temporaray database bTree on page 0")
	}
	tmpBTree := bTree.LoadTree(data, 4096)
	// Assign callback functions
	tmpBTree.Get = getPage
	tmpBTree.New = newPage
	tmpBTree.Set = setPage
	tmpBTree.SetHeader = setHeader
	tmpBTree.Del = delPage

	return tmpBTree
}
