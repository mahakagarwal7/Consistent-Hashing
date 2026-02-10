## Why Consistent Hashing Works

Consistent hashing maps both nodes and keys to a ring. When a node is added, 
only keys between the new node and its predecessor move. This is ~1/N keys 
instead of almost all keys.

Our output shows: Adding S6 moved only 11.78% of keys (close to expected 14.29%).

## Why Modulo Hashing Fails

Modulo hashing uses hash(key) % N. When N changes, the hash of every key 
changes, causing ~100% key movement.

## Effect of Virtual Nodes

Without vnodes: StdDev = 129820 (uneven)

With 10 vnodes: StdDev = 35008 (73% better)

With 500 vnodes: StdDev = 5878 (95% better)

More virtual nodes = more even distribution.

## Complexity

- Lookup: O(log N) using binary search on sorted slice
- Add/Remove: O(K × N) for K virtual nodes
- Space: O(N) for N virtual nodes

