package main

import (
	"sync"
)

// underscore is a blank identifier,
var _ BookDatabase = &memoryDB{}

// simple in-memory persistence layer in memory for books
// mutex allows for locking and unlocking of a memory structure
// we save nextID in the memoryDB to save the next spot that data can be saved
// we use the map to persist data that we have in memory (of books)
// (aside: the sync package in Go has a Map type that provides automcatic locking
// and unlocking, but we chose to use a Mutex for manual locking and unlocking for 
// better control over the data structure)
type memoryDB struct {
	mu     sync.Mutex
	nextID int64
	books  map[string]*Book
}

func newMemoryDB() *memoryDB{
	return &memoryDB{
		books: make(map[string]*Book),
		nextID: 1
	}
}

