# Nós de árvore

O primeiro dia de implementação foi um dia comum. Meu trabalho diário não foi cansativo e resolvi começar a moldar o início do BTree. O conceito de árvore binária é uma estrutura de dados baseada em nós e folhas. Sei que existem vários tipos de árvores mas vou implementar a B+Tree, cujos nós são responsáveis por armazenar outros nós ou posições de folhas, enquanto as folhas armazenam Key-Values (Bytes).

Preparar uma estrutura de bytes para ser salva em uma página de arquivo é algo em que tive que pensar. Normalmente me faço algumas perguntas antes de escrever algo, e as perguntas foram:

- A estrutura necessitará de mais de um tipo de disposição de bytes?
- Eles vão compartilhar algum tipo de informação?
- Que tipo de informação eu preciso?

Respondendo a primeira pergunta: **Sim!** como temos nós e folhas, temos que criar ambas as estruturas. Compartilhar informações seria bom, ambos armazenam algo, seja endereços de nós ou valores reais, então basicamente podemos armazenar informações relacionadas aos itens armazenados, talvez a quantidade de itens na página? Ou melhor ainda, podemos armazenar quanto espaço resta em uma página, ou mesmo qual endereço pertence ao nó pai.

Acho que começar com essas informações básicas para ambos é o suficiente.

A estrutura básica (funcional) de bytes não é uma tarefa difícil, trabalhar com IoT me deu uma boa ideia de como construir protocolos de dados para comunicação sem fio, construindo estruturas de dados a partir de matrizes de bytes ou **estruturas**. A má notícia para mim agora é que tudo foi escrito em C naquela época, não em Golang.

Para quem tem experiência em codificação em C, o que desejo construir para a estrutura básica de dados é algo semelhante ao seguinte código C:

```C
#include <stdint.h>

struct Node {
    uint8_t type; // Node or Leaf
    uint8_t n_items; // Number of items stored
    uint16_t free_bytes; 
    uint64_t p_parent; // Pointer to parent address
}
```
A estrutura acima representa o cabeçalho básico para um Nó ou Folha, vamos chamá-los de Nó Nó e Nó Folha. Implementar algo semelhante em Golang seria:

```go
type Node struct {
    nodeType  uint8; // Node or Leaf
    nItems    uint8; // Number of items stored
    freeBytes uint16; 
    pParent   uint64 ; // Pointer to parent address
}
```

Golang também fornece tipos de struct, com algumas diferenças. Em C, a serialização de estrutura ocorre diretamente. Digamos que eu use malloc para alocar um tamanho de memória de página, neste caso 4.096 bytes. Quando eu converto para um tipo de struct, ele usa automaticamente a quantidade total de bytes que a struct precisa, e o restante fica livre para outra finalidade.

Em Golang também é possível serializar uma struct, mas não é tão direto como em C. Sabendo disso, e considerando que converter bytes para outro tipo de dados para convertê-los novamente e, depois, fazer algumas operações, ficaria lento para grandes operações de dados, decidi construir meu próprio processo de serialização para meus nós e folhas.

Pesquisei e vi que Golang possui duas bibliotecas utilizadas para lidar com arrays de bytes, Binary e Bytes, contendo funções para trabalhar diretamente com bytes e informações binárias.

## Golang Struct

Cada nó é uma página e uma página corresponde a uma matriz de bytes. Portanto, para construir nossa estrutura, precisamos preencher um array de bytes. Achei que seria uma boa ideia ter uma estrutura como esta:

```go
type TreeNode struct {
    data []byte
}
```

Isso significa que, quando serializo dados, apenas pego o campo de dados que contém todos os bytes do nó. Pessoalmente, não gosto de fazer assim, mas não tenho certeza se existe outra opção que seja tão rápida quanto esta.

A desvantagem desse tipo de implementação é que precisamos construir todos os getters e setters para todas as partes específicas da nossa página, o que traz mais complexidade à minha solução. Continuarei com esta solução de qualquer maneira, vamos ver o que consigo.

# O nó nó

Não me inspirei hoje para criar algo fabuloso, então vamos tentar fazer o básico. O nó Node deve ter um cabeçalho e também uma estrutura para salvar dados de nós ou folhas. Btree normalmente requer chaves que são usadas como índices para pesquisas. Que tipo de chave o bTree terá?

