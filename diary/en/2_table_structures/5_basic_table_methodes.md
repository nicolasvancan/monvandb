
# Table methods

Defined structures alone do not guarantee success in reading information. Functions need to be implemented to use the fields I defined in the previous chapter. As I mentioned, the table is the interface closest to the raw data. In other words, it is the one that accesses the *DataFile* files and retrieves both the keys and the values ​​saved there; and it is the one that has the information saved in its structure used to decode the retrieved information. So what are the types of table methods?

I decided to divide them into three distinct parts for now:

1. Converters
2. Validation
3. CRUD

All methods use a specific data type to work, the RawRow.

**RawRow**

I had a lot of contact with *pandas* and *spark* in python, and it was common to use the *dict* data type (dictionary) to handle rows and columns, and I find the concept interesting too, even because of its practicality and similarity to json. So I decided to create a basic data type called **RawRow**, whose definition is as follows:

```go
type RawRow = map[string]interface{}

// Example

oneRow := map[string]interface{} {
    "id":1,
    "name": "Nícolas",
    "age": 32,
    "phone" "+555555555555",
    "email":"nicolas@nicolas.com"
}
```

Although practical, this structure is not saved in this way in *DataFile* files; there is an intermediate conversion process so that the data can be serialized and saved.

# Converters

The serialization seen at the beginning of this "season" will be put to good use at this point in the project. Let's recap:

1. *DataFile* files store binary information for both keys and values, and return a **[]BTreeKeyValue** structure when a reading is performed.

2. When we obtain the bytes from the files, we transform them into another intermediate data type, a type that relates a value to a data type and a column.

3. With this intermediate data type, we can then create our final converter for **RawRow** and vice versa.

What is this intermediate type and why did I decide to develop it and not directly store the **RawRow** types as values ​​in the file? Well, this is a question that I spent many days pondering until I reached an acceptable conclusion.

But the main point is that within the files, we work with and store bytes, making this handling more complicated due to possible implementations and limitations of Golang regarding serialization. The *gob* package, used for this process, serializes the entire structure very quickly. However, imagine the following scenario:

I decide to perform the following query on a fictitious table called **table_test**, whose columns are: **id**, **some_string**, **some_bytes**. To make our lives more complicated, one of the lines saved in the file has the **some_bytes** column with a size close to 1 GB. Let's say I want to run a simple query to bring only the **id** and **some_string** fields from all the lines in the table.

```sql
SELECT id, some_string
FROM table_test
```

If our row value has been serialized as a whole, the system has to retrieve all rows with all columns and only then elimit the **some_bytes** column, leading to possible unnecessary memory usage. What if it were possible to select the columns you or the system want from within the binary tree file itself without having to bring all of them into memory? This is the purpose of the **ColumnValue** structure.

```go
type ColumnValue struct {
	Value interface{} // Value of the column
	Col   uint16      // Refers to the respective column of a table for example:
	// Table X has column Y and column Z, which are stored in a Table struct as an array of Columns struct
	// Each position of this array represents a column in the table. Whenever a column is serialized,
	// the position of the column is stored in the Col field
}
```

The idea is simple -- there are two fields, *Value* and *Col*, which are the column value and the column position number of the table structure, respectively. Let's explain further.

Our table has the following structure:

```go
table := Table{
    Columns: []Columns {
        {
            Name:"id",
            Type: COL_INT
        },
        {
            Name:"name",
            Type: COL_STRING
        },
        {
            Name:"email",
            Type: COL_STRING
        }
    }
}

```

Note that the defined array of columns is ordered with the first column in position zero, the second in position one, and so on. These positions represent the *Col* field of the **ColumnValue** data structure. Isn't that easy? But what about converting an entire row into bytes?

Here is an illustrative example of this conversion, using the table type defined previously:

**Row to be converted**

```go
oneRow := map[string]interface{} {
    "id":1,
    "name": "Nícolas",
    "email":"nicolas@nicolas.com"
}
```

Converting this line to the **[]ColumnValue** data type results in:

```go
columnValues := []ColumnValue{
    {
        Value: 1,
        Col:0
    },
    {
        Value:"Nícolas",
        Col:1
    },
    {
        Value:"nicolas@nicolas.com"m
        Col:2
    }
}
```

