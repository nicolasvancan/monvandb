# Serialização de objetos

Hoje quero validar o conceito de serialização de objetos (Structs) e suas possibilidades em relação à estrutura de dados e como funcionaria se eu tivesse uma estrutura de dados complexa com definições de colunas. O problema é que não sei como isso pode funcionar. Portanto, quero pesquisar um pouco sobre esse assunto e combinar os resultados da pesquisa com alguns pequenos testes de código.

## Por que estou pesquisando serialização?

A resposta a esta pergunta é simples. Como posso armazenar um array de bytes na árvore binária, quero testar se consigo serializar uma estrutura complexa, cujos campos seriam análogos a uma tabela, contendo colunas diferentes e inserí-lo no arquivo. Um dos problemas é a própria serialização, o outro é a estrutura que irá conter valores referentes a alguma estrutura específica da tabela.

Quando um usuário escreve uma instrução SQL de criação de tabela, ele normalmente usa o seguinte:

```SQL
-- Example of SQL STATEMENT
CREATE TABLE IF NOT EXISTS some_table (
    id INT PRIMARY KEY,
    createdAt DATETIME,
    name CHAR(100)
)
```

Observe que a composição de uma tabela é baseada nas colunas que ela pode armazenar e seus tipos, além de outras possíveis restrições (chave composta). Essa é a definição de uma mesa. Mas cada linha desta tabela declarada será armazenada dentro de um arquivo de árvore binária como uma matriz de bytes, e o ponto chave aqui é saber como serializá-la e desserializá-la para ler e gravar dados da tabela.

Outro aspecto de uma serialização de linhas é: E se eu quiser obter apenas uma coluna dos dados, vamos usar a tabela anterior como nosso exemplo, se quisermos obter a coluna **nome** da tabela **alguma_tabela** , escreveríamos alguma consulta, digamos

```SQL
SELECT name
FROM some_table
``` 

O banco de dados pediria à tabela para ler os dados do arquivo da árvore binária e obter apenas o campo de nome. Se serializarmos toda a estrutura, por exemplo, uma estrutura com os mesmos campos.

```go
type some_table struct {
    id int
    createdAt uint64
    name string
}
```

Na leitura dos dados, os dados recuperados seriam toda a estrutura. E pensando na eficiência no armazenamento e recuperação dos dados, obter apenas o que é necessário seria a melhor opção, mas não a mais fácil. Não tenho certeza se minha primeira implementação conterá a obtenção de colunas específicas diretamente do arquivo de dados. O caminho direto é obter todos os dados das colunas e filtrar o que deseja depois, embora possa trazer para a memória alguns dados indesejados que também podem ser enormes.

**Outro problema – Tabelas são variáveis**

Sim, tabelas são variáveis, ou seja, quando o usuário define uma tabela sendo composta por três colunas com tipos de dados diferentes, não podemos criar uma definição de struct em tempo de execução, devemos montar essa tabela com algum tipo de estrutura embutida, e isso é o ponto chave disso. De qualquer forma, se essa estrutura se tornar enorme, uma única linha teria tantos bytes usados ​​apenas para armazenar uma pequena porção de informações. Portanto, a prova de conceito e avaliação do uso de bytes da serialização do pacote golang e da minha própria serialização será de grande valor.

## O Pacote Gob

Pesquisei um pouco e encontrei algumas opções para resolver meu problema de serialização. Um deles é o pacote encoding/gob, que possui métodos de serialização para estruturas, que é exatamente o que procuro.

Como meu primeiro teste, tentarei serializar (converter uma struct em array de bytes), para isso, criarei uma struct chamada TableTest com os campos **id**, **name**,**city**,**número**

```go
type TableTest struct {
    id int
    name string
    city string
    number int
}
```

A principal função originada para fins de teste é mostrada abaixo:

```go
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type TableTest struct {
	Id     int
	Name   string
	City   string
	Number int
}

func main() {
	// Create an instance of the Person struct
	row := TableTest{
		Id:     1,
		Name:   "Nicolas",
		City:   "Paris",
		Number: 123456,
	}

	// Create a new buffer to write the serialized data to
	var b bytes.Buffer

	// Create a new gob encoder and use it to encode the person struct
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(row); err != nil {
		fmt.Println("Error encoding struct:", err)
		return
	}

	// The serialized data can now be found in the buffer
	serializedData := b.Bytes()
	fmt.Printf("Serialized data: %s and its length = %d\n", serializedData, len(b.Bytes()))
}

```

A saída obtida foi a seguinte:

![texto alternativo](../../assets/serialize_1.png)

Seria possível serializar um array da minha struct? Vamos testar. Em vez de usar row como uma estrutura, modificarei seu nome e conteúdo para caber em uma matriz de estruturas TestTable com apenas uma estrutura nela.

```go
// Create an instance of the Person struct
    rows := make([]TableTest, 1)
	rows[0] = TableTest{
		Id:     1,
		Name:   "Nicolas",
		City:   "Paris",
		Number: 123456,
	}
	rows[1] = TableTest{
		Id:     2,
		Name:   "John",
		City:   "London",
		Number: 789456,
	}
	// Create a new buffer to write the serialized data to
	var b bytes.Buffer

	// Create a new gob encoder and use it to encode the person struct
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(row); err != nil {
		fmt.Println("Error encoding struct:", err)
		return
	}

	// The serialized data can now be found in the buffer
	serializedData := b.Bytes()
	fmt.Printf("Serialized data: %s and its length = %d\n", serializedData, len(b.Bytes()))
}

```

Como resultado, é possível serializar também um array do tipo struct, mostrando que esse tipo de serialização é realmente versátil, embora o uso de bytes seja maior do que se fosse serializado diretamente, e isso pode ser facilmente contado. Temos quatro campos, o primeiro é um inteiro, que tem 32 bits ou 4 bytes de comprimento, temos dois dessa variável, e duas strings, somando o que seriam 20 bytes de dados brutos, com o overhead da estrutura básica ele torna-se 86 para um item e, se for uma matriz de apenas um item, torna-se 124. 

Para desserializar, o oposto deve ser feito. Para validar esse processo, adicionei outra parte do código após a serialização referente à desserialização.


```go
// Create a new buffer from the serialized data
	c := bytes.NewBuffer(serializedData)

	// Create a new gob decoder and use it to decode the person struct
	var deserialized []TableTest
	dec := gob.NewDecoder(c)
	if err := dec.Decode(&deserialized); err != nil {
		fmt.Println("Error decoding struct:", err)
		return
	}

	// The person struct has now been deserialized
	fmt.Println("Deserialized struct:", deserialized)

```

O resultado é exatamente o que serializei antes, conforme mostrado abaixo:

![texto alternativo](../../assets/serialized_2.png)

Mesmo que esse processo de serialização exija mais bytes do que seria necessário se eu inserisse cada campo separadamente, ele pode representar exatamente um Go Struct, que fornece muito mais flexibilidade ao trabalhar com dados estruturados e não estruturados. Que tipo de estrutura devo construir para garantir flexibilidade, reutilização e uma solução não complexa? Isso será feito no próximo diário.

**PS: Decidi criar em cada capítulo deste projeto uma pasta chamada arquivos, que contém todos os arquivos .go relacionados a experimentos que eu possa tentar ou realmente faço para entender alguns comportamentos desejados**