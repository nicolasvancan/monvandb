# First Step Understanding what is a rdms

When I decided that I wanted to build my own database from scratch, a lot of doubts emerged, including this principal one: What is in fact a Database?

So I ran looking for projects, files, documents, that could describe how databases are normally built. Unfortunately, some of them are kind of a Black Box, and others open projects were to complex to have a basic understanding rapidly. Nontheless I could find some usefull articles and resumed ideas of what a database idea is and some start points for what I wanted to build.

I took a look into SQLite projects, which gave me a good idea about what is the purpous of a database and some overviews about architectural software implementations. For those who may be interested in reading it, the link to the documentations follows in [link](https://www.sqlite.org/docs.html). The source code of the project can be found [here](https://github.com/sqlite/sqlite/tree/master).

Behind almost every single rdms database there is a efficient Data Structure, used to perform fast searches through indexed data. B-Tree and B+Tree are often used as databases those data structures. Clearly it's just a small part of a Big Project, but we must start small to get big. Indexes, tables, Views, Commands, Cache, and many other features will be further covered as soon I advance and create 

In this first section I'm going to understand deeply how does a B-Tree work and how it is dumped, read, updated, to a database file using OS FileSystem. The main objective of the initials commits and the first generated tag is to undestand some main concepts of the B-Tree data structure, how to work with files containing that data, and build up the core methodes to work with this data.

**Refs**:
- **B-Tree**: [https://builtin.com/data-science/b-tree-index](https://builtin.com/data-science/b-tree-index)
- **Bee Tree Project**: [https://github.com/amit-davidson/btree/](https://github.com/amit-davidson/btree/)

**Note:** It's my first time programming with Golang. Therefore, it's possible that my code might not be the most efficient and well built code for this language. I've found some projects that already implement what I want to do, since I want to learn deeply and improve my knowledge, I'll do it by myself.

# B-Tree

One of many existing data structures, B-Tree is really usefull when dealing with large sets of data and in memory in disk. Seeing that Databases persist data in disk and must be fault tolerant, it's handy to use B-Tree. (edit)

- How B-Tree would work instead of dealing with memory pointer, doing it in disk mapping
- How do pages in Disk work
- Using Node as pages in a disk
- How to update tables in a disk, since it's hard to change fisically every node after every change. Use append only files and change pointers?
- How to save and dump data into table
- File Sync instead of writing?
- Using a proxy for dealing with memory pointing and converting that directly to disk pointing?

Adicionar cada teste feito, cada coisa descoberta e como resolvi cada um destes problemas


day 1 - what I have learnt exploring mmap, files, using fixed length byte array for covering file pages, debate over structure, etc
day 2 - How I've decided firstly to implement my interfaces and so on, how I tested dumping and retrieving data from file directly as BTree
day 3 and 4 - Implementing basic structures for bTreeNodes and bTree and testing (Why I've decided to do so)

