package imdb

import (
	"fmt"
	"reflect"
)

type normalIndex struct {
	keys *set
	recs map[string]*set
}

func (idx *normalIndex) add(rec interface{}) error {
	key, err := idx.buildKey(rec)
	if err != nil {
		return err
	}
	recs, ok := idx.recs[key]
	if !ok {
		recs = NewSet()
		idx.recs[key] = recs
	}
	recs.Add(rec)
	return nil
}

func (idx *normalIndex) remove(rec interface{}) error {
	key, err := idx.buildKey(rec)
	if err != nil {
		return err
	}
	if recs, ok := idx.recs[key]; ok {
		recs.Remove(rec)
	}
	return nil
}

func (idx *normalIndex) match(q query) int {
	for key := range idx.keys.elements {
		if _, ok := q.filters[key.(string)]; !ok {
			return noScore
		}
	}
	return idx.keys.Len()
}

func (idx *normalIndex) build(q query) (map[string]value, map[string]value) {
	if len(q.filters) == idx.keys.Len() {
		return q.filters, nil
	}
	indexed := make(map[string]value, idx.keys.Len())
	nonIndexed := make(map[string]value, len(q.filters)-idx.keys.Len())
	for key, val := range q.filters {
		if idx.keys.Has(key) {
			indexed[key] = val
		} else {
			nonIndexed[key] = val
		}
	}
	return indexed, nonIndexed
}

func (idx *normalIndex) query(q query) []interface{} {
	key := buildKey(q.indexed)
	recs, ok := idx.recs[key]
	if !ok {
		return nil
	}
	if len(q.nonIndexed) == 0 {
		return recs.All()
	}
	var results []interface{}
	for rec := range recs.elements {
		if match(rec, q.nonIndexed) {
			results = append(results, rec)
		}
	}
	return results
}

func (idx *normalIndex) buildKey(rec interface{}) (string, error) {
	v := reflect.ValueOf(rec)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("invalid type of record, got %s expected struct", v.Kind())
	}
	t := v.Type()
	indexed := make(map[string]value)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if idx.keys.Has(sf.Name) {
			indexed[sf.Name] = &boundValue{v.Field(i).Interface()}
		}
	}
	key := buildKey(indexed)
	return key, nil
}
