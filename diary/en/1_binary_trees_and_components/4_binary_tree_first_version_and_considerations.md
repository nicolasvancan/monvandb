# Next steps

Keeping in mind that the basic CRUD of my binary tree has been built (Quality must be improved), I want to advance my database development and move to the next level. Before presenting the next chapter of this adventure, I want to make some considerations and highlight some points that should be reworked in the future:

## Knots and Leaves

The binary tree base worked very well, although I think many lines of code could be removed easily with a good refactoring. Another task that I consider really necessary is to improve the time spent deleting items in sheets and nodes. When I designed the fields belonging to the TreeNode structure, I really thought that the parent address would be used frequently, and as I developed the CRUD functions, it became clear to me that that field would be useless in the way I was implementing the solution. I think I'll remove it in the future, maybe add some other field about whether or not the page is part of the current tree.

Along with these mentioned modifications, I would like to rewrite and rename several functions, making them easier to understand.

## BTree page recordings and updates

This is perhaps one of the most sensible topics in the project, file management. I mean, what is the best way to update a file or expand it. I opted for a naive way of updating the file, I just wanted to make it work, without thinking about any possible multi thread system. I update pages and write them all at once, that is, whenever I update a single page value, such as number of items, I write the total value of a page's size in bytes.

This is not atomic at all, something could go wrong and my page could become corrupted, which is not desirable. Another possible solution is to work directly with the memory map, where I don't have the callback defined, but I update the mapped bytes. The second possible solution would require some CRUD refactoring.

Another variation is to never update a Node, that is, whenever there is an update to the tree, another branch is created, copying the branch to be modified, and the only update is to set the root page to the new generated branch. This approach would easily solve the problem of unexpected shutdowns, but it would leave us with several unused pages, making our bTree file extremely large.

We can't always win, and this also applies when it comes to programming. If you reduce the complexity of one task, there may be another that will also cost you more time and complexity.

When you get to the performance part, you'll want to also evaluate the differences between the mmap and os.FileRead system call for getting and storing information. From now on, it doesn't matter much which one will be used to develop other parts of the project.

## The story

The story could also be rewritten someday, as I explained: my free time was getting scarce, and I wanted to move forward with the project and think as my heart and head told me, even to write. Of course, I didn't spend that much time correcting my English mistakes or even reviewing some sentences and how they were written, I just did it as a diary. Maybe one day I'll do something professional, also based on this project.

Which I will try to do next time. I will document more about the tests and experiences I had during the process, not just the final form of my ideas. For each implementation and test, I will write about it soon after, I think it will give more of an experience rather than an explanation of the final solution.

# Next chapter

The next chapter will cover a deeper part of file processing, and I will mainly cover database and table definitions, such as table name, database, column, row, among others. I'm excited to see what comes of this.