Finally, all you need to do is serialize this **ColumnValue** array and insert the bytes into the *DataFile* file. However, the serialization method here follows the most critical step. In order to retrieve specific columns from the rows without having to bring the entire row into memory, I need to implement my own serialization process, which would take a bit of work; and right now my goal is not optimization yet. For this reason and others, I decided to use the already created utility function **Serialize** to generate the bytes.

The following are the functions used for this entire conversion process:

```go
// Functions that are used to convert a row to column values and vice versa
func (t *Table) MapRowToColumnValues(row RawRow) []ColumnValue {
	// Create a map to hold the column values
	columnValues := make([]ColumnValue, len(t.Columns))

	// Loop through the columns and add them to the map
	for index := range t.Columns {
		// Get the value from the row
		value := row[t.Columns[index].Name]

		// Create a new column value
		columnValues[index] = ColumnValue{
			Col:   uint16(index),
			Value: value,
		}
	}

	return columnValues
}

func (t *Table) FromColumnValuesToRow(columnValues []ColumnValue) RawRow {
	// Create a map to hold the column values
	row := make(RawRow)

	// Loop through the columns and add them to the map
	for index := range t.Columns {
		// Get the column value
		columnValue := columnValues[index]
		// Add the value to the row
		row[t.Columns[index].Name] = columnValue.Value
	}

	return row
}

func (t *Table) FromRawRowToKeyValue(row RawRow) btree.BTreeKeyValue {
	// Create a new column value
	columnValues := t.MapRowToColumnValues(row)

	// Serialize the column values
	serializedColumnValues, _ := utils.Serialize(columnValues)
	key, _ := utils.Serialize(row[t.PrimaryKey.Name])
	return btree.BTreeKeyValue{
		Key:   key,
		Value: serializedColumnValues,
	}
}

func (t *Table) FromKeyValueToRawRow(keyValues []btree.BTreeKeyValue) []RawRow {
	// Create a new column value
	rawRows := make([]RawRow, len(keyValues))

	for i := range keyValues {
		// Create tmp ColumnValue array
		var columnValue []ColumnValue
		// Get the value from the row
		err := utils.Deserialize(keyValues[i].Value, &columnValue)

		if err != nil {
			return nil
		}

		rawRow := t.FromColumnValuesToRow(columnValue)
		// Add the value to the row
		rawRows[i] = rawRow
	}

	return rawRows
}

```

# Validadores

Não queremos que nossos dados salvos fujam do padrão, não é mesmo? Então é sempre bom garantir que os tipos dos dados estejam nas conformidades das definições das colunas das tabelas. Como por enquanto as necessidades da tabela não são tantas, digo, não há muitas restrições que possam aparecer além dos mais básicos, como: Validação de tipos de dados; Verificação de dados nulos, Verificação de duplicados, Inserção automática de chave primária.

Para suprir essas necessidades, criei um arquivo chamado: **validator.go** dentro do módulo de tabelas, o qual contém todas as funções de validação e alteração de dados, caso necessário.

Começando com o método de validação se a coluna não existe nas definições da tabela. Já pensou no caso onde alguém tenta inserir uma coluna que não existe na definição da tabela e acaba conseguindo? Isso pode causar sérios problemas caso não seja evitado. Por isso, vamos criar nossa primeira função de validação, **validateIfColumnExist**

```go
func validateIfColumnExist(column Column, t *Table) error {
	if t.GetColumnByName(column.Name) == nil {
		return fmt.Errorf("column %s does not exist in table %s", column.Name, t.Name)
	}

	return nil
}
```

Ela é simples e verifica se nas definição da estrutura **Table**, no campo **Columns**, existe alguma coluna com o nome da variável *column* (parâmetro da função). Caso essa coluna recebida pela função não exista nas definições da tabela, o sistema retorna um erro.

Passado isso, tem o caso inverso, no qual menos colunas do que as totais existentes na tabela são fornecidas. Isso é comum e pode ocorrer, pois o próprio servidor do banco de dados encarrega-se de preencher as colunas faltantes com campos nulos e, no caso dos campos auto increment, ele trata de incrementar automaticamente caso seja uma chave da tabela.

```go
func fillupMissingFields(t *Table, row *RawRow, column Column) {
	// Fill up missing fields with default values
	if _, ok := (*row)[column.Name]; !ok {
		if !column.AutoIncrement && !column.Primary && !ok {
			(*row)[column.Name] = column.Default
			return
		}
	}

	// Get last Value, cast it to int64 and add one (This should be only with autoincrement columns)
	// Meaning that the columns is an integer between all possiblities
	// SMALL_INT, BIG_INT, INT, etc
	if column.AutoIncrement && column.Primary {
		lastValue := t.getLastItem()
		if lastValue == nil {
			(*row)[column.Name] = lastValue[column.Name].(int64) + 1
		} else {
			(*row)[column.Name] = 1
		}
	}
}

```

