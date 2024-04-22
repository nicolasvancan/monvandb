# Next Steps

Having in mind that the basic CRUD for my binary tree was built (The quality must be improved), I want to advance on the development of my database and move to the next level. Before I event introduce the next chapter of this adventure, I want to make some considerations and to highlight some points that must be reworked in the future:

## Nodes and Leaves

The base of binary tree worked really good, although I feel that many line of code could be removed easly with a good refactoring. Another task that I feel that is indeed needed is to improve the time spent for deleting items in leaves and nodes. When I designed the fields belonging to the TreeNode Structure, I realy thought that the parent address would be frequently used, and as I developed the CRUD functions, it became clear to me that that field would be useless in the way I was implementing the solution. I think I'll remove it in the future, maybe I add some other field regarding whether or not the page is part of the current tree.

Along with these mentioned modifications, I'd like to rewrite and rename a lot of functions, making it easier to understand.

## BTree pages writes and updates

This might be one of the most sensible subjects of the project, the file management. I mean, what is the best way of updating a file, or expanding it. I chose a naive way of updating the file, I just wanted to make it work, without thinking in any possible multi thread system yet. I do update pages, and I write them all at once, that means, whenever I update a single page value, such as number of items, I write the total amount of a page size in bytes.

This is not atomic at all, something could go wrong and my page becomes corrupted, which is not something desireble. Another possible solution is to work directly with the memory map, where I don't have the callback set, rather I update the mapped bytes. The second possible solution would require some refactoring to the CRUD.

Another variation is to never update a Node, meaning that whenever there is an update on tree, another branch is created, copying the branch to be modified, and the only update is to set the root page to the new generated branch. This approach would solve easly the problem of unexpected shutdowns, but it would led us to various unused pages, making our bTree file extremly large. 

We cannot always win, and that also applies when dealing with programming. If you reduce complexity of one task, there might be another one that will cost you more time and complexity as well.

When I get to the part of performance, it will nice to evaluate also the differences between system call mmap and the os.FileRead, for getting and storing information. From now, it doesn't metter much which one is going to be used for developing other project parts.

## The story

The story may also be rewritten sometime, as I explained: my free time is becoming scarse, and I wanted to move forward with the project and do thinks as my heart and head told me, even for writting. It's clear that I didn't spend so much time os correcting my english errors or even reviewing some sentences and how they were written, I just did it as a diary. Maybe I'll do something professional one day, also based os this project.

What I'll try to do next time. I'll document more about tests and experiences I had during the process, and not only the final form of my ideas. For every implementation and test, I'll write shortly afterwards about that, I think that will give more a touch of experiencing than of explaining the final solution.

# Next Chapter

The next chapter will cover a deeper part of file handling, and I'll mainly cover the database and tables definitions, such as table name, database, column, row, among other. I'am excited to see what will emerge out of it.