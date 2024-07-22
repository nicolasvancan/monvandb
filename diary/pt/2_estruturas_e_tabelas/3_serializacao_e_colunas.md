# Serialização de colunas

Um eventual problema é como serializar as linhas da tabela. Digamos que temos uma tabela com quatro colunas diferentes, cada uma delas contém dados de um tipo específico, como poderiam ser transformados em bytes?

Minha última experiência com serialização foi bem-sucedida e converti um tipo de estrutura genérica em bytes e também o processo de desserialização. Não sei se funciona com um tipo genérico. Em C, para declarar um tipo genérico usamos um ponteiro para void, em java é o próprio Objeto, em JavaScript é um Json Vazio {}, que é o Objeto do JavaScript, e em Golang o tipo genérico é representado pelo ** interface{}**.

Fiz alguns testes com o tipo genérico, cujo alias útil em Golang é o tipo **any**. Posso usar os dois tipos para representar que uma variável pode ter qualquer tipo.

Muitas dúvidas surgem ao lidar com esse tipo de serialização. Existe uma sobrecarga de bytes ao usar a serialização de estrutura normal? É rápido o suficiente?

Eu queria testar alguns conceitos usando Golang também, então decidi que seria bom implementar minha serialização pessoal para alguns tipos de variáveis.

Em vez de usar o campo **Value** como tipo genérico, escolhi que fosse um array de bytes, e a serialização seria feita por mim. Para cada novo tipo de variável, uma abordagem diferente seria aplicada.


```go
type SerializedColumnValue struct {
	Type  int
	Value []byte
}
```

Por enquanto, eu queria provar esse conceito com os seguintes tipos de variáveis:

```go
// Column represents a column in a table
const (
	COL_TYPE_INT = iota
	COL_TYPE_SMALL_INT
	COL_TYPE_BIG_INT
	COL_TYPE_STRING
	COL_TYPE_FLOAT
	COL_TYPE_DOUBLE
	COL_TYPE_BOOL
	COL_TYPE_TIMESTAMP
	COL_TYPE_BLOB
)
```

O campo **Type** da estrutura SerializedColumnValue representa um dos tipos acima. Eu sei que existem muitos casos e provavelmente usaria uma instrução switch case, mas como Golang nos fornece a estrutura do mapa, por que não usá-la para implementar essa lógica de escolher que tipo processaríamos?

Isso é algo que sempre faço em JavaScript ou mesmo em Python ao gerenciar uma grande instrução switch case. Realmente não sei se é mais rápido (provavelmente não), mas o código continua não tão poluído com muitas instruções condicionais. O que eu queria fazer: para cada tipo de valor, tive que implementar a função de serialização e a função de desserialização, e aqui está o porquê


```go
var typesMap = map[int]map[string]interface{}{
	COL_TYPE_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeInt,
	}
}
```

Observe que criei um mapa com chaves de variável inteira, que é o caso dos nossos tipos de colunas, e o valor pode ser qualquer coisa, até mesmo uma função. Então se eu quiser aumentar as possibilidades deste mapa, bastaria inserir outro case dentro do meu **typesMap**, conforme mostrado abaixo:

```go
var typesMap = map[int]map[string]interface{}{
	COL_TYPE_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeInt,
	},
	COL_TYPE_SMALL_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeSmallInt,
	},
	COL_TYPE_BIG_INT: {
		"serialize":   genericSerialize,
		"deserialize": deserializeBigInt,
	}
}
```

O campo serialize contém a respectiva função para serialização, obviamente, o deserialize, contém a referência da função de desserialização. Sabendo disso, é hora de implementar as funções de serializar e desserializar.

```go
func getFuncForType(t int, fn string) (interface{}, bool) {
	f, ok := typesMap[t][fn]
	return f, ok
}

func (s *SerializedColumnValue) DeserializeValue() interface{} {

	fn, ok := getFuncForType(s.Type, "deserialize")
	if !ok {
		fmt.Println("Falhou pra pegar a função deserialize")
		return nil
	}

	deserializeFn, ok := fn.(func([]byte) interface{})
	if !ok {
		fmt.Println("Falhou cast deserialize")
		return nil
	}

	return deserializeFn(s.Value)
}

func SerializeValue(value interface{}, valType int) *SerializedColumnValue {
	fn, ok := getFuncForType(valType, "serialize")

	if !ok {
		fmt.Println("Falhou pra pegar a função")
		return nil
	}

	serializeFn, ok := fn.(func(interface{}) []byte)
	// print serializeFn type
	if !ok {
		fmt.Println("Falhou cast da função")
		return nil
	}

	return &SerializedColumnValue{
		Type:  valType,
		Value: serializeFn(value),
	}
}
```

O que está acontecendo para que isso funcione corretamente? Quando obtenho um valor de mapa, devo inserir o tipo de variável que desejo serializar/desserializar e também uma string indicando qual função desejo usar (este caso tem apenas duas disponíveis). Fazendo isso, recupero a respectiva função que desejo usar.

Como as funções **getFuncForType** retornam um tipo genérico, devo convertê-lo para o respectivo tipo de função de interface, que pode diferir dependendo da função.

Para as funções de serialização, as entradas são diversas e a saída é uma matriz de bytes. Para desserialização acontece o oposto. Todas as funções que serão inseridas nesta interface deverão estar de acordo com esses tipos, conforme exemplo abaixo:

```go
// Follows the interface correctly
func serializeString(value interface{}) []byte {
	return []byte(value.(string))
}

func deserializeString(value []byte) interface{} {
	return string(value)
}
``` 

## A forma final

Eu me diverti muito entendendo alguns conceitos do Golang por meio do primeiro experimento de serialização e desserialização. Mas tudo isso poderia ser evitado se eu usasse a **interface{}** em vez da matriz de bytes no campo Valor da tabela. Mesmo assim, ainda validei o que queria e aprendi a trabalhar corretamente com tipos genéricos, lançando também serialização.

**Como fica agora com **qualquer** tipo? **

O esquema projetado mudou. Não preciso rastrear o tipo de valor do tipo genérico em Golang. A biblioteca *gob* resolve esse problema muito bem e pode converter até valores **nil** quando necessário. Tornou-se muito mais simples. Criei outro pacote, **table** que contém todos os arquivos relacionados às tabelas e suas definições, junto com alguns utilitários, que serão movidos para outro local futuramente, no qual adicionei o arquivo **serialization.go**, possuindo duas funções: Serialize e Deserialize, funcionando para todos os tipos de structs, conforme mostrado abaixo:

```go
func Deserialize(value []byte, dst interface{}) error {

	// Create a new buffer from the serialized data
	c := bytes.NewBuffer(value)

	// Create a new gob decoder and use it to decode the person struct
	dec := gob.NewDecoder(c)
	if err := dec.Decode(dst); err != nil {
		fmt.Println("Error decoding struct:", err)
		return err
	}

	return nil
}

func Serialize(value interface{}) ([]byte, error) {
	var b bytes.Buffer

	// Create a new gob encoder and use it to encode the person struct
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(value); err != nil {
		fmt.Println("Error encoding struct:", err)
		return nil, err
	}

	// The serialized data can now be found in the buffer
	return b.Bytes(), nil
}

```

Obviamente, também criei um bom caso de teste para essas funções acima, no qual imito uma inserção de linha com diferentes tipos. A struct é chamada **ColumnValue**, possuindo o campo **Value** e também o Col, que é referenciado a qual posição da coluna está relacionada nas configurações da tabela (Isso será abordado após esta parte).

```go
func TestSerializationWithDataFile(t *testing.T) {
	// Create tmp dir path
	tmpDirPath := t.TempDir()

	// Get final temporary path for DataFile file
	dbFilePath := tmpDirPath + string(os.PathSeparator) + "test.db"

	// Create a mock row with 4 columns, each with a different type
	row := make(table.RowValues, 5)

	row[0] = table.ColumnValue{Col: 1, Value: uint8(10)}
	row[1] = table.ColumnValue{Col: 2, Value: "Hello"}
	row[2] = table.ColumnValue{Col: 3, Value: 0.8}
	row[3] = table.ColumnValue{Col: 4, Value: 123456}
	row[4] = table.ColumnValue{Col: 4, Value: nil}

	// Serialize them
	serializedRow, err := table.Serialize(row)

	if err != nil {
		t.Errorf("Error encoding struct: %s", err)
	}

	// Open the DataFile (Create new if does not exist)
	dFile, err := files.OpenDataFile(dbFilePath)

	if err != nil {
		t.Errorf("Error opening data file: %s", err)
	}

	// Insert one key
	dFile.Insert([]byte("1"), serializedRow)
	dFile.ForceSync()

	dFile.Close()

	// Reopen recent closed DataFile
	dFile, err = files.OpenDataFile(dbFilePath)

	if err != nil {
		t.Errorf("Error opening data file: %s", err)
	}

	// Create new variable to hold value returned from DataFile
	serializedRow2 := dFile.Get([]byte("1"))

	if serializedRow2 == nil {
		t.Errorf("Error getting data from data file")
	}

	var row2 table.RowValues
	// Deserialize it
	err = table.Deserialize(serializedRow2[0].Value, &row2)

	if err != nil {
		t.Errorf("Error decoding struct: %s", err)
	}

	if len(row) != len(row2) {
		t.Errorf(error_v_string, row, row2)
	}

	for i := 0; i < len(row); i++ {
		if row[i].Col != row2[i].Col || row[i].Value != row2[i].Value {
			t.Errorf(error_v_string, row, row2)
		}
	}
}
```

Olha que legal isso. Eu crio um bTree para armazenar a linha. Eu serializo a linha e insiro o valor da chave na árvore binária para key = '1'. Depois disso, fecho o arquivo binário da Árvore e o reabro, leio o valor armazenado para key = '1', desserializo-o e verifico se consegui ou não armazenar e recuperar o valor mantendo sua integridade. Funcionou.

# Conclusão

O experimento correu bem, consegui serializar e desserializar efetivamente todos os tipos de valores listados em iota enumerate, embora eu ainda use as funções de serialização **gob** para o tipo struct, pude validar alguns conceitos que podem ser usados ​​para mim em os próximos capítulos. 

Acho que valerá a pena todo o trabalho duro de serializar e desserializar usando esse tipo de esquema, somente se o consumo de tempo diminuir extremamente, o que ainda não é o caso. Vou persistir no uso de estruturas de serialização com meus dados e apenas refatorar isso para obter desempenho no futuro.

