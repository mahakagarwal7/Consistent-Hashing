# Consistent-Hashing Implementation

# What This Project Does?
This project implements consistent hashing - a way to distribute data across multiple servers so that when you add or remove servers, only a small amount of data needs to move.

# Why Not Regular Hashing?
Regular hashing uses hash(key) % N. When N changes (server added/removed), almost all keys move. This is bad for caches and databases.

Consistent hashing fixes this: when you add/remove a server, only ~1/N keys move.


# What the Code Contains?
1)  Two Types of Hash Rings

     a)ConsistentHasher :- Basic version-one node = one position on ring
     
     b)ConsistentHasherV :- With virtual nodes-one node = multiple positions

3) Main Functions:-
     
     a)basic := NewConsistentHasher(ringSize) //without virtual nodes
     
     b)vnode := NewConsistentHasherV(ringSize) // With 10 virtual nodes per server
     
     c)custom := NewConsistentHasherVWithVNodes(size, 100) // With 100 virtual nodes

4) Adding a server:-
     ch.AddNode("S1")
5) Finding which server handles a key:-
     server := ch.FindNodeFor("User123")
6) Removing a server:-
     ch.RemoveNode("S1")
   
# Key Features

Virtual Nodes: Each physical server gets multiple positions on the ring (default 10). This makes data distribution more even.

Rebalancing Metrics: When you add or remove a server, the code tells you exactly how many keys moved.  
metrics, _ := ch.AddNodeWithMetrics("S6", keys)

Load Distribution Analysis: Shows how evenly keys are spread across servers using standard deviation.
dist := analyzeDistribution(ch, keys, servers)

# Time Complexity

1)Find Node(Lookup) :- O(log N) - binary search
2)Add Node:- O(K × N) - K virtual nodes, insert into sorted slice
3)Remove Node:- O(K × N) - K virtual nodes, delete from slice

Insert and delete is O(N) which is acceptable as ring changes are rare.

# Design Decisions
Q)Why slices instead of trees? 

  Slices with binary search give O(log N) lookup as required. Insert/delete is O(N) but acceptable since ring changes are rare.

Q)Why MD5? 

  Fast, good distribution, produces 128-bit hash giving large ring space (0 to 2^32-1).

Q)Why 10 default virtual nodes? 

 Good balance between distribution quality and memory usage.




   



