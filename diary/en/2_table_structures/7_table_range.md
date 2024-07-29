# The range method

This function took a lot of my time and creativity, mainly because I still haven't decided on several other layers of the software. Before going into any technical details, I should contextualize the premises and possibilities of this function, and what took me so long to think about.

Let's go! By definition, the **Range** method is a range search. In an indexed binary tree, the range is dictated by the range of indexes that you want to search. Let me give you an example: A table that has integer indexes has key values ​​that vary from 1 to a very large number. Therefore, searches can be performed between any combination of conditions of these values, for example: greater than one; less than five greater than zero; not equal to two. All these verification conditions generate different search ranges in the table.

**Why not use scan for all searches?**

I'll answer with another question: Why use binary trees then? The purpose of this structure is to make searches for the keys saved in it more agile. Always using scan to perform queries would make our system inefficient.

**So what is the best scenario for this?**

Given a SQL query, there must be a system capable of analyzing all the conditions of the table, its existing indexes, subjoins, etc.; and creating a search for the most efficient range for the possible scenario.

# The idea behind this implementation

Now I will create several fictitious queries to demonstrate the idea and the reason behind this implementation. You will notice that some things, mainly in the data structure, will probably change over time, considering that I made several assumptions and thought that the system would work in a certain way.

Let's go, our test scenario contains the following tables and their definitions in these structures

```sql
CREATE TABLE IF NOT EXISTS person (
    person_id AUTO INCREMENT KEY BIGINT,
    address_id FOREIGN KEY BIGINT,
    name STRING,
    age SMALLINT,
    email STRING
)

CREATE TABLE IF NOT EXISTS addresses (
    address_id AUTO INCREMENT KEY BIGINT,
    street_name STRING,
    number INT,
    postal_code INT,
    country STRING,
    indexed_column BIGINT
)
```

These are two fictitious tables with the sole purpose of being able to make queries. And these queries will also be simple, but they will give us the basic idea of ​​what I want to show. Another detail worth mentioning is that the person table is also indexed by the address_id key.

Suppose you want to make a query with the person table, for example, to search for all people who are under 30 years old. We would have the following query:

```sql
SELECT * FROM person WHERE age < 30
```

A single simple condition in which the age column is less than 30. What is the scope that our table needs to perform to check all this data? Let's go to the explanation: The person table has a binary tree with key **person_id** and an index created later for the **address_id** column; in the case of the age column, we have no way of knowing, in an indexed way, where the records are located. Therefore, it is necessary to scan the entire table and check, record by record, who are the people who are under 30 years old. But what if we made a change in the search and put it in the following format:

```sql
SELECT * FROM person WHERE age < 30 AND address_id BETWEEN 18 AND 78
```

Now the situation has changed because we have another restrictive condition *AND* that says that the address_id, which is indexed in this table, must be between 18 and 78. Therefore, now our search range must be greater than 18 and less than 78.

**What happens with joins?**

Joins are junctions of tables, or parts of these tables with others. When joins are performed, the entire table must normally be searched, in both cases, where the records are the same. Obviously, there are techniques that reduce the processing time for these joins, but the range must be total, unless there is some restriction condition of a table that already reduces the size of the range. Below is another example of this:

```sql
SELECT *
FROM person p
INNER JOIN addresses a ON p.address_id = a.address_id
WHERE a.indexed_column > 700
```

Note that here a join is performed and, as a rule, both tables would need maximum range searches (scan). However, it is known that for table a there is a condition in which the indexed column **indexed_column** must be greater than 700. In this case, the range of the addresses table is reduced to that condition.

**What about multiple conditions?**

The case of multiple conditions always follows the restriction rule based on the comparative columns. Among all possibilities, the worst case is always the scan of the table. Let's see, if I can't infer a range limit for the search, it directly means that I need to scan the table.

For *OR* conditionals, the worst case is always taken into account, while for *AND* conditionals, the best case is always taken into account, for example:

```sql
-- First query
SELECT *
FROM person p
INNER JOIN addresses a ON p.address_id = a.address_id
WHERE 1 = 1 OR a.indexed_column > 700

-- Second query
SELECT *
FROM person p
INNER JOIN addresses a ON p.address_id = a.address_id
WHERE 1 = 1 AND a.indexed_column > 700
```

The queries are similar, with a single difference in the conditional for the **indexed_column** column. In the first case, the range must be total (Scan), while for the second case, the range must be > 700 for the keys.

Working with multiple conditions requires the system to evaluate all columns placed as conditional joins, so that it identifies the indexed columns as well as their type of operation, whether it is an *AND* or *OR*, and group them for each specific table to decide which is the best and most optimized range for a given query.

**Why do I want to perform only one search in the tables?**

The worst case of data access is precisely in files, where reading is much slower than in memory. In this case, if we reduce the number of times a file is accessed, it is better for our system and the results will come faster. Therefore, I determine the best case of accessing the files and do it only once, and all other checks, transformations, and joins come from this data returned from the Range

# How it works

Now comes the implementation part and really shows how I thought of the solution strategy for this problem, finding the best range of the Range function.

There are two possibilities for the range, returning a pointer to a **BTreeCrawler** structure or returning an array of **RawRow** **[]RawRow**. For now, I have chosen to implement only the return of an array of raw rows, although the work for the first option would be the same. Therefore, our range function has the following scope:

```go
func (t *Table) Range(input []ColumnComparsion, limit int, order int) []RawRow {
// Implement Here
}
```

## ColumnComparsion

Note that there is an input variable for the function called *input*, whose type is an array of **ColumnComparsion**, not yet mentioned here in the project. What would this new structure be?

Now comes the great part of my suffering for this task: The definition of this structure!

After much reflection, I came to the conclusion that what can delimit or restrict the scope of searches in *DataFile* files (Binary Tree) is the type of column comparison. It is exactly what I had introduced as a concept at the beginning of this chapter, comparisons like: "***WHERE x > 1 AND (y = 1 OR z = 3)*** ", ***INNER JOIN sometable t ON t.x = x***.

The big problem is analyzing huge queries, because they can be extremely simple as well as absurdly complex, cascaded, nested, recursive; and the big point here is to extract the information from columns and try to determine the best search scenario for a data file.

I then decided to explain the possibilities of comparing a column in a data structure called **ColumnComparsion**, whose schema is defined below:

```go
type ColumnComparsion struct {
ColumnName string // Column name
TableName string // Table Name ()
Condition int // Condition EQ, NEQ, GT, GTE, LT, LTE, IN, NIN, LIKE, NLIKE
Alias ​​string // Alias ​​for the column
Value ColumnConditionValue // Value to be used in the comparsion
Id int // Identifier for comparsion, example WHERE x < 1 AND x > -3, the identifier belongs
// to the layer of comparsion. Bosh x<1 and x>-3 have the same identifier, meaning that they are in the same layer 
ParentId int // Parent layer id, if it is base layer, it is -1 
ParentLogicalOp int // Parent AND or OR 
LayerLogicalOp int // AND or OR 
} 

type ColumnConditionValue struct {
	IsOtherColumn        bool                                         // Indicates if the value is a column comparsion
	IsOtherTable         bool                                         // Indicates if the value is a column from another table
	ColumnName           string                                       // Column name
	TableHash            string                                       // Table Hash (Used to get the table from the database)
	Value                interface{}                                  // Value is only used if isOther columns and isOtherTable are both false
	Transformation       func(interface{}, []interface{}) interface{} // Still not implemented
	TransformationParams []interface{}                                // Still not implemented
}

```

This structure was designed to also be used for the commands section, which was not even planned, so I try to take several points into consideration. Among them are the data comparison and transformation section. Although I am not sure if this structure will remain the same until the commands section is implemented, I decided to continue with this implementation, since what would change in the future would only be a few details regarding the structure's internal fields. However, the implementation logic would remain the same, and no absurd refactoring would be necessary.

