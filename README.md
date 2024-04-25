# MonvanDB

This project originated from tremendous boredom I recently experienced. It's really challenging for me to stay motivated when learning new technologies or tools if I don't pick a challenging project to implement what I want to learn. In other words, creating REST APIs in different languages is no longer my chosen learning method. Nor is creating test templates, or just some POCs (unless I'm really interested in seeing how something works).

After working with ML projects, Python, and a lot of data for a while, I strongly felt that I had learn a new programming language (and probably a spoken language as well). I saw some interesting projects built with Golang, Rust, and many other languages. However, this time, Golang won the race.

I've read a lot of content about the language and followed some tutorials on how to work with it (while conducting my own personal tests). I even read the official book, but I lost interest halfway through.

The advantage of working with so many different technologies is that, at some point, one of them will capture your interest. I was involved in a project where I had to extract data from various types of sources, such as Oracle tables, PDF files, and Data Lakes tables, among others.
I then processed this data and consolidated it into a single output format, before storing it in a structure of six different tables. The problem was that the client was uncertain whether the tables would remain in an Oracle Database, in the Data Warehouse, or be moved elsewhere. This indecision slowed down my development process, which was undesirable not only for me but also for the client.

Despite that, I decided not to worry about where the data would be stored. Instead, I would build an interface, the methods of which could be implemented by any type of database that might be used in the future. This was a great idea, but I had to develop it locally due to some environmental constraints (Client VM). I couldn't install any database server or even use Docker. As my wife says "The only fate that we cannot change is death", and it's really true. The main problem was to test the solution locally while I developed it, I'd need to build my own database. Why not? In my mind I honestly tought that It would not be so dificult, and it turned out to be a complex task, even working with well known frameworks.

I created and implemented the interface with a small database system based on pandas and **parquet** files. I was amazed by how well that worked, but far more amazed on how I had built a small but functional database. Far from perfection, that would be a problem with large datasets, handling concurrence, and many different problems, but this tiny action started a spark of curiosity inside of me about how databases work, bringing me the will and courage to start my own database project.

Combining both desires, to learn Golang and to build the Database, this project was initiated. Why would I keep building something that is not an inovation nor a game changer technollogy that could revolutionize IT? Acctually I don't have the answer. To be honest, I do think that a part of this is to fulfill my curiosity to know how a robust Database Server work, and also to tell my mom that I've built one. (She would be proud even though she didn't understand a word of it)

Without any time and quality expectation, I'll finish what I've started and this project will be functional.

I hope this project makes someone laugh, and hope this project brings the joy and will of working with programming to many different people.

I am not a writer, nor a English native speaker, which means that my english will be something curious to those that are more familiarized with the language. Although I wanted to use chat gpt to correct my comic english, I will not do that! I think that AI removes personallity of things and padronize everything to something near perfection, which is not my case. 

> **I HOPE YOU ENJOY IT!!!**

## What is this project?

Talking serious now, this project is simply a Key Value Database Server. The main idea is to run it as a server that can respond to simultaneous requests. Can be managed, and can guarantee that the data will be kept safe.

## Why have I decided to build my own Database from scratch?

I'm still thinking why, probably the best answer would be the following: "Not enough problems at work"

## What do I excpect with this work?

I expect to tell the history and the project evolution as a Story, every obstacle, every bug, every situation that I face, will be documented and also the solutions that I had. I think this kind of diary is a good idea, and I have never seen anything like it.

## What are my goals developing this software?

- Learn Golang properly (I'll be able to create some REST API endpoints -- Just kidding, I want to work with multiple threads/go routines and processes)
- Learn and share many knowledges acquired during the development of this project
- Create a project story.

## What are the requirements to consider the project completed?

- It must be a server
- It must be secure
- It must be manageble
- It must be fault tolerant (This will be hard to do)
- It must handle multiple CRUD requests simultaneoslly
- It must read queries
- It must have cache
- It must have backup
- It must have a API in one or more languages, such as Python and JS, for instance

# Where to follow?

The stories are saved in the diary folder at the root **/diary** as markdown files. They are separated by languages, for now, just portuguese and english are available, because I am not using AI to help me write my texts and I'll not do it, I want the texts to be original. Inside the folders, the chapter are numerated, followed by a small title about it. Entering the chapter folder, different markdown files can be found, also following a numerical order. Than is just click and read.