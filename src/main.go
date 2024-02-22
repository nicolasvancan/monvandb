package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"

	bTree "github.com/nicolasvancan/monvandb/src/btree"
	files "github.com/nicolasvancan/monvandb/src/files"
	"golang.org/x/exp/mmap"
)

func main() {
	//fmt.Printf("Comparsion between these bytes 0 and 196 in string %d\n", bytes.Compare([]byte("99"), []byte("988")))
	//os.Remove("/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db")
	testeCreateDb()
	tree := testeLoadDb()
	fmt.Println(tree.GetName())
	fmt.Println(tree.GetRoot())
	//
	tree.New = NewPage
	tree.Get = getPage
	tree.Set = SetPage
	tree.SetHeader = SetHeader
	//
	fmt.Printf("Root page = %d\n", tree.GetRoot())
	for i := 182; i <= 256; i++ {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(i))
		bTree.BTreeInsert(tree, b, []byte(string("teste_"+strconv.Itoa(i))))
	}
	/*
	   // For Node
	   page2 := tree.Get(2)
	   nItens := page2.GetNItens()

	   	for i := 0; i < int(nItens); i++ {
	   		k := page2.GetNodeChildByIndex(i)
	   		keyInt := binary.LittleEndian.Uint32(k.GetKey())
	   		fmt.Printf("Key %d and page %d\n", keyInt, k.GetAddr())
	   	}

	   // For leaf

	   page3 := tree.Get(3)
	   nItensL := page3.GetNItens()

	   	for i := 0; i < int(nItensL); i++ {
	   		k := page3.GetLeafKeyValueByIndex(uint16(i))
	   		keyInt := binary.LittleEndian.Uint32(k.GetKey())
	   		fmt.Printf("Key %d and page %s\n", keyInt, k.GetValue())
	   	}
	*/
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

func loadMMapFromFile() *bTree.BTree {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db"
	fmap, err := mmap.Open(fileName)
	verifyAndPanic(err)
	bTreeHeader := make([]byte, bTree.PAGE_SIZE)
	_, err = fmap.ReadAt(bTreeHeader, 0)
	verifyAndPanic(err)
	fmap.Close()
	return bTree.LoadTree(bTreeHeader, 4096)
}

func testLeafValues() *bTree.TreeNode {
	leaf := bTree.NewNodeLeaf()
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

func testeCreateDb() {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db"
	bTree := bTree.NewTree(4096)
	bTree.SetName("my_first_db")
	bTree.SetRoot(0)

	fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	fp.Write(bTree.GetBytes())
	fp.Close()
}

// Global variable to be used for testing purpose
var Fp, err = os.OpenFile("/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db", os.O_CREATE|os.O_RDWR, 0666)

func getPage(page uint64) bTree.TreeNode {
	_, data, err := files.MmapPage(Fp, page, 4096)
	if err != nil {
		fmt.Println(err)
	}

	return *bTree.LoadTreeNode(data)
}

func SetPage(node bTree.TreeNode, page uint64) bool {
	_, err := Fp.WriteAt(node.GetBytes(), int64(page*bTree.PAGE_SIZE))
	if err != nil {
		fmt.Errorf("Error writing page %d\nError = %w\n", page, err)
		return false
	}

	return true
}

func NewPage(node bTree.TreeNode) uint64 {

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

func SetHeader(bTree bTree.BTree) {
	Fp.WriteAt(bTree.GetBytes(), 0)
}

func testeLoadDb() *bTree.BTree {
	docSize, data, err := files.MmapPage(Fp, 0, 4096)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Total document lenght %d\n", docSize)
	return bTree.LoadTree(data, 4096)
}
