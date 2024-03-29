package imdb

import (
	"sync"
)

type Database struct {
	lck     sync.RWMutex
	recs    *set
	indexes []index
}

// NewDatabase returns new in-memory database.
func NewDatabase() *Database {
	return &Database{recs: NewSet()}
}

// Index creates a non-unique index.
func (db *Database) Index(keys ...string) {
	idx := normalIndex{NewSet(), make(map[string]*set)}
	for i := range keys {
		idx.keys.Add(keys[i])
	}
	db.indexes = append(db.indexes, &idx)
}

// UniqueIndex creates a unique index.
func (db *Database) UniqueIndex(keys ...string) {
	idx := uniqueIndex{NewSet(), make(map[string]interface{})}
	for i := range keys {
		idx.keys.Add(keys[i])
	}
	db.indexes = append(db.indexes, &idx)
}

// Has checks if database has the reccord.
func (db *Database) Has(rec interface{}) bool {
	db.lck.RLock()
	defer db.lck.RUnlock()
	return db.recs.Has(rec)
}

// Add adds a record to database.
func (db *Database) Add(rec interface{}) error {
	db.lck.Lock()
	defer db.lck.Unlock()
	if !db.recs.Has(rec) {
		for i := range db.indexes {
			if err := db.indexes[i].add(rec); err != nil {
				// Rollback
				for j := i; j >= 0; j-- {
					db.indexes[j].remove(rec)
				}
				return err
			}
		}
		db.recs.Add(rec)
	}
	return nil
}

// Remove removes a record from database.
func (db *Database) Remove(rec interface{}) error {
	db.lck.Lock()
	defer db.lck.Unlock()
	if db.recs.Has(rec) {
		for i := range db.indexes {
			if err := db.indexes[i].remove(rec); err != nil {
				// Rollback
				for j := i; j >= 0; j-- {
					db.indexes[j].add(rec)
				}
				return err
			}
		}
		db.recs.Remove(rec)
	}
	return nil
}

// Query creates a new query.
func (db *Database) Query() *query {
	return &query{db: db, filters: make(map[string]value)}
}