Olhem que interessante, na primeira parte da função eu verifico se a linha que vai ser inserida tem a coluna que estou validando. Caso ela não exista, o sistema insere o valor da coluna **Default**, que por definição é *nil*.

Na segunda parte tratamos os casos das colunas que não existem e são chaves primárias auto incrementais. Neste caso, buscamos o último item adicionado à tabela e recuperamos sua respectiva chave, fazemos um cast nesse valor para **int64** e adicionamos um ao número. Se a tabela for vazia significa que o campo de auto increment deve ser um.

Notem que eu considero que obrigatóriamente a coluna auto increment seja um inteiro, e realmente deve ser. Irei criar validadores nos comandos de criar tabela para impedir com que colunas auto increment sejam criadas quando tiverem tipos de dados diferentes de inteiros.

**Validar Não Nulos**

```go
func validateNull(column Column, value interface{}) error {
	if !column.Nullable && value == nil {
		return fmt.Errorf("column %s cannot be null", column.Name)
	}

	return nil
}
```

Aqui é simples e não tem segredo, verificamos se o campo é anulável e se não for, verificamos se o valor é *nil*. Caso as condições sejam satisfeitas, retornamos um erro na função.

**Validar Duplicidade**

Chegou uma função mais complicada para verificações, se o dados já existe ou não. Normalmente, dados duplicados são verificados na inserção da chave. Por exemplo, em tabelas com chaves únicas, não podemos inserir duas vezes uma linha com a chave igual a 1. O mesmo ocorre para tabelas que possuem chaves compostas e uma restrição criada para essas tabelas. Normalmente vê-se essas tabelas com as **constraints**. Bom, vamos lá, a implementação inicial é a seguinte:

```go
func validateUnique(table *Table, row RawRow) error {

	// It is a table without any constraints
	if table.PrimaryKey == nil && table.CompositeKey == nil {
		return fmt.Errorf("not indexed table. Cannot validate uniqueness")
	}

	// Uniqueness works only for primary keys or composite keys (Constraints)
	var pk *Column = nil
	var key []byte = make([]byte, 0)
	if table.PrimaryKey != nil {
		pk = table.PrimaryKey
		// Indicates that the PK is present and not composite key
		tmp := row[pk.Name]
		serialized, err := utils.Serialize(tmp)

		if err != nil {
			return err
		}

		key = append(key, serialized...)
	}

	if table.CompositeKey != nil {
		// If composite key is different than nil, then it is a composite key
		for _, column := range table.CompositeKey {
			tmp := row[column.Name]
			serialized, err := utils.Serialize(tmp)

			if err != nil {
				return err
			}

			key = append(key, serialized...)
		}
	}

	if len(table.PDataFile.Get(key)) > 0 {
		return fmt.Errorf("row already exists in table")
	}

	return nil
}
```

Nada muito complicado. Existem duas possibilidades, quando a chave é única ou a chave é composta, indicando que a restrição é baseada nessa chave e no arquivo principal da tabela. Após conseguirmos o valor da chave serializado, utilizamos diretamente a interface do ponteiro do arquivo **DataFile** para ver se essa chave existe ou não. Caso exista, o sistema emite um erro indicando que uma chave desse valor já fora adicionado.

**Função Única**

Bom, encapsulei todas essas validações em uma função única que roda antes das inserções de dados e recebe como parâmetro um array de **RawRow**, como mostrado abaixo:

```go
func (t *Table) ValidateRawRows(rows []RawRow) ([]RawRow, error) {
	validatedRows := make([]RawRow, 0)
	for _, row := range rows {
		err := t.ValidateColumns(&row)
		if err != nil {
			return nil, err
		}
		validatedRows = append(validatedRows, row)
	}
	return validatedRows, nil
}

func (t *Table) ValidateColumns(row *RawRow) error {
	columns := t.Columns
	for _, column := range columns {
		fillupMissingFields(t, row, column)
		err := validateIfColumnExist(column, t)
		if err != nil {
			return err
		}
		err = validateNull(column, (*row)[column.Name])

		if err != nil {
			return err
		}

		err = validateUnique(t, *row)

		if err != nil {
			return err
		}
	}

	return nil
}
```