Normalmente o que vejo são colunas com números inteiros não negativos que aumentam automaticamente com o tempo, acho que a primeira abordagem seria usar um número inteiro como chave, não quero complexidade agora. A chave pode se tornar enorme, então presumo que seja o maior número disponível na linguagem: **uint64**.

Para cada chave armazenamos valor, mas no caso do Node Node, armazenamos um endereço, não um endereço de memória, mas um endereço de página, que também pode ser enorme, então assumimos que o endereço da página também é **uint64**.

Então, o primeiro passo é criar a estrutura básica dos nossos nós:

```go
/* Base Node */
type TreeNode struct {
	// This holds bytes to be dumped to the disk
	data []byte
}
```

Também é necessário criar um tipo enum para diferenciar o tipo de nó, usando iota

```go
/* TreeNode implementation */
const (
	TREE_NODE = iota
	TREE_LEAF
)
```

É hora de pensar e construir a estrutura básica do nó. Achei que seria bom assim mostrado abaixo:

- **type**: Tipo de Btree TREE_NODE ou TREE_LEAF **uint16**
- **nItems**: Indica quantos itens o nó possui **uint16**
- **freeBytes**: Indica quantos bytes livres o nó possui **uint16**
- **pParent**: Página referente ao nó pai (Caso precisemos) **uint64**
- **n * NodeStructure**: Esta é apenas uma representação para explicar que após os bytes pParent existem apenas dados relacionados a outros nós e endereços **Pode ter muitos bytes**
- **Estrutura do Nó**:
- - **key**: chave mais baixa da página referenciada **uint64**
- - **addr**: Endereço da página **uint64**

Criei uma imagem para mostrar um exemplo de como o endereço chave **NodeStructure** funcionaria de forma prática:

![Node Diagram](../../assets/node_diagram.png)

Existe o cabeçalho, ele é composto pelos campos: tipo, nItems, etc; seguido por espaço livre usado para armazenar informações relacionadas aos principais endereços. Sempre que uma nova estrutura de endereço de chave é adicionada à página, o valor é concatenado à página após os cabeçalhos, a contagem de nItems é aumentada em um e o número de bytes livres diminui.

A mesma ideia se aplicaria aos pares de valores-chave, diferindo apenas pelo fato de que o número de bytes armazenados no valor é variável, portanto, precisamos de informações sobre quantos bytes o valor consiste.

## Primeira implementação

Começando pelos cabeçalhos tive que transformar aquele cabeçalho planejado em algo útil. Como queremos construir nossa própria serialização por meio das informações da estrutura, devemos começar escrevendo getters e setters (quem programa em Java adora isso).

Para esta tarefa utilizarei o pacote **binary**, que nos permite trabalhar com bytes e escrever informações seja em little endian ou big endian. Usarei little endian, mas nada impede que alguém use big endian.

Antes de começar a escrever qualquer getter ou setter, decidi criar algumas variáveis de macros, que são escritas com a instrução **const**, especificando o comprimento de cada informação em bytes, e também a posição em que essas informações começam nas matrizes de bytes.

```go
// Example of const declaration
const (
	NODE_TYPE_LEN            = 2 // Length of node type field in bytes
	NODE_OFFSET_LEN          = 2 // Length of offset field in bytes
	NODE_N_ITENS_LEN         = 2 // Length of field n Itens in bytes
	NODE_FREE_BYTES_LEN      = 2 // Length of node free bytes field in bytes
	NODE_PARENT_ADDR         = 8 // Length of node parent address field in bytes
	NODE_P_KEY_LEN           = 8 // Length of node key len
	NODE_P_CHILD_ADD_LEN     = 8 // Length of node children address in bytes
	NODE_P_KEY_ADDRESS_LEN   = NODE_P_KEY_LEN + NODE_P_CHILD_ADD_LEN // Length key address len for node
	LEAF_HAS_SEQ_LEN         = 2 // Length of leaf has sequence field
	LEAF_SEQ_P_LEN           = 8 // Length of leaf sequence address field
	LEAF_KEY_LEN_LEN         = 2 // Length of leaf key lenght field in bytes
	LEAF_VAL_LEN_LEN         = 8 // Length of leaf val lenght field in bytes
	LEAF_SEQ_N_BYTES         = 2 // Length number of bytes in leaf sequence
	LEAF_SEQ_FREE_BYTES_SIZE = PAGE_SIZE - LEAF_SEQ_BYTES_OFFSET // Length of sequences free bytes size
)
```

