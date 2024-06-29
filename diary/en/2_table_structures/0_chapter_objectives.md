# Why this chapter

It is always good to know where we want to reach. Having objectives and knowing why are you doing something, leads you to a more successful path. Normally, people tend to accept impositions better when the explanation is given, for example: You are at the University and the Professor comes to the class and says, you are going to learn about capacitors today!

Imedietely you ask yourself why? which is reasonable, but if he had given an explanation, such as: "You are going to learn about capacitor today because they are widely used in electronics and are part of almost every eletronic project serving many purpouses in this field". Would have been much better, no? 
That applys to everything in our lives. If the explanation is given, even though it is not in the best form, people tend to accept it better, and that has the same effect when undertaking actions in your job or projects.

In the first chapter I didn't say why I started my project with binary trees. Perhaps that has something to do with the fact that I was researching about database files and findout that they were data structures in files. But I forgot to say why.

Since I have already started the development of it and got somewhere, I'll try my best to give you an exaplanation of what are the next steps and why I have chosen them, in addition to my chapter goals.

## Goals for this chapter

I created the basic CRUD for binary trees, I implemented tests and it seems to be working good in its most basic form (The binary tree must still be refactored). But what now? 

I want to save structured data into it. I also want to insert and manipulate table rows as if it were real table data, with columns, types and so on. In other words, the table structure layer will be built in this chapter.

To achieve this goal, I must be able of dealing with structed data and that requires a lot of organization in table structures and some definitions, along with understanding serialization in golang and implementing other tests to prove some concepts.

For the next sub-chapters, I'll deal with: 

- Serialization
- Integration of Binary Tree into a File Structure, which is going to be used in the future for files access management purpous, 
- How will I serialize table columns 
- Binary Tree Crawler (A type of binary tree iterator, that walks through the btree leaves - Used for range queries)
- Database and Tables structures
- The minimal table CRUD

At the end of the sub-chapters, the system will already have all basic table functions, for manipulating table data in the binary tree files trought the table interface.

A short example:

```go
// Table interface example
table := LoadTable("person")
// Table data rows
tableData := []map[string]interface{} {
    {
        "name": "Nicolas",
        "age": 32,
        "email": "nicolas@teste.com",
        "country": "brazil"
    },
    {
        "name": "Enrico",
        "age": 0,
        "email": "enrico@teste.com",
        "country": "brazil"
    }
}

// Insert both for indexed and not indexed table
table.insert(tableData)

// Or for updating table as well
table.update(tableData)

// Also for deleting
table.delete(tableData)

// There is also a rang query
table.rangeGet(rangeParameters)

```

Although the functions are not implemented yet, the goal is to have all of them working properly for a table with different indexes and composite keys as well.