# CRUD Tabelas

Cheamos na parte final deste capítulo, finalmente. Passei tanto tempo pensando nessa parte de estruturas e como deveria ser que, agora que estou perto de finalizar suas implementações iniciais, estou me sentindo aliviado. Bom, de todo modo, as tabelas precisam adicionar, ler, deletar e atualizar infomações, certo? Todas terão os seguintes métodos básicos:

- Get
- Insert
- Update
- Delete

Qual diferença de implementação destes inserts da implementação original do **DataFile**? Os inputs recebidos pelos métodos são todos baseados em **RawRow**, enquanto que os data file são baseados em bytes. Outro ponto muito importante, as tabelas podem ter índices, então, quando uma busca é feita por exemplo, os dados retornados podem vir tanto do arquivo original, quanto de um arquivo de índice; a depender de qual campo é passado.

As tabelas abstraem a utilização dos índices, tornando as buscas mais rápidas e inteligentes. Bom, mas bora para as implementações.

## Get

```go
func (t *Table) Get(column string, value any) ([]RawRow, error) {
	
	row := make([]RawRow, 0)
	// Get the pk column
	pk := t.PrimaryKey
	serializedValue, err := utils.Serialize(value)

	if err != nil {
		return nil, err
	}

	// If the column is the primary key, we can use the datafile to get the data
	if pk.Name == column {
		// Serialize the value

		row = append(row, t.FromKeyValueToRawRow(t.PDataFile.Get(serializedValue))...)
		return row, nil
	}

	// We try to find indexed DataFiles
	index, ok := t.Indexes[column]
	if ok {
		// Get the index DataFile
		indexDataFile := index.PDataFile
		// Get the value from the index
		indexValue := indexDataFile.Get(serializedValue)
		row = append(row, t.FromKeyValueToRawRow(indexValue)...)
	} else {
		// Worst case, there must be a scan through the datafile
		// Create a crawler at the beginning of the file
		crawler := btree.GoToFirstLeaf(t.PDataFile.GetBTree())
		// Loop through the datafile
		keyValues := make([]btree.BTreeKeyValue, 0)

		for {
			kv := crawler.GetKeyValue()
			keyValues = append(keyValues, *kv)

			// If the crawler is at the end of the file, break the loop
			if err = crawler.Next(); err != nil {
				break
			}
		}
		row = append(row, t.FromKeyValueToRawRow(keyValues)...)
	}

	return row, nil
}
```

What happens inside this function? The inputs *column* and *value* are analogous to a query condition, like **"WHERE x = 1"**. So the input would be **column = 'x'** and **value = 1**.

They are received and the input *value* is then serialized to be used in a possible search. If it is a table with a primary key, the system gets the pointer to the **DataFile** structure and performs the search for the proposed key.

If the column passed is not the main column, the system checks if there is an index for that column and, if so, uses it to perform the search. Otherwise, the table is scanned.

But wait a minute, where did this scan come from and what is it?

This is the surprise and introduction to the next chapter. The scan is an operation performed by the **BTreeCrawler** structure and theme of the next chapter, which basically scans the table, taking each added row and checking whether or not there is a column with the respective value.

This get function is useful for search cases where indexed columns are matched to values, for example:

```sql
WHERE indexed_column = 13 OR indexed_column IN (2,3,4,5,6,7)
```
## Insert

```go
func (t *Table) Insert(rows []RawRow) (int, error) {
	// First step - Validate rows

	validatedRows, err := t.ValidateRawRows(rows)

	// Returns case there is an error on rows validation process
	if err != nil {
		return 0, err
	}

	// We have to insert it into the base table and also into the indexes
	for i, row := range validatedRows {
		// Insert the row into the base table
		serializedRow := t.FromRawRowToKeyValue(row)
		t.PDataFile.Insert(serializedRow.Key, serializedRow.Value)
		// If there is indexed tables, we have to insert the row into the indexed tables
		for _, index := range t.Indexes {
			// Get the index DataFile
			indexDataFile := index.PDataFile
			// get the column name to be serialized
			indexKey := row[index.Column]
			// Serialize the value
			serializedIndexKey, err := utils.Serialize(indexKey)

			if err != nil {
				return i, err
			}
			// Insert the row into the index table
			indexDataFile.Insert(serializedIndexKey, serializedRow.Value)
		}
	}

	return len(rows), nil
}

```