Eu sei que os nomes possivelmente não são tão fáceis de entender nem muito bonitos, mas foi assim que os escrevi (se você quiser ver todos eles, basta encontrar o arquivo bTreeNode.go no módulo btree)

**Escrevendo getters e setters**

Nunca escrevi alguns getters e setters em golang antes. Já fiz isso em diferentes linguagens, C, Java, JS, Python, mas em Golang é algo realmente novo. Eu li que é possível escrevê-los vinculados a alguma estrutura, ou seja, métodos de classe, ou escrever passando algum endereço de estrutura como parâmetro de função, e modificar o que você deseja dentro da função.

Eu vou ser honesto. Eu misturei os dois tentando diferir algumas características da folha do nó e do nó, mas isso acabou sendo uma confusão e uma forma não padronizada de escrever código. Consertar isso é uma tarefa para a posteridade. Mas vamos codar agora

```go
func (n *TreeNode) GetType() uint16 {
	return uint16(binary.LittleEndian.Uint16(n.data[NODE_TYPE_OFFSET : NODE_TYPE_OFFSET+NODE_TYPE_LEN]))
}

func setType(n *TreeNode, nType uint16) {
	binary.LittleEndian.PutUint16(n.data[NODE_TYPE_OFFSET:NODE_TYPE_OFFSET+NODE_TYPE_LEN], nType)
}

// Example of use
var node TreeNode
type_of_node := node.GetType()

setType(&node, 1)
```

Aqui estão dois exemplos de como você pode usar métodos com estruturas. Tentei diferenciá-las, coloquei as funções públicas com os métodos struct e as funções privadas separadas, embora nem tudo siga esse padrão.

**A função NewNode**

Queria ter também uma função que me retornasse um Node Node vazio para ser usado na minha aplicação, seria o mesmo que um Construtor para minha estrutura. Portanto, a criação de uma função para isso me pareceu uma boa ideia, o problema é que eu não sabia como inicializar uma struct contendo array de bytes como propriedade. Depois de alguns testes, descobri que existem algumas funções integradas que ajudam você a criar facilmente um objeto a partir de um tipo, são elas: funções **make** e **new**. Make é útil ao criar arrays de tipos, no meu caso, um array de bytes. Considerando que a nova função é usada para criar estruturas vazias para um determinado tipo. Além disso, você pode criar manualmente a estrutura, como em C, como mostrado em meu exemplo, onde combinei a criação de estrutura make e formulário C para criar uma nova estrutura TreeNode.


```go
func NewNodeNode() *TreeNode {
	// Create a pointer to a new Node Structure
	nodeNode := &TreeNode{data: make([]byte, PAGE_SIZE)}
	// Set headers
	setType(nodeNode, TREE_NODE)
	setNItens(nodeNode, 0)
	setFreeBytes(nodeNode, PAGE_SIZE-NODE_FREE_BYTES_LEN-NODE_TYPE_LEN-NODE_N_ITENS_LEN-NODE_PARENT_ADDR)
	return nodeNode
}

```

## Avanço - Tipos de nós e métodos de exclusão de inserção

Getters e Setters foram implementados usando métodos *binary.LittleEndian* para trabalhar com array de bytes. Agora é hora de implementar funções que nos permitam inserir, excluir, atualizar e obter valores-chave e endereços de nossas páginas Node. Devo confessar que estou com muita preguiça e não quero perder tanto tempo nessa tarefa nem desenvolver uma lógica complexa para isso.

Acho que a única restrição que tenho é manter todos os valores das páginas ordenados, ou seja, armazená-los ordenados (Desired), mas recuperá-los ordenados é suficiente.

Antes de avançar mais no assunto, tive que criar alguns tipos de struct concretos para demonstrar quais são os tipos e informações armazenadas dentro das páginas, que são: **NodeKeyAddr**, **LeafKeyValue**, conforme mostrado abaixo:


```go
type LeafKeyValue struct {
	keyLength   uint16
	valueLength uint64
	key         []byte
	value       []byte
}

type NodeKeyAddr struct {
	keyLen uint16
	key    []byte
	addr   uint64
}

```

Resolvi mudar a chave de uint64 para array de bytes, considerando qualquer possibilidade de utilização de chaves indexadas, como string, bytes diferentes, objetos, entre outros, mudar o campo para array de bytes foi a melhor opção que pude escolher.

