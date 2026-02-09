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

2) Main Functions:-
     
     a)basic := NewConsistentHasher(ringSize) //without virtual nodes
     
     b)vnode := NewConsistentHasherV(ringSize) // With 10 virtual nodes per server
     
     c)custom := NewConsistentHasherVWithVNodes(size, 100) // With 100 virtual nodes

4) Adding a server:-
     ch.AddNode("S1")
5) Finding which server handles a key:-
     server := ch.FindNodeFor("User123")
6) Removing a server:-
     ch.RemoveNode("S1")     



