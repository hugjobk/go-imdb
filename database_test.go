package imdb_test

import (
	"fmt"
	"imdb"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const DatabaseSize = 500000

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

	// Node: unique index is normally faster non-unique index.
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

	// Node: you can query by any field even if it is not indexed.
	// However it can be slow because it has to scan the database to find matched records.
	// You can use PrepareFilter to reuse a query multiple times for different query parameters (see TestDatabase2).

	// Build a query where `Id` = 3.
	q1 := db.Query().Filter("Id", 3).Build()
	// Build a query where `Email` = "email2".
	q2 := db.Query().Filter("Email", "email2").Build()
	// Build a query where `Age` = 21.
	q3 := db.Query().Filter("Age", 21).Build()
	// Build a query where `Age` = 21 AND `Name` = "name4".
	q4 := db.Query().Filter("Age", 21).Filter("Name", "name4").Build()

	// Print the queries and their results.
	t.Log(q1)
	t.Log(printResults(q1.Run()))

	t.Log(q2)
	t.Log(printResults(q2.Run()))

	t.Log(q3)
	t.Log(printResults(q3.Run()))

	t.Log(q4)
	t.Log(printResults(q4.Run()))
}

// Benmark database performance with multiple concurrent inserts and queries.
func TestDatabase2(t *testing.T) {
	// Create a database and indexes to speed up query speed.
	db := imdb.NewDatabase()
	db.UniqueIndex("Id")
	db.UniqueIndex("Email")
	db.Index("Age", "Name")

	numCpu := runtime.NumCPU()
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup
	wg1.Add(numCpu)
	wg2.Add(numCpu)

	start1 := time.Now()
	// Start multiple goroutines to insert into the database.
	for i := 0; i < numCpu; i++ {
		j := i
		go func() {
			for ; j < DatabaseSize; j += numCpu {
				id := j
				email := fmt.Sprintf("email%d", j)
				name := fmt.Sprintf("name%d", j%1000)
				age := j % 100
				db.Add(record{id, email, name, age})
			}
			wg1.Done()
		}()
	}
	wg1.Wait()
	end1 := time.Now()
	t.Logf("Added %d records in %v\n", DatabaseSize, end1.Sub(start1))

	start2 := time.Now()
	// Start multiple goroutines to query the database.
	for i := 0; i < numCpu; i++ {
		j := i
		go func() {
			var id int
			var email string
			var name string
			var age int
			// Build queries that can be reused multiple times.
			q1 := db.Query().PrepareFilter("Id", &id).Build()
			q2 := db.Query().PrepareFilter("Email", &email).Build()
			q3 := db.Query().PrepareFilter("Age", &age).PrepareFilter("Name", &name).Build()
			for ; j < DatabaseSize; j += numCpu {
				// Update query parameters after the query is built.
				id = j
				email = fmt.Sprintf("email%d", j)
				name = fmt.Sprintf("name%d", j%1000)
				age = j % 100
				// Execute the queries.
				q1.Run()
				q2.Run()
				q3.Run()
			}
			wg2.Done()
		}()
	}
	wg2.Wait()
	end2 := time.Now()
	t.Logf("Executed %d queries in %v", DatabaseSize*3, end2.Sub(start2))
}