Para cada tipo de nó é criada uma função para inserir, excluir e atualizar valores. Na verdade, o método de atualização não será implementado, atualizar é o mesmo que excluir e inserir a mesma chave.

**Insert**

```go
func (n *TreeNode) PutNodeNewChild(key []byte, addr uint64) error {

	// Verify whether it will exceed total bytes
	aditionalLength := len(key) + 2 + 8
	if int(GetFreeBytes(n))-(aditionalLength) < 0 {
		return errors.New("exceeds total bytes")
	}
	keyLen := uint16(len(key))
	// takes offset
	offset := getNodeOffset(n)

	/*
		2B - Len of key
		Len of Key B - Key
		8B - Address
		Example:
		key = ["a","t","o","m","i","c"]
		addr = 157

		keyLen = 6 - Therefore the size will be 2B + 6B + 8B = 16B
	*/

	// Write len 2B
	binary.LittleEndian.PutUint16(n.data[offset:offset+2], keyLen)
	// Write Key (variable)
	copy(n.data[offset+2:offset+2+keyLen], key)
	// Write Address 8B
	binary.LittleEndian.PutUint64(n.data[offset+2+keyLen:offset+2+keyLen+8], addr)

	// Set new offset
	setNodeOffset(n, offset+2+keyLen+8)
	// Set new Free Bytes
	setFreeBytes(n, GetFreeBytes(n)-(2+8+keyLen))
	// Set NItems
	setNItens(n, n.GetNItens()+1)

	return nil
}

func (n *TreeNode) PutLeafNewKeyValue(key []byte, value []byte) error {
	aditionalLength := len(key) + 2 + 8 + len(value)

	if int(GetFreeBytes(n))-(aditionalLength) < 0 {
		return errors.New("exceeds total bytes")
	}
	keyLen := uint16(len(key))
	valLen := uint64(len(value))
	// takes offset
	offset := getNodeOffset(n)
	/*
		2B - Len of key
		8B - Len of value
		Len of Key B - Key
		8B - Value
		Example:
		key = ["a","t","o","m","i","c"]
		value = []byte("some value inserted in here")

		keyLen = 6 - Therefore the size will be 2B + 8B + 6B + 27B = 43B
	*/

	// Write keylen 2B
	binary.LittleEndian.PutUint16(n.data[offset:offset+2], keyLen)
	// Write valuelen 8B
	binary.LittleEndian.PutUint64(n.data[offset+2:offset+2+8], valLen)
	// Write Key (variable)
	copy(n.data[offset+10:offset+10+keyLen], key)
	// Write Address 8B
	copy(n.data[offset+10+keyLen:offset+10+keyLen+uint16(valLen)], value)

	// Set new offset
	setNodeOffset(n, offset+10+keyLen+uint16(valLen))
	// Set new Free Bytes
	setFreeBytes(n, GetFreeBytes(n)-(10+keyLen+uint16(valLen)))
	// Set NItems
	setNItens(n, n.GetNItens()+1)

	return nil
}
```

A ideia é simples. Para facilitar minha vida, criei outro campo chamado offset, que indica qual a posição do primeiro byte livre para ser a referência para novos bytes de entrada serem salvos. A diferença entre ambos é que o valor len de um LeafKeyValue é variável, portanto, possui um campo a mais referente ao comprimento dos bytes a serem salvos.

A cada novo item a ser adicionado, o sistema analisa o deslocamento e então armazena os dados após o deslocamento, atualiza a quantidade de itens na página e também atualiza o valor do deslocamento para o novo. Fazendo isso, não preciso fazer muitas contas e nem preciso criar uma solução lógica milagrosa para isso. Mas você pode me perguntar: **Como você garante que está ordenado?**

Na verdade não garanto nada ao inserir dados, apenas que os dados estão lá. Os dados ordenados vêm durante a leitura da página, conforme mostrado abaixo para métodos get:

```go
func getAllLeafKeyValues(n *TreeNode) []LeafKeyValue {
	nItems := n.GetNItens()
	r := make([]LeafKeyValue, nItems)
	baseOffset := LEAF_VAL_START_OFFSET
	for i := 0; i < int(nItems); i++ {
		kLen := binary.LittleEndian.Uint16(n.data[baseOffset : baseOffset+2])
		vLen := binary.LittleEndian.Uint64(n.data[baseOffset+2 : baseOffset+2+8])
		key := n.data[baseOffset+10 : baseOffset+10+int(kLen)]
		value := n.data[baseOffset+10+int(kLen) : baseOffset+10+int(kLen)+int(vLen)]
		r[i] = LeafKeyValue{
			keyLength:   kLen,
			valueLength: vLen,
			key:         key,
			value:       value,
		}
		// Add values to offset
		baseOffset += (int(kLen)) + int(vLen) + 2 + 8
	}

	sortLeafKeyValues(r)
	return r
}


func getAllNodeKeyAddr(n *TreeNode) []NodeKeyAddr {
	// Get number of items
	nItens := n.GetNItens()
	// Initiate return array
	r := make([]NodeKeyAddr, nItens)
	// Start always at the very beginning
	lastStart := NODE_P_KEY_ADDR_OFFSET
	for i := 0; i < int(nItens); i++ {
		// Get key Length
		kLen := binary.LittleEndian.Uint16(n.data[lastStart : lastStart+2])
		// Get key value in []bytes
		key := n.data[lastStart+2 : lastStart+2+int(kLen)]
		// Get key address
		addr := binary.LittleEndian.Uint64(n.data[lastStart+2+int(kLen) : lastStart+2+int(kLen)+8])
		r[i] = NodeKeyAddr{
			keyLen: kLen,
			key:    key,
			addr:   addr,
		}
		lastStart += 2 + int(kLen) + 8
	}

	sortNodeChildren(r)
	return r
}
```

A ideia de obter todos os dados é a mesma de armazenar, exceto pelo fato de ser feito de forma inversa, retornando um array do respectivo tipo, seja retornando um array de LeafKeyValues, ou NodeKeyAddress.

Depois de obter arrays em golang, podemos aplicar facilmente o método sort

```go
func sortNodeChildren(c []NodeKeyAddr) {
	sort.Slice(c, func(i, j int) bool {
		return bytes.Compare(c[i].key, c[j].key) <= 0
	})
}
```

Tive que criar o mesmo método para LeafKeyValue, e o código está ficando duplicado em algumas partes, o que não gosto, mas não é motivo de preocupação no momento.

**Excluir funções**

Excluir parte específica da matriz de bytes, especialmente quando não é um byte de comprimento fixo, pode ser uma tarefa difícil. Eu sei que esta implementação feita aqui não é algo especial e não é rápida como eu quero, mas é útil avançar com a primeira versão do meu bTree.

O que faço é não deletar nada, leio todos os valores de uma página (Node), removo a chave desejada, e depois armazeno tudo novamente no array renovado. Simples mas lento, uma das melhores alternativas é identificar a posição real do campo desejado na página, e copiar outros bytes para a sua posição, nomeadamente deslocando os bytes da direita para a esquerda.

Um exemplo do método implementado é mostrado abaixo:

```go
func (n *TreeNode) DeleteLeafKeyValueByKey(key []byte) {
	allLeafKeyValues := getAllLeafKeyValues(n)
	// Reset Node
	tmp := NewNodeLeaf()
	setParentAddr(tmp, n.GetParentAddr())
	setLeafHasSeq(tmp, n.GetLeafHasSeq())
	setLeafSeqPointer(tmp, n.GetLeafSeqPointer())
	for i := 0; i < len(allLeafKeyValues); i++ {
		if bytes.Equal(allLeafKeyValues[i].key, key) {
			continue
		}
		tmp.PutLeafNewKeyValue(allLeafKeyValues[i].key, allLeafKeyValues[i].value)
	}
	copy(n.data, tmp.data)
}
```

# Hora de seguir em frente

Depois de implementar tudo que cansei e precisava para avançar nesse assunto, quis começar a criar as funções crud da árvore binária, e farei isso. Percebi que poderia mesclar as duas estruturas **LeafKeyValue** e **NodeKeyAddress**, remover métodos duplicados para inserir, excluir, classificar e assim por diante. Farei isso mais tarde, para que seja mantido em um arquivo TODO.

Implementar o método delete diretamente nos bytes será muito mais rápido e melhorará a velocidade do meu software, embora eu deixe essa tarefa para mim no futuro.

As últimas considerações para este pacote é que vou escrevê-lo novamente usando algum padrão para nomenclatura e deixando-o menos detalhado, e a constante PAGE_SIZE poderia ser alterada para funcionar como uma PAGE_SIZE para cada arquivo bTree, esta informação seria armazenada no primeiro página do arquivo.