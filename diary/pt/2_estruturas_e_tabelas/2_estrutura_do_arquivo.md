# Estrutura do arquivo

Antes de avançar mais no tema tabela, resolvi que tinha que finalizar uma tarefa que, por algum motivo, achei que poderia esperar um pouco mais, que é a struct de acesso a arquivos. De fato desenvolvi a árvore binária funcional (mas não eficiente hehehe), mas criando funções de acesso a arquivos para fins de teste e não para fins de produção.

Agora é hora de criar e incorporar todos os campos e informações possíveis em uma Estrutura. Como não sei quais mudanças farei no sistema de gerenciamento de arquivos, apenas criarei a estrutura básica com os retornos de chamada básicos para a tabela, juntamente com as informações básicas para meus arquivos de árvores binárias.

Sabendo que para cada arquivo binaryTree, deve haver um caminho de arquivo e também uma estrutura de arquivo, vamos criar uma estrutura contendo um ponteiro bTree, bem como um caminho e um ponteiro de arquivo. Esta estrutura será nomeada como **DataFile**.

```go
type DataFile struct {
	path  string
	bTree *bTree.BTree
	fp    *os.File
}

```

Com esta estrutura, posso facilmente criar ou carregar um DataFile, que é um arquivo que contém Dados, no nosso caso atualmente a árvore binária. Para possibilitar a criação de árvores binárias e seus arquivos com facilidade, resolvi criar uma função **OpenDataFile**, cujas responsabilidades são: criar um arquivo de árvore binária caso ele não exista; carregue a árvore binária que existe; carregue todas as funções de retorno de chamada para o arquivo (este pode ser um sistema complexo posteriormente) e carregue o ponteiro do arquivo em fp.

Embora seja possível utilizar separadamente todos os métodos para a função da árvore binária carregando seu módulo, quero utilizar a interface DataFile para todas as operações possíveis com um DataFile específico, desta forma também implementarei todas as funções CRUD nesta struct, como segue:

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

As funções de retorno de chamada são funções anônimas escritas especificamente com o propósito de lidar com o ponteiro de arquivo (fp) relacionado ao arquivo binário da árvore. Foi criada uma função para inicializá-los e colocar todos os callbacks em suas respectivas variáveis ​​do ponteiro bTree.

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

Esses retornos de chamada acima não são definitivos. Num futuro próximo, pretendo criar outro módulo responsável por lidar com solicitações de acesso a arquivos para obter e modificar arquivos de dados, permitindo que muitos threads diferentes rodem em paralelo no mesmo arquivo. Para o propósito atual, basta utilizar as funções como estão.

# Próximos passos

Este foi um passo importante para avançar em meus projetos, mas não estava diretamente relacionado a estruturas de tabelas e assim por diante. Agora que tenho uma interface para acessar e utilizar facilmente a árvore binária, armazenada em pastas de arquivos, irei seguir as divisões da tabela.