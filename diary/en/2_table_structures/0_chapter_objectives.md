# Why this chapter

It is always good to know where we want to reach. Having objectives and knowing why are you doing something, leads you to a more successful path. Normally, people tend to accept impositions better when the explanation is given, for example: You are at the University and the Professor comes to the class and says, you are going to learn about capacitors today!

Imedietely you ask yourself why? which is reasonable, but if he had given an explanation, such as: "You are going to learn about capacitor today because they are widely used in electronics and are part of almost every eletronic project serving many purpouses in this field". Would have been much better, no? 
That applies to everything in our lives. If the explanation is given, even though it is not in the best form, people tend to accept it better, and that has the same effect when undertaking actions in your job or projects.

In the first chapter I didn't say why I started my project with binary trees. Perhaps that has something to do with the fact that I was researching about database files and findout that they were data structures in files. But I forgot to say the reason to do so.

That is why from now on I'll explain the chapter goals and why I am developing some feature or some package.

## Goals for this chapter

The binary tree can be used to store generic information. It is time to define how tabular data will be stored in the files, knowing that they only accept byte array as information, both for key and value, than what do I need to do next?

First step is to understand how to work with serialization in Golang. Knowing how Golang deals with this sort of problem, I must ensure that real table rows are stored in the binary trees and, for that problem, the solution is to build the table, columns, databases definitions. Essentially, all types of structures that a Database might need.

Tables are essentially made of columns and rows. Columns are the core definition of a table, holding informations of their types, contrainst and values. Rows are collections of multiple columns aggregated in an unique data type. Besides, a table is always indexed by at least one column, which is normally called the primary key column, not excluding the possibility of having other indexes added to it. 

So, at the end of the chapter it will be possible to manipulate data in a table format using the **Table** interface. Getting data, updating and deleting rows, doing range queries for indexed columns; all these functionalities are going to be called *base table functions*, whose level is the lowest in the layer of all possible functions.

Whenever a query is written by an user, after parsed, a stack of database operations is going to be created, in which the lowest level possible is to get and manipulate data in the files.
