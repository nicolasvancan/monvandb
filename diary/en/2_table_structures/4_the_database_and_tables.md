# An outburst

Writing diaries has been a bit chaotic in my daily life. There are times when I write it as an experiment report. Other days, I write down topics and things I thought of on sheets of paper to rewrite when I have time. But what has taken me the most time is not even writing, but making writing more understandable for the reader.

It's interesting to note that when I'm thinking about something, whether it's a solution or some development idea, it's very clear in my head. However, when I write down all the logic I'm coming up with, it's not as clear as what I had imagined, and that makes me think about some hypotheses for that:

1. My lack of writing experience influences how I express myself in texts (I express myself poorly)
2. The lack of time to write makes me try to be very quick and direct with the points, lacking details

I believe there is a little of both cases above. I really don't have that much writing experience, even though I feel like I've improved after a few chapters of this project. Lack of time is normal, work, parenthood, family; All these things demand my time, but I do them with great pleasure.

One last point was missing above, which is perfectionism. It's not something I struggle with, although I'm very critical of myself. Whenever I review things I've done, I criticize myself and try to improve them. I think that's why I keep rewriting texts and ideas. Maybe if I remove the pressure on myself a little it will improve.

Anyway, after starting to release chapters on Linkedin and revisiting not only the codes, but mainly the texts, I'm thinking about rewriting them without losing the essence of the diary, but making them better.

## Baby Milestones

When I started implementing the binary tree, my son was unborn. Time passes quickly, as I write this text here it is almost 6 months old. And that brings news, like for example him trying banana this morning. He didn't like it very much. But it must be strange after 6 months of breastfeeding try food.

# Table Structures

Returning to the main subject, an important moment in the project has arrived and which accompanies many definitions, which are the **Tables**. In my view, this is the core of the database and therefore should be treated with great care.

I didn't snoop around other projects to see how they did their table definitions, I based it on my experiences. So let's go! What I imagine as table structure and its main functionalities.

## The context

Whenever I think about any type of database functionality, I immediately imagine a query being sent by the user to the server. From there things start to unfold. Processes are created, threads executed, files accessed, etc. In the same way, I will use a simple query but as a good example of reflection on the point I want to reach today.

**The query**

```sql
SELECT 
    u.user_id,
    u.name,
    u.age,
    a.address,
    a.address_number,
    a.city,
    CAST(u.ssn AS STRING) as ssn
FROM tab_users as u
    INNER JOIN tab_addresses a ON u.user_id = a.user_id
WHERE a.city in ('London', 'São Paulo')
```

This simple query searches the **tab_users** table, which is used to join the **tab_addresses** table using the **user_id** column. Furthermore, there is a conditional clause that limits the search, whose **city** column of the address table is equal to *London* or *São Paulo*.

Ideally, the system reads this query, identifies the reading request and makes execution calls to the system. But what are these calls?

Well, I don't really know either. In fact, I don't know specifically what they are, as they can be minimal, complex; depending on how I implement them. And that's something I haven't even planned yet. However, what I know should exist are the readings in the table files. Remember those binary tree files I put together in previous chapters? Then, they must be read every now and then by the system to bring the data as results of the queries.

This leads me to think about different solutions and possible problems. Of course, the first version of the database will be simple, with simple functions and simple access. The most complex problems are addressed as they arise. But this is also a subject to discuss later.

Between joins, commands, conversions, comparisons, or any action that needs to be done with the data, the most basic and necessary action among them is the acquisition of the data itself, right?

That's where our active **Table** comes in. In addition to containing table definitions, such as column types, indices; it is also responsible for containing the access methods in a simple interface. The table must be simple enough to be called by several server instances or connections and deliver the data that clients need.

Given the initial context, let's move on to the brief definitions made.

## Definitions and structures

For those who are familiar with other RDBMS, the tables are contained in databases (*databases*), which are isolated from each other. A database can have tables, views, procedures. At first, my database will only contain information and access to tables. Below is an initial example of what my **Database** structure looked like:

```go
type Database struct {
	Name       string            // Database's name
	Tables     map[string]*Table // reference to Tables
	TablePaths map[string]string // Paths to the tables
	Path       string            // Path to the database dir
}
```