The insert function doesn't have many secrets. All rows set to be inserted go through the validation function. If there are no problems, the rows are inserted, which occurs both in the original datafile and in all existing indexes. The total number of rows added is returned.

## Delete

```go
func (t *Table) Delete(rows []RawRow) (int, error) {
	// Iterate over the input
	for _, row := range rows {
		// Delete the row from the base table
		serializedRow := t.FromRawRowToKeyValue(row)

		t.PDataFile.Delete(serializedRow.Key)
		// If there is indexed tables, we have to delete the row from the indexed tables
		for _, index := range t.Indexes {
			// Get the index DataFile
			indexDataFile := index.PDataFile
			// Delete the row from the index table
			indexColumn := row[index.Column]
			serializedIndexKey, err := utils.Serialize(indexColumn)

			if err != nil {
				return 0, err
			}
			// Get the keys from the index
			existingKeys := indexDataFile.Get(serializedIndexKey)

			// Indicates that there is more than one key for the same value
			// All keys must be deleted and the keys that are different from
			// the one that is being deleted must be inserted again
			if len(existingKeys) > 1 {
				deserializedExistingKeys := t.FromKeyValueToRawRow(existingKeys)
				reinsertKeys(t, indexDataFile, deserializedExistingKeys, row, serializedIndexKey)
			} else {
				indexDataFile.Delete(serializedIndexKey)
			}
		}

	}
	return len(rows), nil
}

func reinsertKeys(
	t *Table,
	indexDataFile *files.DataFile,
	existingKeys []RawRow,
	row RawRow,
	serializedIndexKey []byte,
) {
	musReinsert := make([]RawRow, 0)
	for _, iRow := range existingKeys {
		indexDataFile.Delete(serializedIndexKey)
		if iRow[t.PrimaryKey.Name] != row[t.PrimaryKey.Name] {
			musReinsert = append(musReinsert, iRow)
		}
	}

	// Reinsert
	for _, iRow := range musReinsert {
		serializedIndexRow := t.FromRawRowToKeyValue(iRow)
		indexDataFile.Insert(serializedIndexKey, serializedIndexRow.Value)
	}
}
```

There are also not many differences for the delete function compared to the insert function. The same process occurs, where the rows are received as input to the function, the primary key field is extracted and serialized to the table, and then the row is deleted. Since the same row must be deleted from the other index files, we must do it for all of them.

There is only one point of attention in this case. Since there is a remote possibility of having duplicate keys in index files, when the deletion process occurs, all rows containing the key end up being deleted, but this should not happen. Below is an illustrated example of this problem:

The table below contains three columns, **pk** (unique primary key), **sk** (indexed secondary key) and **other_column** which is any column.

![alt text](../../assets/pk_sk_example.png)

We want to delete the row where the value **pk** = 2. What happens in the system?

1. The table has two *DataFile* files, the main one indexed by the key **pk** and the secondary one indexed by the key **sk**.
2. When deleting the row with the value **pk** = 2, the system easily deletes it in the main *DataFile* file. However, when proceeding to delete the file indexed by the key **sk**, a problem occurs.
3. Within the function, when the system searches within the indexed file for the value of the **sk** column of the row where **pk** = 2, that is, **sk** = 2 as well, two results are returned; one referring to the row that should be deleted, and the other, referring to the row where **pk** = 3.
4. Since the datafile delete function deletes all items that have this key, an unwanted deletion will occur of the row where **pk** = 3. Therefore, this key must be reinserted together with its value.

## Update

The update function is a deletion followed by an insertion. There was no separate implementation.

# Finalization and the Range function

I don't know if you noticed, but there was a specific method that wasn't mentioned much, the Range. While I was building the table methods and creating some tests, I realized that when I wanted to execute a specific query or the entire file, I would have to somehow read all the information from the sheets.

The big problem, including when assembling the Binary Tree functions, was having this information tracked, that is, the location of all the sheets of the binary tree. After doing some research and thinking about how I could solve this problem, I came up with the idea of ​​implementing some kind of binary tree page browser. In homage to internet crawlers, I also called this search engine BTreeCrawler.

Its role is to navigate binary trees and find specific sheets and enable navigation forward or backward, through the structure's methods. In addition, it is also possible to read (dereference) the values ​​of your current position.

I hadn't realized it, but without this structure, it would be very slow and difficult to perform queries. Therefore, the next chapter will deal with its implementation and how it works. Until then