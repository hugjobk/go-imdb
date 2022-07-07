package imdb_test

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hugjobk/imdb"
)

// Define a record struct.
type record struct {
	Id    int
	Email string
	Name  string
	Age   int
}

// Utility function to print query results.
func printResults(recs []interface{}) string {
	var arr []string
	for i := range recs {
		arr = append(arr, fmt.Sprintf("%+v", recs[i].(record)))
	}
	return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
}

// Simple test for creating, indexing, insertion and query database.
func TestDatabase(t *testing.T) {
	db := imdb.NewDatabase() // Create new database.
	db.UniqueIndex("Id")     // Create a unique index on field `Id`.
	db.UniqueIndex("Email")  // Create a unique index on field `Email`.
	db.Index("Age")          // Create a non-unique index on field `Age`.
	db.Index("Age", "Name")  // Create a non-unique index on fields `Age` and `Name`.

	// Note: unique index is normally faster than non-unique index.
	// If a query mathes both unique and non-unique index, it will choose unique index to query.

	// Add first recrod.
	if err := db.Add(record{1, "email1", "name1", 20}); err != nil {
		t.Log(err)
	}
	// Add second record.
	if err := db.Add(record{2, "email2", "name2", 20}); err != nil {
		t.Log(err)
	}
	// Add third record.
	if err := db.Add(record{3, "email3", "name3", 21}); err != nil {
		t.Log(err)
	}
	// Add forth record.
	if err := db.Add(record{4, "email4", "name4", 21}); err != nil {
		t.Log(err)
	}
	// Add fifth record.
	// This should return an error since a record with `Id` = 1 already exists (violate unique index on `Id`).
	if err := db.Add(record{1, "email5", "name5", 20}); err != nil {
		t.Log(err)
	}

	// Note: you can query by any field even if it is not indexed.
	// However it can be slow because it has to scan the database to find matched records.
	// You can use PrepareFilter to reuse a query multiple times for different query parameters (see TestDatabase2).

	// Build a query where `Id` = 3.
	q1 := db.Query().Filter("Id", 3).Build()
	// Build a query where `Email` = "email2".
	q2 := db.Query().Filter("Email", "email2").Build()
	// Build a query where `Age` = 21.
	q3 := db.Query().Filter("Age", 21).Build()
	// Build a query where `Name` = "name4" AND `Age` = 21.
	q4 := db.Query().Filter("Name", "name4").Filter("Age", 21).Build()

	// Print the queries and their results.
	t.Logf("%s -> %s", q1, printResults(q1.Run()))
	t.Logf("%s -> %s", q2, printResults(q2.Run()))
	t.Logf("%s -> %s", q3, printResults(q3.Run()))
	t.Logf("%s -> %s", q4, printResults(q4.Run()))
}

var db *imdb.Database

var DatabaseSize uint32

func init() {
	// Create a database and indexes to speed up query speed.
	db = imdb.NewDatabase()
	db.UniqueIndex("Id")
	db.UniqueIndex("Email")
	db.Index("Name", "Age")
}

func BenchmarkAdd(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			i := atomic.AddUint32(&DatabaseSize, 1)
			id := int(i)
			email := fmt.Sprintf("email%d", i)
			name := fmt.Sprintf("name%d", i%10000)
			age := int(i % 100)
			db.Add(record{id, email, name, age})
		}
	})
}

func BenchmarkQueryById(b *testing.B) {
	var c uint32
	b.RunParallel(func(p *testing.PB) {
		var id int
		q := db.Query().PrepareFilter("Id", &id).Build()
		for p.Next() {
			i := atomic.AddUint32(&c, 1)
			id = int(i % uint32(DatabaseSize))
			q.Run()
		}
	})
}

func BenchmarkQueryByEmail(b *testing.B) {
	var c uint32
	b.RunParallel(func(p *testing.PB) {
		var email string
		q := db.Query().PrepareFilter("Email", &email).Build()
		for p.Next() {
			i := atomic.AddUint32(&c, 1)
			email = fmt.Sprintf("email%d", i%DatabaseSize)
			q.Run()
		}
	})
}

func BenchmarkQueryByNameAndAge(b *testing.B) {
	var c uint32
	b.RunParallel(func(p *testing.PB) {
		var name string
		var age int
		q := db.Query().PrepareFilter("Name", &name).PrepareFilter("Age", &age).Build()
		for p.Next() {
			i := atomic.AddUint32(&c, 1)
			name = fmt.Sprintf("name%d", i%10000)
			age = int(i % 100)
			q.Run()
		}
	})
}