Well, first of all, what are these fields? As I haven't even defined the tables yet, there is no way to store the system information in tables yet. It's that same compiler paradox, the first compiler for a new language is built in another language. After being able to program in the new language, you can build a compiler in that same language.

So to be able to save my files in some location, I decided to create the structure containing the name and path of the **database**, which tables this database has, the paths to access these structures. Why save the paths of these structures? When starting, the system will load the metadata of all existing databases and tables saved in it. For this and for now, you will use these files.

But this brings us to another problem: **How ​​to save these files?** 

We can save anyway. I chose a readable and simple way to serialize, which is **json**. I used it for this purpose only for debugging and later reading purposes. Ideally, the metadata of anything within the database should also be stored in database tables, especially because, in the RDBMS concept, everything becomes a table. 

Below is the serialization example for **json**

```go
// Function that converts struct to json
func ToJson(value interface{}) ([]byte, error) {

	// Create a new gob decoder and use it to decode the person struct
	enc, err := json.Marshal(value)
	if err != nil {
		fmt.Println("Error encoding struct:", err)
		return nil, err
	}

	return enc, nil
}

// Function that converts json to struct
func FromJson(value []byte, dst interface{}) error {
	// Create a new buffer from the serialized data
	err := json.Unmarshal(value, dst)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

```

An additional module was created to contain utility functions for all modules, called **utils**. In addition to the serialization functions, I created another file in the same module containing functions to deal with generic files, located in **src/utils/files_utilities.go**. Later I will comment specifically on this file and some details about its implementation.

### Tables and Columns

What should the table structure have? Columns, of course. Maybe a name? What else does a column have? Well, from my knowledge and guesswork, columns are necessary, a primary key column is also necessary.

**Like this?**

There is no binary tree file without index, right? When we define some table, we need to define its primary key. Normally it is a column with the name **id**. Here's an example:

```SQL
CREATE TABLE IF NOT EXISTS tab_users (
    user_id AUTO INCREMENT PRIMARY KEY BIGINT,
    name STRING,
    age INT,
)
```

Our table has a column that is identified as the primary key. In other words, when created, the table will index the rows by the **user_id** column of the table, and searches will be carried out based on it.

There are also cases where composite keys are created, see another example below:

```SQL
CREATE TABLE IF NOT EXISTS tab_users (
    user_id BIGINT,
    name STRING,
    age INT,
    PRIMARY KEY (user_id, name)
)
```

Above, the primary key is defined as the combination of the user_id and name keys. Therefore, it is also necessary to know whether the table is a composite key table or a simple key table. Finally, a table can have indexes outside the primary key. Another simple example is a table that contains data in a column such as **insert_timestamp** and we want a search index to be created for it. This way, when we provide a filter based on the **insert_timestamp** column, the search becomes indexed, and there is no need to scan the entire table. But we got to that at some point.

Well, I think for the first version of our micro database this is already good. Putting the information together, as a result, we had the table structure shown below:

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

type Table struct {
	Name         string            // Table's name
	Path         string            // Where the table configuration is stored
	Columns      []Column          // reference to Columns
	PrimaryKey   *Column           // reference to PrimaryKey
	CompositeKey []Column          // Case column is composite
	Indexes      map[string]*Index // reference to Indexes
	PDataFile    *files.DataFile   // private Access btree (Simple)
}

type Index struct {
	Name      string
	Column    string
	Path      string
	PDataFile *files.DataFile
}

type Column struct {
	Name          string
	Type          int
	Default       interface{}
	Nullable      bool
	AutoIncrement bool
	Primary       bool
}
```

Starting with the **Column** structure, whose fields are respectively: the name; the data type being one of the enumerate options; Default if nothing is passed and there is a default value for the column; Nullable if the field is allowed to be null; AutoIncrement used to automatically increment indexes; Primary indicating whether or not it is a primary column.

For the **Table** structure, in addition to the **Columns** field, there is the PrimaryKey which is an address of a **Column** structure and the Composite key field which is an array of values ​​from the ** structure Column**. When the table has only one primary key, the CompositeKey array has the value *nil*. If it is a composite column, the primary key pointer PrimaryKey is *nil*.

The Indexes field uses a map of key equal to string and value equal to pointers to Index structures. The PDataFile field points to the DataFile structure of the composite key or Primary key, depending on the table configuration.

Finally, index structures contain information such as name, column and path, as well as the pointer to their respective DataFile.