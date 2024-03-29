package cache

import (
	"sync"
)

type Node struct {
	prev  *Node
	next  *Node
	key   string
	value []byte
}

// An LRU is a thread-sife, fixed-size in-memory cache with a least-recently-used eviction policy
type LRU struct {
	m        sync.RWMutex
	entries  map[string]*Node
	head     *Node
	tail     *Node
	stats    *Stats
	capacity int
	used     int
}

// NewLRU returns a pointer to a new LRU with a capacity to store limit bytes
func NewLru(limit int) *LRU {
	lru := &LRU{
		capacity: limit,
		entries:  make(map[string]*Node),
		stats:    new(Stats),
		m:        sync.RWMutex{},
	}
	return lru
}

// MaxStorage returns the maximum number of bytes this LRU can store
func (lru *LRU) MaxStorage() int {
	lru.m.RLock()
	defer lru.m.RUnlock()

	return lru.capacity
}

// RemainingStorage returns the number of unused bytes available in this LRU
func (lru *LRU) RemainingStorage() int {
	lru.m.RLock()
	defer lru.m.RUnlock()

	return lru.capacity - lru.used
}

// Get returns the value associated with the given key, if it exists.
// This operation counts as a "use" for that key-value pair
// ok is true if a value was found and false otherwise.
func (lru *LRU) Get(key string) (value []byte, ok bool) {
	lru.m.Lock()
	defer lru.m.Unlock()

	item, ok := lru.entries[key]
	if ok {
		lru.stats.Hits++

		// move node to head
		if item == lru.head {
			return item.value, true
		}

		if item == lru.tail {
			lru.tail = item.next
		}

		prev := item.prev
		if item.prev != nil {
			item.prev.next = item.next
		}
		if item.next != nil {
			item.next.prev = prev
		}
		if lru.head != nil {
			lru.head.next = item
		}
		item.prev = lru.head
		lru.head = item

		return item.value, true
	}
	lru.stats.Misses++
	return nil, false
}

// Remove removes and returns the value associated with the given key, if it exists.
// ok is true if a value was found and false otherwise
func (lru *LRU) Remove(key string) (value []byte, ok bool) {
	lru.m.Lock()
	defer lru.m.Unlock()

	item, ok := lru.entries[key]
	if ok {
		memory := len(key) + len(item.value)
		if item == lru.head {
			lru.head = item.prev
		}
		if item == lru.tail {
			lru.tail = item.next
		}
		prev := item.prev
		if item.prev != nil {
			item.prev.next = item.next
		}
		if item.next != nil {
			item.next.prev = prev
		}
		value := item.value
		delete(lru.entries, key)
		lru.used -= memory
		return value, true
	}
	return nil, false
}

// Set associates the given value with the given key, possibly evicting values
// to make room. Returns true if the binding was added successfully, else false.
func (lru *LRU) Set(key string, value []byte) bool {
	lru.m.Lock()
	defer lru.m.Unlock()

	memory := len(key) + len(value)
	if memory > lru.capacity {
		return false
	}
	item, ok := lru.entries[key]
	if ok {
		oldMemory := len(key) + len(item.value)
		if memory > lru.capacity-lru.used+oldMemory {
			return false
		}

		// Remove the old key-value pair from the cache
		if item == lru.head {
			lru.head = item.prev
		}
		if item == lru.tail {
			lru.tail = item.next
		}
		prev := item.prev
		if item.prev != nil {
			item.prev.next = item.next
		}
		if item.next != nil {
			item.next.prev = prev
		}
		delete(lru.entries, key)
		lru.used -= oldMemory
	}

	// Evicting until enough memory is available
	for memory > lru.capacity-lru.used {
		tail := lru.tail
		tailMemory := len(tail.key) + len(tail.value)
		if tail.next != nil {
			tail.next.prev = nil
		}
		lru.tail = tail.next
		delete(lru.entries, tail.key)
		lru.used -= tailMemory
	}

	// Adding new key-value pair
	node := new(Node)
	node.prev = lru.head
	node.key = key
	node.value = value
	lru.entries[key] = node
	if lru.head != nil {
		lru.head.next = node
	}
	lru.head = node

	if lru.tail == nil {
		lru.tail = node
	}
	lru.used += memory

	return true
}

// Len returns the number of bindings in the LRU.
func (lru *LRU) Len() int {
	lru.m.RLock()
	defer lru.m.RUnlock()

	return len(lru.entries)
}

// Stats returns statistics about how many search hits and misses have occurred.
func (lru *LRU) Stats() *Stats {
	lru.m.RLock()
	defer lru.m.RUnlock()

	return lru.stats
}
