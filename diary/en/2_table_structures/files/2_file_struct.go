package table_structures

// This is just a notebook file, used to store my exaplanation code from the diary folder

import (
	"fmt"
	"os"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
)

// Prototype
type DataFile struct {
	path  string
	bTree *bTree.BTree
	fp    *os.File
}

func OpenDataFile(path string) (*DataFile, error) {
	p := DataFile{
		path:  path,
		bTree: nil,
		fp:    nil,
	}

	p.path = path
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)

	if err != nil {
		return nil, err
	}

	stat, _ := file.Stat()
	// If the file is empty, create a new tree
	if stat.Size() == 0 {
		bTree := bTree.NewTree(bTree.PAGE_SIZE)
		bTree.SetName("bTree")
		p.bTree = bTree
		// Write the tree to the file
		file.WriteAt(bTree.GetBytes(), 0)
		defer file.Sync()
	}

	treeHeader := make([]byte, bTree.PAGE_SIZE)
	file.ReadAt(treeHeader, 0)
	p.fp = file

	p.bTree = bTree.LoadTree(treeHeader, bTree.PAGE_SIZE)

	err = p.loadCallbacks()

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *DataFile) Get(key []byte) []bTree.BTreeKeyValue {
	return bTree.BTreeGet(p.bTree, key)
}

func (p *DataFile) Insert(key []byte, value []byte) {
	bTree.BTreeInsert(p.bTree, key, value)
}

func (p *DataFile) Delete(key []byte) {
	bTree.BTreeDelete(p.bTree, key)
}

func (p *DataFile) Update(key []byte, value []byte) {
	bTree.BTreeUpdate(p.bTree, key, value)
}

func (p *DataFile) ForceSync() {
	defer p.fp.Sync()
}

func (p *DataFile) Close() {
	defer p.fp.Close()
}

func (p *DataFile) loadCallbacks() error {
	// Set callbacks
	p.bTree.Set = func(node bTree.TreeNode, page uint64) bool {
		_, err := p.fp.WriteAt(node.GetBytes(), int64(page*bTree.PAGE_SIZE))

		if err != nil {
			fmt.Println(fmt.Errorf("could not write to page %d", page))
			return false
		}

		return true
	}

	p.bTree.SetHeader = func(bTree bTree.BTree) {
		p.fp.WriteAt(bTree.GetBytes(), 0)
	}

	p.bTree.Get = func(page uint64) bTree.TreeNode {
		data := make([]byte, bTree.PAGE_SIZE)
		_, err := p.fp.ReadAt(data, int64(page*bTree.PAGE_SIZE))

		if err != nil {
			panic(err)
		}

		return *bTree.LoadTreeNode(data)
	}

	p.bTree.Del = nil

	p.bTree.New = func(node bTree.TreeNode) uint64 {

		// get Stat from file
		fileInfo, err := p.fp.Stat()

		if err != nil {
			panic(err)
		}

		// Without header
		lastPage := (fileInfo.Size() / bTree.PAGE_SIZE)

		p.fp.WriteAt(node.GetBytes(), fileInfo.Size())

		return uint64(lastPage)
	}

	return nil
}

// Test function to test something
func FileStruct() {
	dataFile, err := OpenDataFile("test.db")

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	//bTree.BTreeInsert(dataFile.bTree, []byte("3"), []byte("Pedro"))
	res := bTree.BTreeGet(dataFile.bTree, []byte("3"))
	fmt.Printf("%s", res[0].Value)
	//panic("implement me")
}
