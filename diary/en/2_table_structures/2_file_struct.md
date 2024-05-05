# File Struct

Before advancing more on the table topic, I decided that I had to finish one task that, for some reason, I tought it could wait a little more, which is the file access struct. Notably, I have develop the functional binary tree, but I have just created file access functions for testing purpose and not production purpose.

Now it's time to create and incorporate all the fields and possible informations into one Structure. Because I don't kwnow what changes I'll commit to the files management system, I'll just create the basic structure with the basic callbacks for the table, alongside with the basic informations for my binary trees files.

Knowing that for every binaryTree file, there must be a file path and also a File struct, let's create a Struct containing both a bTree pointer as well as a path and File pointer. This struct will be named as **DataFile**.

```go
type DataFile struct {
	path  string
	bTree *bTree.BTree
	fp    *os.File
}

```

With this struct, I can easly create or load a DataFile, which is a file that contains Data, in our case currently the binary tree. To enable the creation of binary trees and it's file easly, I decided to create a function **OpenDataFile**, whose responsabilities are: create a binary tree file if it doesn't exist; load the binary tree is exists; load all callbacks functions for the file (This can be a complex system later on), and load the File pointer onto fp.

Although being possible to use separetedly all methodes for the binary tree function loading its module, I want to use the DataFile interface for all possible operations with one specific DataFile, this way I'll also implement all CRUD functions in this struct, as follows:

```go
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
```

The callback functions are anonimous functions written specifically to this purpose of dealling with the File pointer (fp) related to the binary Tree file. One function was created to initialize them and put all callbacks into their respective variable of the bTree pointer.

```go
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

```

# Next steps

This was an important step for advancing with my projects but was not direcly related to tables structures and so on. Now that I have an interface to access and use the binary tree easly, stored in folder files, I'll pursuit the table divisions. 

