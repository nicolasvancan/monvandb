# The beginning of the Journey

Sometimes, while working on projects, I feel like I'm playing an RPG, where I have some quests to complete for achieving new quests, gain experience, and gather new weapons, all of that to enable me to face new and more challenging situations.

I think I really like this feeling of RPG in development, and that is something that motivates me to keep pushing and learning new subjects. I must confess, I felt upset and unmotivated recently; there were some things in my life that were worrying me, which is normal. Concerns are part of our life; despite that, you must keep yourself positive and work to push away those concerns and bring joy to your life.

Among the problems I have to face at the end of 2023, there is my boredom in my free time. I'm a very active guy; I go to the gym, I watch anime, but I also like to involve myself in programming or engineering projects, and I was feeling too lazy to do any of those things. I had to change it.

You know, I need money for me and my family. My wife is pregnant, and I have a lot of bills to pay; we are renovating our apartment to welcome the baby. I am investing a lot of my free time in building a SaaS platform (which is still in development) for accounting services. The project itself is interesting from the perspective of software and cloud architecture and also in terms of money, but it is leaving me tired, and in the sense of reasoning, the challenge was not satisfying.

Some people just say to me to focus on what gives you money directly, but doing that, with all the situations in my life right now, I started becoming really stressed and upset. I have to recover my energy.

Then the idea of a Database came up. Why not spend time building something great and also learn something new? I combined both the Golang language and the database project to be my free time partner for a while. And as I expected, I was diving deep into the project right from the beginning.

# Hands On - First day

I remember that at the university, I had some tests and small projects regarding Data Structures, such as Linked lists, Binary Trees, Graphs, among others. Since I started my career as a developer, I haven't had so many opportunities to implement them in real life (real projects), although they are naturally part of our life (like it or not). Perhaps that is the reason why the big tech companies care so much about this sort of knowledge in their interviews; who knows, you might implement them one day?

I knew that the database files could be similar to some data structures, and reading more about databases, I realized it was true. The implementation of data structures in files is possible and is used a lot in the software industry. That is the case of B-tree.

### The BTree and what that is

The B-tree, or Binary Tree, is a data structure used to store data, but the key point is how it stores it. It works as a tree, having nodes and leaves. The tree also has some limitations, such as the maximum number of members within a node or a leaf, the values must be sorted, and the tree must stay balanced to prevent some possible slow searches.

