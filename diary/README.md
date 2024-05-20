# Chapters

The Diary is composed of many different chapters, each containing one or more diary files (Markdown), resulting in a real story of how the database was built. The division of chapters was made by milestones in the project; in other words, when key parts are completed (its basic version), one good example is the Binary Trees module, which comprises the first chapter of the series. That means that every new feature of the software will gain a storytelling list of markdown files.

I've been thinking about how to divide the chapters to make a good story, but I got nowhere. I listed some possible modules that may enter the software but not all of them. So, it's possible that, as the project evolves, the order and this file itself change. But anyway, the first chapters are the following:

- Binary Trees and Components
- Tables and Databases Definitions
- File Access and Files Management
- Dataframes
- Cache and Proxy access to tables
- Queries and Commands
- Server
- 
If one day I finish what is in this list, I'll probably be much older (Depending on the effort and time I spend on the project), and I'll also be much wiser (in terms of software development).

# The Project

A database can be practically anything. Yes, a simple text file can store relevant data that is important to a company, for instance. But storing data is not the only thing that a database can do. In the real world, data is stored, retrieved, replicated, and consulted all the time by different sources and businesses.

Data has become the new gold. The better the data quality, the more you can earn money with it or even create better products for your business and clients. That is why storing it is so important, but far more important is how we deliver it, ensuring safety, velocity, and security.

That said, a simple text file would not fulfill all requirements to be a good database (in an enterprise level), but what would in fact be a good database file?

To be honest, I don't know the answer to this question. That lack of knowledge led me to study more about this topic, including its definition and much more that will be covered as I learn it.

## What is a Database?

According to Oracle's definition, a database is an organized collection of structured information or data, typically stored electronically in a computer system. A database is usually controlled by a database management system (DBMS). Together, the data and the DBMS, along with the applications that are associated with them, are referred to as a database system, often shortened to just a database.

Okay, fair enough. That describes well what I want to build: a Database Management System. But the previous description is the same as describing a car as a driving cabin built upon at least three or more wheels that can take the driver and more passengers to another place. In other words, it's too generic. I need to know a little bit more about its functions and inner parts.

Using my research abilities, I started the quest of understanding the inner components of a generic database, starting from database files and their structures. After some reading, I found some interesting articles saying that database files and how they are structured depend on the database. Some are open source, so you can read their code and see what is happening there and how they save their data, and many others are completely hidden.

I don't want to copy a project and build something based on another one. I want to suffer and make it my own way, out of nowhere. Of course, I am not reinventing the wheel; I am just building some parts of the wheel again in my own way (for the purpose of learning and also suffering).

Nonetheless, I thought it would be nice to read some information about existing database projects to get some insights. I researched more and found out that the library SQLite was not only open source, but also provided some architectural definitions and how they acted in solving some common problems related to databases.

If anyone is interested in seing it, just click on this [link](https://www.sqlite.org/arch.html).

### What else might a Database need?


Beyond database file structures, there is much more to be learned in order to gather all pieces of the project. Whenever we modify a file, using Operating Systems interface, we must ensure that all the computations will succeed, that the concurrency is not going to be a problem, and also when happening unprecedented faults, the file will never be corrupted. All of this must be treated by a File Access Manager or some sort of software layer that deals with this sort of stuff.

On the application layer, on the other hand, there must be an interface for queries that can receive one or more commands, execute them, and return the results correctly. Are those queries going to use SQL interface? What are going to be the capabilities of the database? All of that will be covered by the part of Query System (I must invent a better name for that).

Manipulating data, joining tables, editing columns, is a common work for developers, and the database must use a good mechanism to deal with this sort of job. Therefore, a cache and a joining system must be implemented.

Enabling not only data management, but also tables and Databases management, working with views, indexes and so on.

Users and permissions must be implemented; imagine that you hire a new brave but reckless employee that runs commands without double-checking them, and he/she does it in a production environment, that is risky, no?

A server must be built so that the requests can be received and processed by the system.

Of course, there is much more to be implemented for a Database. It doesn't mean that I must implement all at once though. Therefore, I've made a list of the principal components that will be built for the first version of the functional database.


### Database and Its Core components


#### Binary Tree

Storing data permanently requires at least a file (or a composition of them). Those files have their own structure, meaning that tables, views, information, rely on the data structure and file structure organization. The ability to read data fast depends on how data is stored, whether they are stored in a sorted list or stored in hash tables, binary trees, every data structure has its pros and cons. Binary Tree is typically chosen as a data structure for a database. It is fast to store/delete and also fast to get data, having its complexity of O(log(n)) for both cases.

#### Tables and Databases Definitions

Creating structures that define what a table and databases are and how they are defined. Methods that

#### File Access and Management

Accessing files in a Database is a crucial move. When we are dealing with only one thread to access and modify data, there is no concurrency problem, since everything is going to happen serialized. The other case, which is where I want to reach is enabling the Database to do concurrently many different operations in the same file, without corrupting it, without blocking it, and also providing a way of rolling back some modification whenever I need.

Those situations are practically what guarantees a database's success, and it must work well and fast. Therefore, a good file access management system will be necessary for the Database System.

#### Cache Database

Transactional databases must retrieve data rapidly to the client. To achieve that, cache is a key actor in the process, storing recently researched data from different tables, cache plays a big role here.

#### Commands

Database must get data based on what the client wants. Have you ever heard that the client is always right? In a database, that doesn't apply; sometimes the client writes wrong queries or nonexistent commands. Jokes aside, the command module will contain not only the commands interface executed by the database and tables, but also the queries and commands decoder, optimizer, and translator. If I am not mistaken, it will be one of the largest modules in this project (All the decoders, such as natural language processing and understanding will be implemented out of nothing)

#### Server

A server must be built to receive requests from client connections. Will it conform to the (ODBC) model? This is a hard question, must think over.

#### Multi-Thread System

I think this might not be a module, but a concept used in each component of this project. Everything must be prepared to work concurrently, using multi-threads and processing.