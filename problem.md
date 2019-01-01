I'm working on a new project in golang. A key part of it is serializing a tree datastructure as leafs are moved around, removed, and added.

My tree in normal operation looks something like this:
Root
- MuxA
  - Leaf
  - MuxB
    - Leaf
    - Leaf
  - MuxB
    - Leaf
    - Mux A
      - Leaf
      - Leaf
      - Leaf

Since each MuxA/MuxB can be storing any of MuxA/MuxB/Leaf, my assumption is that MuxA/MuxB should each be a struct storing an interface (and in runtime that interface would store one of MuxA/MuxB/Leaf)

This works perfectly (so far), since I can serialize the tree by calling on the root's serialize method and have it recursively call serialization interfaces trickling down the tree.

When I am modifying a tree, however, interfaces become a problem. A common operation is moving xxxxxxxxxxx