If you want to read a good article on B-trees and the types of Trees that exist, just check this [link](https://builtin.com/data-science/b-tree-index).

Normally, all the examples on the internet are done using memory pointers, which is not our use case. The tree we are about to build is based on disk information. The idea is quite similar, but instead of addressing memory pointers to nodes, we are going to address one file page.

### How would that work?

The operating system provides an API to access files through its File System. In other words, we can access data from disk, read, write, modify, save files, and so on.

The data of a database is written in one or more files, but the data doesn't stay there randomly; it has an order, it is correctly divided to ease the access to it.

Have you ever read a book? Normally books have indices that demonstrate where to find chapters or some specific part of it. You read what is the page that you want to reach from the index, let's say Chapter 2 of a book, and go directly there, instead of keep turning pages until you reach that page. It is much faster than the normal turning page process.

Imagine that now for a database that has millions, sometimes billions, or more lines in a table. How to access it rapidly? How does a computer access just a specific page of a file, without bringing all the file bytes into memory?

That is the magic of the database, accessing the right page to get the right data rapidly.

### First Objective

Knowing that I must build a B-tree based on disk and I have to separate the file into pages, I have to make some tests. And those tests require some understanding of the language that I want to use (Golang) to build the database. So I decided to divide my first tasks into three different:

- Learn Golang API for managing files
- Research and understand more about file page size
- How to load just a specific part of a file, let's say, from byte 200 to 400?

#### Day One - Golang API for Files

Let's code now. I am new to Go programming. To be honest, I've downloaded the official Go book and read some basic chapters. I prefer to try myself some implementations and see what comes out.

**PS: Some code generated here will not be present in commits, because they were used just to prove concepts and test functions**

For testing purposes, I started a new file in Golang **main.go** with the package name **main** and created an empty function **main**.

```go
package main

func main() {
    // I'll implement the code here
}
```

So far so good, but I don't know how to import external packages, or even built-in packages. Using the most famous website for that **StackOverflow** I read that you must use **import** statement with the library name that you want to use. Let's import the library os, I think that is the right one to import, and use the method for creating a file, normally is the method Open present in many other languages as well.

I was right, there is a public function in the library **os** from Golang that enables us to create a file. So I wrote the following code:


```go
package main

import "os"

func main() {
    // I'll implement the code here
    fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/text.txt"
    // Filename, modes for opening file Create (if doesn't exist and Read Write)
    // And also permissions for file
    fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)

    if err != nil {
        panic("Could not create file")
    }

    fp.Write([]byte("Hello From Nicolas"))

    defer fp.Close()
}
```

We can see that the file was created correctly and also that the string wrote into file is correct

![alt text](../../assets/txt_file_created.png)

![alt text](../../assets/text_txt_text.png)

Ok, I think that was really easy, since the API is similar to all other languages, this is not something new to me. But let's say I have a file of a length of 4096 bytes, and I want to read the first 100 bytes without loading the rest of the file into memory, how can I do that?

One hypothesis is to use the **os** library as well. Let's see if it works. Using the same function **OpenFile**, we can access the data saved in the file, but I want to access the first 100 bytes.

First of all, I'll create an array of bytes of size 4096, concatenating the same phrase *Hello From Nicolas* followed by a lot of hashtags *#*, then I'll write it to the file, using the same script.

```go
package main

import "os"

func generateByteArray() []byte {
	header := []byte("Hello From Nicolas")
	// Create new empty 4096 byte array
	mArray := make([]byte, 4096)
	// Copy header to mArray
	copy(mArray[:len(header)], header)

	// fill out the rest with hashtag
	for i := len(header); i < len(mArray); i++ {
		mArray[i] = byte('#')
	}

	return mArray
}

func main() {
    // I'll implement the code here
    fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/text.txt"
    // Filename, modes for opening file Create (if doesn't exist and Read Write)
    // And also permissions for file
    fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)

    if err != nil {
        panic("Could not create file")
    }

    fp.Write(generateByteArray())

    defer fp.Close()
}
```
After the file is updated, I must write some code to read out data out of the file.

```go
func main() {
	// I'll implement the code here
	fileName := "/home/nicolas/Desktop/nicolas/projetos/monvandb/test.txt"
	fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		panic("Could not create file")
	}

	// Output with the size that we want to read
	output := make([]byte, 18)
	_, err = fp.ReadAt(output, 0)

	if err != nil {
		panic("Could not read file")
	}

	// print readData
	fmt.Printf("First 18 characteres = %s\n", output)

	defer fp.Close()
}
```

Also returning Hello From Nicolas, which was expected. I don't know if using **os.OpenFile** is the proper way of doing access to files when constant modifications keep happening. The OS requests resources to the kernel to access files and other peripherals. Why not using directly the Kernel to read file bytes? The answer to this question remains for posterity. With all of the testing made today, I have already a notion of how I should use Golang's API for files.


#### Day Two - Understanding more about File Systems

Reading books has not always been my passion. When I was younger, I didn't have the habit of reading. Out of curiosity, I gave reading a try, bought some history books, and one of them was a mystery-solving book. After finishing the first book (the mystery one), I fell in love with reading. The feeling of mystery beginning to reveal itself is exciting; you literally can't stop reading when that feeling comes up.

And the same feeling comes to me when learning new subjects for solving problems or building projects. Obviously, after I started researching about databases, files, I had that feeling of revealing the mysteries in my mind along with the excitement. I needed to know more about files and how the OS deals with them. How to minimize the chance of corrupting a file during unexpected failures.

Ideally, when modifying a file, the operations should be atomic according to some blogs and topics I've found. But how do I ensure that all operations are atomic? What does that mean? How does the file writing process work behind the scenes?

I think that to respond to this question, I shall learn how the machine deals with this sort of task.

**Atomicity**

Atomic operations are defined as operations that cannot be interrupted by any other process or task. To achieve that in file system operations, such as Write or Read, may be sometimes hard.

One good way of modifying files with one single atomic operation is to rename the file. Yes, the idea is simply to always have a copy of the original file (Database), and whenever we want to commit changes to the original database, the file name is changed. This action is always atomic, meaning that if any fault occurs during the process, either the operation occurs or not, the file would not be corrupted.

**Page Size**

Since reading and writing data to disk means more disk operations, we want to ensure that those operations are done in an optimized way. But how? Knowing the size of the OS Page Size helps us determine what will be the size of our pages used in our data structure.

Ensuring that the pages have the size of the OS page size, we not only have efficient memory usage by the kernel and system, but also many more advantages, both for concurrency control, disk operations, and recovery mechanisms.

If you have a Unix-like OS, using **getconf PAGE_SIZE** will display the page size of your computer.

#### Conclusion

I really thought that the file subjects would be sensitive, but not that there were so many possibilities and details in the subject. I think that during the project building, I'll have more and more modifications to the code, in order to fit all requirements presented at the beginning of the project.

For now, I'll focus on enabling the most basic blocks of the project. The first part is to build the B-tree and its basic operations. I am really excited to start this chapter of the project. I don't have much experience with Golang, nor do I have the fixed idea of how it will be implemented, but I'll do it anyway.

