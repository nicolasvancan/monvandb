package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/exp/mmap"
)

func main() {
	//os.Remove("/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db")
	//testeCreateDb()
	bTree := testeLoadDb()
	fmt.Println(bTree.GetName())
	fmt.Println(bTree.GetRoot())

	bTree.New = NewPage
	bTree.Get = getPage
	bTree.Set = SetPage
	bTree.SetHeader = SetHeader

	//fmt.Printf("Root page = %d\n", bTree.GetRoot())
	for i := 0; i < 1; i++ {
		BTreeInsert(bTree, []byte(strconv.Itoa(i)), []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))
	}

}

type Teste struct {
	JabulaniAhhhh   string
	BangssauroAAAAA int
	Groselia        []string
}

func createFileTest() {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db"
	mapped, err := mmap.Open(fileName)
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0664)
	verifyAndPanic(err)

	len := mapped.Len()
	fmt.Printf("Number before mapping = %d\n", len)
	n, err := f.WriteAt([]byte("Teste2\n"), int64(len))
	len = mapped.Len()
	defer f.Close()
	defer mapped.Close()
	fmt.Printf("Number written = %d and number after mapping = %d\n", n, len)
}

func verifyAndPanic(e error) {
	if e != nil {
		panic(e)
	}
}

func loadMMapFromFile() *BTree {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db"
	fmap, err := mmap.Open(fileName)
	verifyAndPanic(err)
	bTreeHeader := make([]byte, PAGE_SIZE)
	_, err = fmap.ReadAt(bTreeHeader, 0)
	verifyAndPanic(err)
	fmap.Close()
	return LoadTree(bTreeHeader, 4096)
}

func testLeafValues() *TreeNode {
	leaf := NewNodeLeaf()
	m := Teste{
		JabulaniAhhhh:   "Meu Nome é nicolas",
		BangssauroAAAAA: 10,
		Groselia:        []string{"opa", "xulapa"},
	}

	var bin_buf bytes.Buffer
	enc := gob.NewEncoder(&bin_buf)
	enc.Encode(m)

	leaf.PutLeafNewKeyValue([]byte("34"), bin_buf.Bytes())
	leaf.PutLeafNewKeyValue([]byte("35"), []byte("BUTATATAN"))

	return leaf
}

func testeSplits() {
	node := NewNodeNode()

	node.PutNodeNewChild([]byte("Opa e i"), 12)
	fmt.Printf("Free bytes here = %d and number Items %d\n", getFreeBytes(node), node.GetNItens())
	node.PutNodeNewChild([]byte("3asda54ds"), 14)
	fmt.Printf("Free bytes here = %d and number Items %d\n", getFreeBytes(node), node.GetNItens())
	node.PutNodeNewChild([]byte("asdasdasdasd"), 15)
	fmt.Printf("Free bytes here = %d and number Items %d\n", getFreeBytes(node), node.GetNItens())
	node.PutNodeNewChild([]byte("2"), 17)
	fmt.Printf("Free bytes here = %d and number Items %d\n", getFreeBytes(node), node.GetNItens())

	fmt.Printf("Chave é essa %s\n", node.GetNodeChildByKey([]byte("Opa e i")).addr)
}

func testeCreateDb() {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db"
	bTree := NewTree(4096)
	bTree.SetName("my_first_db")
	bTree.SetRoot(0)

	fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	fp.Write(bTree.data)
	fp.Close()
}

// Global variable to be used for testing purpose
var Fp, err = os.OpenFile("/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db", os.O_CREATE|os.O_RDWR, 0666)

func getPage(page uint64) TreeNode {
	_, data, err := MmapPage(Fp, page, 4096)
	if err != nil {
		fmt.Println(err)
	}

	return *LoadTreeNode(data)
}

func SetPage(node TreeNode, page uint64) bool {
	_, err := Fp.WriteAt(node.data, int64(page*PAGE_SIZE))
	if err != nil {
		fmt.Errorf("Error writing page %d\nError = %w\n", page, err)
		return false
	}

	return true
}

func NewPage(node TreeNode) uint64 {

	// get Stat from file
	fileInfo, err := Fp.Stat()

	if err != nil {
		fmt.Errorf("Could not write fileInfo %w\n", err)
	}

	// Without header
	lastPage := (fileInfo.Size() / PAGE_SIZE)

	Fp.WriteAt(node.data, fileInfo.Size())

	return uint64(lastPage)
}

func SetHeader(bTree BTree) {
	Fp.WriteAt(bTree.data, 0)
}

func testeLoadDb() *BTree {
	docSize, data, err := MmapPage(Fp, 0, 4096)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Total document lenght %d\n", docSize)
	return LoadTree(data, 4096)
}