**Fields**

The column that receives some type of comparison has its name noted in the *ColumnName* field and its respective table in *TableName*. Next, the conditions field, which follows an enumeration of conditions such as: GTE (Greater Equal); LT (Less Than); EQ (Equal). The alias field is specified if the column has an alias linked to it. Finally, there is the Value field, which is specified by the **ColumnConditionValue** structure, and is nothing more or less than the specification of the comparison. Whether or not it is another column from another table, or whether it is a value; or also whether it is a value that needs to be transformed before being compared.

These are the main fields related to values ​​for comparisons. However, it should be noted that I have not yet commented on the *Id*, *LayerLogicalOp*, *ParentId* and *ParentLogicalOp* fields, as they are part of a comparison level abstraction layer.

## Comparison Levels

Analyze the following query and try to analyze which levels of comparison layers exist.

```sql
SELECT
    c.colA,
    b.colB,
    b.colC,
    a.colD,
    a.colE
FROM tableA a
    INNER JOIN tableB b on a.colB = b.colB
    INNER JOIN tableC c on a.colC = c.colC AND c.colD < 30
WHERE c.colA IS NOT NULL AND (b.colA BETWEEN 13 AND 89 OR b.colC != 'some_value')
```
There are three tables whose data files must be accessed. The system must therefore try to determine the best case for accessing the files by comparing the columns of each table. To do this, we must analyze what types of comparisons are being made and in which layers they are located.

**Layers**

The root layer, or layer zero, is the direct layer from the *Where* clause and also from the *Joins* and their conditions. Any other compound conditions that are enclosed in parentheses are considered conditions of other layers, and the more nested layers there are, the higher the logical level of the comparisons. Below is a picture of the same query above with the layers specified:

![alt text](../../assets/range_query_1.png)

The green colors indicate that the comparisons are in layer zero, while the red part indicates layer 1. If there were more layers within layer 1, they would be respectively 2, and so on.

Layers identified, what do we need to do?

Now it's time to segregate the comparisons for the respective tables, let's go. There are three tables, so there should be three segregations, for tables a, b and c respectively. Each of them corresponds to an array of the **ColumnComparsion** structure.

To make things easier, I won't define the structures here yet, but I will only put the comparisons in text format.

- **tableA**: ['a.colB = b.colB on layer 0','a.colC = c.colC on layer 0']
- **tableB**: ['b.colB = a.colB on layer 0', 'OR b.colC != 'some_value' on layer 1', 'b.colA between 13 AND 89 on layer 1']
- **tableC**: ['c.colC = a.colC on layer 0', 'c.colA != NULL on layer 0', 'c.colD < 30 on layer 0']

Now we have the conditions of each table in text form, passing them to the data structure format would be the same thing, but in a format that the software would understand. I will leave just one example of one of the comparisons so that you can see how it works:

```go
// Comparsion = a.colB = b.colB on layer 0

comparsionExample := ColumnComparsion{
    ColumnName: "colB",
    TableName: "tableA",
    Condition: EQ,
    Alias: nil,
    Value: ColumnConditionValue{
        IsOtherColumn: true,
        IsOtherTable: true,
        ColumnName: "colB",
        TableHash: "tableB",
        Value: nil,
        Transformation: nil,
        TransformationParams:nil
    },
    Id: 0, // layer
    ParentId: -1, // No parent
    ParentLogicalOp: AND, // Default
    LayerLogicalOp: AND // Default
}
```
## RangeOptions

This structure is the result of the final evaluation of our system. With **RangeOptions** we can quickly define from which key the search will be performed, as well as up to which key the search will follow. The structure is shown below:

```go
type RangeOptions struct {
	From        []byte
	To          []byte
	FComparator int             // From Comparator, indicates what type of comparation should be done with the From value ex: GTE or GT
	TComparator int             // To Comparator, indicates what type of comparation should be done with the To value
	Order       int             // Order of the range wheter is ASC os DESC
	Limit       int             // Limit of the range
	PDataFile   *files.DataFile // Pointer to the data file to be used
}
```
In addition to the *From* and *To* fields, we have two additional fields that represent the type of operation that delimits the stop and start of the search. The FComparator delimits the start, for example: when FComparator is equal to GTE (Greater Equal), it means that the RangeOptions starts at the key in From and goes until the comparison of the To key. For the stop condition for the To key, we have to evaluate what is written inside TComparator, if it is GTE (Greater Equal), it means that the stop condition is when the key reaches a value greater than *To*.

The other fields, *Order*, indicate the search order of the **RangeFromOptions** method, whether ascending or descending, and limit indicates whether or not there is a data limit for the search (-1 has no limit). Finally, the PDataFile is a pointer to the file that the search is being directed to, which can be the base file of a table or even an index.

When the From or To values ​​are equal to *nil* it means that there is no delimiter. Therefore, the scan of a table is a **RangeOptions** structure where the *From* and *To* values ​​are equal to *nil*.

**RangeFromOptions**

```go
func RangeFromOptions(t *Table, options RangeOptions) ([]RawRow, error) {
	// If from and to are not set, we return all the rows
	if options.From == nil && options.To == nil && options.Limit < 0 {
		// Scan the whole file
		scannedData, err := scan(options.PDataFile)
		return t.FromKeyValueToRawRow(scannedData), err
	}

	// Get crawler based on the options
	crawler := getCrawlerBasedOnOptions(options)
	// Returns all data gathered from crawling the data file
	return crawlDataFileBasedOnOptions(t, crawler, options)
}

```

This method performs the Range of the table based on the range options, returning an array of **RawRow**. What it basically does is analyze the passed options file and return a **BTreeCrawler** and then use it to search for all rows given the condition of **RangeOptions**.

**Note: The other methods were omitted so as not to make this section too verbose. To see them, just access the file src/database/range.go**

## Merge

Having separated all the comparisons and their levels for each table, we need to perform the respective merges for each one. But how do we do this?

We now group them by comparison level and create a function called **Merge**, which transforms two **ColumnComparsion** structures into a single one, based on their settings. As shown below:

```go
func mergeAnd(r *RangeOptions, other RangeOptions) {
	// Create a new RangeOptions
	if (bytes.Compare(other.From, r.From) < 0 && other.From != nil) || r.From == nil {
		r.From = other.From
	}

	if (bytes.Compare(other.To, r.To) > 0 && other.To != nil) || r.To == nil {
		r.To = other.To
	}
}

func mergeOr(r *RangeOptions, other RangeOptions) {
	// Create a new RangeOptions
	if r.Order == DESC {
		if bytes.Compare(other.From, r.From) < 0 || r.From == nil {
			r.From = other.From
		}

		if bytes.Compare(other.To, r.To) > 0 || r.To == nil {
			r.To = other.To
		}

	} else {
		if bytes.Compare(other.From, r.From) > 0 || r.From == nil {
			r.From = other.From
		}

		if bytes.Compare(other.To, r.To) < 0 || r.To == nil {
			r.To = other.To
		}
	}
}

func (r *RangeOptions) Merge(other RangeOptions, op int) {
	// For end operation we restrict the range based on the lower and maximum values
	if op == AND {
		// Case it is ASC
		mergeAnd(r, other)
	} else {
		mergeOr(r, other)
	}
}
```

The premise is simple: for merges performed with the AND operator, the search range will always be the smallest of the two. In other words, *nil* values ​​are overridden by any different value and the values ​​that make the range smaller and more restricted are also overridden. While for the *OR* option, the opposite is true, everything that tends to increase the range overrides the most restricted one.

**What do we do after that?**

We merge all the levels inside the parentheses, then we take the highest levels and incorporate them recursively into smaller levels, until we reach the root. The example below illustrates what would happen with the **tableB** table from the query created above. With the following fields as text conditions:

1. b.colB = a.colB on layer 0
2. OR b.colC != 'some_value' on layer 1
3. b.colA between 13 AND 89 on layer 1

From the conditions of **tableB** we have conditions for the columns *colA*, *colC* and *colB*, with *colB* being the primary key of the table and supposing that the column *colA* is another search index, how would the merges be done?

Layer one, which contains the comparison **"b.colA between 13 AND 89"** and **"OR b.colC != "some_value""** receives the first merge. However, since this is an index that is different from the primary key, the correct thing to do is to separate and group the conditions based on the indexes, taking into account that any column that is not an index other than the primary key, that is, any other non-indexed column, will be grouped with columns from the primary key itself. This will be clear in the example below:

**Primary key grouping**

1. **Layer 0**: b.colB = a.colB
2. **Layer 1**: OR b.colC != 'some_value'

The merge is done from layer 1 to layer 0, which incorporates the values ​​from layer 1. Since this is an *OR* merge, the larger values ​​will be assumed in this join. In this case, since there are two cases of scanning the table, there would be no difference, with the *From* and *To* fields of **RangeOptions** being equal to *nil*.

**ColA key grouping**

1. **Layer 1**: b.colA between 13 AND 89

For the case of colA grouping, layer 1 is the only available layer, therefore, the generated **RangeOptions** has as keys *From* = 13, *To* = 89 and FComparator and TComparator equal to GT (Greater than).

**Resolution of merges of distinct indexes**

Well, we have reached a critical moment in which we have two range options of distinct indexes. When there is this type of operation, how to deal with it?

Since we cannot compare apples with apples, we analyze the possibilities in these cases. When merges of different types of indexes are made and the merge is the *AND* operation, either of the two can be used.

Really?

Yes, but the ideal is to choose the one with the most restrictions, that is, the one in which the values ​​of *From* and *To* are both different from *nil*. If both structures present this representation, you can choose either one.

In the case where merges are made using the *OR* option, you should look for the option that is most comprehensive, right? But we also don't know which would or wouldn't fulfill this condition. So, we see the one that presents the highest incidence of *nil* values ​​in the *To* and *From* fields. If both present these fields with values ​​other than *nil*, we assume that the system must perform a **scan** on the table, since we cannot infer which is the greatest distance to travel.

All of this logic described here can be accessed directly from the **src/database/range** files.

# The Range Table Method and its Variations

Finally, we can conclude the general table methods with the **Range** function, which receives an array of **ColumnComparsion** from the SQL command extractor translator, performs the joins of these column comparators, generates the **RangeOptions** and searches for the range in the specified data file, returning an array of **RawRow**.

```go
func (t *Table) Range(input []ColumnComparsion, limit int, order int) []RawRow {
	// create temporary variable for holding values of RangeOptions
	rangeOperation := MergeOperationsBasedOnIndexedColumnsAndReturnRangeOptions(t, input)
	rangeOperation.Limit = limit
	rangeOperation.Order = order

	// Invert the order of from and to
	if order == DESC {
		reverseAscToDesc(&rangeOperation)
	}

	// Get the range
	rows, err := RangeFromOptions(t, rangeOperation)

	if err != nil {
		return nil
	}

	return rows
}

```

## Problems and considerations

It is now possible to extract data from tables based on queries, although not as literally as in SQL. However, bringing an absurd amount of data into memory can cause some problems when several operations are performed or also when the volume of data saved in the files is immense.

In order to mitigate such action, instead of returning an array of **RawRow**, we can return a special iterator for this function, with advances and data extraction just like **BTreeCrawler**. However, thinking about the possible simultaneous use of data by more than one query at the same time, in the case of two users wanting data from a table with the same search range, an *Observer* design pattern can be used, where each new iteration of the result of this iterator, all observers are alerted with the respective row, and then they can do whatever they want with the data.

However, this is a problem that I will address in the next season, the command season.