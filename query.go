package imdb

import (
	"fmt"
	"reflect"
	"strings"
)

type query struct {
	db         *database
	filters    map[string]value
	bestIndex  index
	indexed    map[string]value
	nonIndexed map[string]value
}

// Filter adds a filter to the query.
func (q *query) Filter(key string, val interface{}) *query {
	q.filters[key] = &boundValue{val}
	return q
}

// PrepareFilter adds a filer to the query.
// valPtr is pointer to the query value which can be updated after the query is built.
func (q *query) PrepareFilter(key string, valPtr interface{}) *query {
	if t := reflect.TypeOf(valPtr); t.Kind() != reflect.Ptr {
		panic(fmt.Errorf("invalid type for 2nd argument, got %s expected ptr", t))
	}
	q.filters[key] = &unboundValue{valPtr}
	return q
}

// Build builds the query.
func (q *query) Build() *query {
	bestScore := noScore
	for i := range q.db.indexes {
		if score := q.db.indexes[i].match(*q); score > bestScore {
			q.bestIndex = q.db.indexes[i]
			if score == maxScore {
				break
			}
			bestScore = score
		}
	}
	if q.bestIndex != nil {
		q.indexed, q.nonIndexed = q.bestIndex.build(*q)
	}
	return q
}

// Run returns all records that match the query.
func (q *query) Run() []interface{} {
	q.db.lck.RLock()
	defer q.db.lck.RUnlock()
	if q.bestIndex != nil {
		return q.bestIndex.query(*q)
	}
	var recs []interface{}
	for rec := range q.db.recs.elements {
		if match(rec, q.filters) {
			recs = append(recs, rec)
		}
	}
	return recs
}

// String returns the string representation of the query.
func (q *query) String() string {
	arr := make([]string, 0, len(q.filters))
	for key, val := range q.filters {
		arr = append(arr, fmt.Sprintf("%s = %#v", key, val.value()))
	}
	return fmt.Sprintf("(%s)", strings.Join(arr, " & "))
}
