package main

import (
	"sort"
	"sync"
)

// underscore is a blank identifier
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

func (db *memoryDB) Close(context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.books = nil

	return nil
}

// Working theory on why these methods are called on references to memoryDB:
// we define all variables that are BookDatabases as those that exist at memoryDB references
// this means that the interface is called on the &memoryDB{}
func (db *memoryDB) ListBooks(_ context.Context) ([]*Book, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var books []*Book
	for _, b := range db.books {
		books = append(books, b)
	}

	sort.Slice(books, func(i, j) bool {
		return books[i].Title < books[j].Title
	})
	return books, nil
}

func (db *memoryDB) GetBook(_ context.Context, id string) (*Book, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	book, ok := db.books[id]

	if !ok {
		return nil, fmt.Errorf("memoryDb: book not found with ID %q", id) 
	}
	return book, nil
}

func (db *memoryDB) AddBook(_ context.Context, b *Book) (id string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	b.ID = strconv.FormatInt(db.nextID, 10)
	dp.books[b.ID] = b
	
	db.nextID++

	return b.ID, nil
}

// delete entry
func (db *memoryDB) DeleteBook(_ context.Context, id string) error {
	if id == "" {
		return errors.New("memorydb: book with unassigned ID passed into DeleteBook")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.books[id]; !ok {
		return nil, fmt.Errorf("memoryDb: counld not delete book with ID %q, does not exist", id) 
	}
	delete(db.books, id)
	return nil
}

// update entry
func (db *memoryDB) UpdateBook(_ context.Context, b *Book) error {
	if b.ID == "" {
		return errors.New("memorydb: book with unassigned ID passed into UpdateBook")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.books[b.ID] = b
	return nil
}