### Go project structure:
- /cmd -> things we run
- /internal -> things we build
## LRU Cache Design
###### Image
The cache provides O(1) access since it uses a 
map from key K to a DLL node (Entry)
which is a (Key,Value) pair and pointers to previous and next element.
The map allows fast lookup to the nodes, while DLL maintains the order for least and
most recently used elements.
The cache keeps a pointer to the head, which represents the most 
recently used element, and a pointer to the tail, 
which represents the least recently used element.