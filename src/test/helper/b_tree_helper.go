package helper

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
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
	tmp := make([]byte, len(data))
	copy(tmp, data)
	return *bTree.LoadTreeNode(tmp)
}

func setPage(node bTree.TreeNode, page uint64) bool {
	_, err := Fp.WriteAt(node.GetBytes(), int64(page*bTree.PAGE_SIZE))
	if err != nil {
		fmt.Println(fmt.Errorf("could not write to page %d", page))
		return false
	}

	return true
}

func newPage(node bTree.TreeNode) uint64 {

	// get Stat from file
	fileInfo, err := Fp.Stat()

	if err != nil {
		panic(err)
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

func CreateFakeDbPagesForMapping(t *testing.T, tree *bTree.BTree) *bTree.BTree {

	/*
						(1 9)
						/   \
				 (1 3 5)     (7 8 9)
				 / / /         \ \ \
		(1 2) (3 4) (5 6)     (7)(8)(9)
	*/
	// Create First page
	firstPage := bTree.NewNodeNode()

	// Create first page with following pages
	firstPage.PutNodeNewChild([]byte("1"), 2)
	firstPage.PutNodeNewChild([]byte("9"), 3)

	// Create first page
	tree.New(*firstPage)
	tree.SetRoot(1)
	tree.SetHeader(*tree)
	// Create second node
	secondNode := bTree.NewNodeNode()
	secondNode.PutNodeNewChild([]byte("1"), 4)
	secondNode.PutNodeNewChild([]byte("3"), 5)
	secondNode.PutNodeNewChild([]byte("5"), 6)

	thirdNode := bTree.NewNodeNode()
	thirdNode.PutNodeNewChild([]byte("7"), 7)
	thirdNode.PutNodeNewChild([]byte("8"), 8)
	thirdNode.PutNodeNewChild([]byte("9"), 9)

	tree.New(*secondNode) // 2
	tree.New(*thirdNode)  // 3

	// Create leaves
	firstLeaf := bTree.NewNodeLeaf()
	firstLeaf.PutLeafNewKeyValue([]byte("1"), []byte("teste"))
	firstLeaf.PutLeafNewKeyValue([]byte("2"), []byte("teste"))

	// Create leaves
	secondLeaf := bTree.NewNodeLeaf()
	secondLeaf.PutLeafNewKeyValue([]byte("3"), []byte("teste"))
	secondLeaf.PutLeafNewKeyValue([]byte("4"), []byte("teste"))

	// Create leaves
	thirdLeaf := bTree.NewNodeLeaf()
	thirdLeaf.PutLeafNewKeyValue([]byte("5"), []byte("teste"))
	thirdLeaf.PutLeafNewKeyValue([]byte("6"), []byte("teste"))

	// Create leaves
	fourthLeaf := bTree.NewNodeLeaf()
	fourthLeaf.PutLeafNewKeyValue([]byte("7"), []byte("teste"))

	fithLeaf := bTree.NewNodeLeaf()
	fithLeaf.PutLeafNewKeyValue([]byte("8"), []byte("teste"))

	sixthLeaf := bTree.NewNodeLeaf()
	sixthLeaf.PutLeafNewKeyValue([]byte("9"), []byte("teste"))

	tree.New(*firstLeaf)  // 4
	tree.New(*secondLeaf) // 5
	tree.New(*thirdLeaf)  // 6
	tree.New(*fourthLeaf) // 7
	tree.New(*fithLeaf)   // 8
	tree.New(*sixthLeaf)  // 9

	return tree
}

func FillUpLeafUntilItSplits(btree *bTree.BTree) {
	// Fill up leaf until it splits
	for i := 0; i < 282; i++ {
		bTree.BTreeInsert(btree, []byte(strconv.Itoa(i)), []byte(string("teste_"+strconv.Itoa(i))))
	}
}

func FillUpLeafWithNumericValuesUntilItSplits(btree *bTree.BTree, number int, offset int) {
	for i := 1 + offset; i <= number+offset; i++ {
		// I'm going to use little endian 32 bits so 4 bytes
		integerBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(integerBytes, uint32(i))
		bTree.BTreeInsert(btree, integerBytes, []byte(string("teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_teste_"+strconv.Itoa(i))))
	}
}

func FillUpLeavesWith16kvalues(btree *bTree.BTree, number int, offset int) {
	for i := 1 + offset; i <= number+offset; i++ {
		// I'm going to use little endian 32 bits so 4 bytes
		integerBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(integerBytes, uint32(i))
		bTree.BTreeInsert(btree, integerBytes, CreateValueOf16kLen())
	}
}

// Basic setup for testing
func SetupTests(t *testing.T) string {
	// Create tmp file path
	tmpFilePath := t.TempDir()
	t.Logf("created tmpFile path %s\n", tmpFilePath)
	// Create a new bTree
	return CreateBtreeFileAndSetFile(t, tmpFilePath)
}

func CleanUp() {
	// We close Fp after test is concluded
	fmt.Println("Cleaning up test")
	defer Fp.Close()
}

func FillUpNode(t string) *bTree.TreeNode {
	// Fill up Node until it reaches max Limit
	var newNode *bTree.TreeNode

	if t == "node" {
		newNode = bTree.NewNodeNode()
		for i := 0; i < 322; i++ {
			newNode.PutNodeNewChild([]byte(strconv.Itoa(i)), uint64(i))
		}

	} else {
		newNode = bTree.NewNodeLeaf()
		for i := 0; i < 280; i++ {
			newNode.PutLeafNewKeyValue([]byte(strconv.Itoa(i)), []byte(string(strconv.Itoa(i)+"teste")))
		}
	}

	return newNode
}

func CreateValueOf16kLen() []byte {
	r := make([]byte, 0)

	for i := 0; i < 16450; i++ {
		r = append(r, byte(i%255))
	}
	return r
}
