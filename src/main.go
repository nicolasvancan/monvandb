package main

import (
	"fmt"
	"os"

	"golang.org/x/exp/mmap"
)

func main() {
	// Lets try to dumpout some structure to the path
	// Btree
	//bTree := NewTree()
	//bTree.SetName("Teste do papis aqui kkkkkkk")
	//bTree.SetRoot(12345678910)
	//bTree := loadMMapFromFile()
	//fmt.Printf("BTree:\n- Name = %s\n- Root = %d\n", bTree.GetName(), bTree.GetRoot())
	//dumpDataIntoFile(bTree)
	//n := createTreeNodeAndInsertValuesToIt()
	//
	//itens := n.getNItens()
	//free := n.getFreeBytes()
	//info, _ := n.getIdxPointer(1)
	//fmt.Printf("Itens Number = %d, Memory Free = %d, key = %d, addr = %d\n", itens, free, info.key, info.addr)
	createFileTest()
}

func createFileTest() {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste1.db"
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

func createTreeNodeAndInsertValuesToIt() *TreeNode {
	node := NewNodeNode()
	err := node.InsertNodeNewChild(2, 1234)
	node.InsertNodeNewChild(1, 11)
	if err != nil {
		fmt.Println(err)
	}
	return node
}

func dumpDataIntoFile(tree *BTree) {
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/teste.db"
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	verifyAndPanic(err)

	f.Write(tree.GetBytes())
	f.Close()
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
	return LoadTree(bTreeHeader)
}
